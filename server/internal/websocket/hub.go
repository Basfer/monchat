package websocket

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // В продакшене нужно настроить CheckOrigin правильно
	},
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
	roomID string
}

type Hub struct {
	clients    map[string]*Client            // userID -> client (one connection per user)
	rooms      map[string]map[string]*Client // roomID -> map[userIDs]
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
	mu         sync.RWMutex
	db         *sql.DB
}

func NewHub(db *sql.DB) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
		db:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			// Если у пользователя уже есть подключение, закрываем старое
			if existing, ok := h.clients[client.userID]; ok && existing != client {
				close(existing.send)
				existing.conn.Close()
			}
			h.clients[client.userID] = client

			// Добавляем клиента в комнаты, если roomID указан
			if client.roomID != "" {
				if h.rooms[client.roomID] == nil {
					h.rooms[client.roomID] = make(map[string]*Client)
				}
				h.rooms[client.roomID][client.userID] = client
			}
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.userID]; ok {
				delete(h.clients, client.userID)
				close(client.send)
				if client.roomID != "" {
					if roomClients, ok := h.rooms[client.roomID]; ok {
						delete(roomClients, client.userID)
						if len(roomClients) == 0 {
							delete(h.rooms, client.roomID)
						}
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			var roomMsg RoomMessage
			err := json.Unmarshal(message, &roomMsg)
			if err != nil {
				log.Printf("Error unmarshaling broadcast message: %v", err)
				h.mu.RUnlock()
				continue
			}

			// Получаем имя отправителя из базы данных
			senderName := ""
			if roomMsg.SenderID != "" {
				senderName, _ = h.getUsernameByID(roomMsg.SenderID)
			}

			// Рассылаем сообщение всем подключенным клиентам
			// (клиенты не привязаны к комнатам при подключении, поэтому рассылка всем)
			for _, client := range h.clients {
				// Отправляем только если у клиента нет roomID или он совпадает с комнатой сообщения
				// Для status и deleted событий отправляем всем
				if client.roomID == "" || client.roomID == roomMsg.RoomID ||
					roomMsg.Type == "message_status" || roomMsg.Type == "message_deleted" {
					msgWithSender, _ := json.Marshal(map[string]interface{}{
						"type":       roomMsg.Type,
						"roomID":     roomMsg.RoomID,
						"userID":     roomMsg.UserID,
						"senderID":   roomMsg.SenderID,
						"senderName": senderName,
						"content":    roomMsg.Content,
						"data":       roomMsg.Data,
						"timestamp":  roomMsg.Timestamp.Format(time.RFC3339),
					})
					select {
					case client.send <- msgWithSender:
					default:
						// Если канал полный, пропускаем
					}
				}
			}

			h.mu.RUnlock()
		}
	}
}

// getUsernameByID получает username по user ID
func (h *Hub) getUsernameByID(userID string) (string, error) {
	if h.db == nil {
		return "User", sql.ErrConnDone
	}

	var username string
	err := h.db.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		return "User", err
	}
	return username, nil
}

type RoomMessage struct {
	Type      string                 `json:"type"`
	RoomID    string                 `json:"roomID"`
	UserID    string                 `json:"userID,omitempty"`
	SenderID  string                 `json:"senderID,omitempty"`
	Content   string                 `json:"content,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %s: %v", c.userID, err)
			}
			break
		}

		// Обработка входящих сообщений от клиента
		var msg RoomMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			msg.SenderID = c.userID
			msg.Timestamp = time.Now()

			// Сохраняем сообщение в базу данных
			if c.hub.db != nil && msg.RoomID != "" && msg.Content != "" {
				_, err := c.hub.db.Exec(
					"INSERT INTO messages (id, room_id, sender_id, content, type, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
					uuid.New().String(), msg.RoomID, msg.SenderID, msg.Content, "m.text", msg.Timestamp,
				)
				if err != nil {
					log.Printf("Error saving message to database: %v", err)
				}
			}

			// Пересылаем сообщение в хаб для рассылки другим участникам комнаты
			hubMsg, _ := json.Marshal(msg)
			c.hub.Broadcast <- hubMsg
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)
		if err := w.Close(); err != nil {
			return
		}
	}
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Получаем userID из query параметров
	userID := r.URL.Query().Get("userId")
	roomID := r.URL.Query().Get("roomId")

	if userID == "" {
		conn.Close()
		return
	}

	hub := GetHub()
	client := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
		roomID: roomID,
	}

	hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

var globalHub *Hub

func GetHub() *Hub {
	if globalHub == nil {
		globalHub = NewHub(nil)
	}
	return globalHub
}

func SetHub(h *Hub) {
	globalHub = h
}
