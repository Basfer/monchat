package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/matrix-messenger/server/internal/models"
	"github.com/matrix-messenger/server/internal/services"
	"github.com/matrix-messenger/server/internal/websocket"
)

type Handler struct {
	authService       *services.AuthService
	roomService       *services.RoomService
	directChatService *services.DirectChatService
	db                *sql.DB
	hub               *websocket.Hub
}

func NewHandler(db *sql.DB, hub *websocket.Hub) *Handler {
	return &Handler{
		authService:       services.NewAuthService(db),
		roomService:       services.NewRoomService(db),
		directChatService: services.NewDirectChatService(db),
		db:                db,
		hub:               hub,
	}
}

// RegisterUser обрабатывает регистрацию пользователя
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		http.Error(w, "Registration failed: "+err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// LoginUser обрабатывает вход пользователя
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// CreateRoom создает новую комнату
func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Type        models.RoomType `json:"type"`
		Members     []string        `json:"members"` // userIDs для добавления в комнату
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем userID из заголовка или контекста (в реальном проекте - из JWT)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	// Обработка direct чатов через DirectChatService
	if req.Type == models.RoomTypeDirect {
		if len(req.Members) != 1 {
			http.Error(w, "Direct chat requires exactly one member", http.StatusBadRequest)
			return
		}

		room, err := h.directChatService.GetOrCreateDirectChat(userID, req.Members[0])
		if err != nil {
			http.Error(w, "Failed to create direct chat: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(room)
		return
	}

	room, err := h.roomService.CreateRoom(req.Name, req.Description, req.Type, userID)
	if err != nil {
		http.Error(w, "Failed to create room: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Добавляем других участников, если они указаны
	for _, memberID := range req.Members {
		if memberID != userID {
			h.roomService.AddMember(room.ID, memberID, models.RoleMember)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

// GetUserRooms возвращает список комнат пользователя
func (h *Handler) GetUserRooms(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	// Опциональный параметр фильтрации по типу комнаты
	roomType := r.URL.Query().Get("type")
	var roomTypePtr *string
	if roomType != "" {
		roomTypePtr = &roomType
	}

	rooms, err := h.roomService.GetUserRooms(userID, roomTypePtr)
	if err != nil {
		http.Error(w, "Failed to get rooms: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

// GetMessages возвращает сообщения комнаты
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	limit := 50
	start := r.URL.Query().Get("start")
	if start != "" {
		limit = 50
	}

	msgs, err := h.roomService.GetMessages(roomID, limit, start)
	if err != nil {
		http.Error(w, "Failed to get messages: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

// SendMessage отправляет новое сообщение
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	var req struct {
		Content string             `json:"content"`
		Type    models.MessageType `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	// Создаем сообщение в БД
	now := time.Now()
	msg := models.Message{
		ID:        uuid.New().String(),
		RoomID:    roomID,
		SenderID:  userID,
		Content:   req.Content,
		Type:      req.Type,
		CreatedAt: now,
	}

	_, err := h.db.Exec(
		"INSERT INTO messages (id, room_id, sender_id, content, type, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		msg.ID, msg.RoomID, msg.SenderID, msg.Content, msg.Type, msg.CreatedAt,
	)
	if err != nil {
		http.Error(w, "Failed to send message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем сообщение через WebSocket всем участникам комнаты
	wsMsg := websocket.RoomMessage{
		Type:      "message",
		RoomID:    roomID,
		SenderID:  userID,
		Content:   req.Content,
		Timestamp: now,
	}
	msgBytes, _ := json.Marshal(wsMsg)
	h.hub.Broadcast <- msgBytes

	// Set initial status to 'new' for all recipients except sender
	_, err = h.db.Exec(`
		INSERT INTO message_read_status (message_id, user_id, status, updated_at)
		SELECT $1, user_id, 'new', NOW()
		FROM room_members
		WHERE room_id = $2 AND user_id != $3
		ON CONFLICT (message_id, user_id)
		DO NOTHING
	`, msg.ID, roomID, userID)
	if err != nil {
		// Log error but don't fail the request
		fmt.Println("Warning: failed to initialize message read status:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

// SearchUsers ищет пользователей по username или email
func (h *Handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		// Если нет query, возвращаем всех пользователей кроме текущего
		searchQuery = ""
	}

	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.updated_at
		FROM users u
		WHERE u.id != $1
	`
	args := []interface{}{userID}
	pos := 2

	if searchQuery != "" {
		placeholder := fmt.Sprintf("$%d", pos)
		query += ` AND (u.username ILIKE ` + placeholder + ` OR u.email ILIKE ` + placeholder + `)`
		args = append(args, "%"+searchQuery+"%")
		pos++
	}

	query += ` ORDER BY u.username ASC LIMIT 20`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to search users: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type UserSearchResult struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	results := make([]UserSearchResult, 0)
	for rows.Next() {
		var result UserSearchResult
		var createdAt, updatedAt time.Time
		err := rows.Scan(&result.ID, &result.Username, &result.Email, &createdAt, &updatedAt)
		if err != nil {
			return
		}
		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// GetOrCreateDirectChat получает существующий или создает новый личный чат
func (h *Handler) GetOrCreateDirectChat(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "target user_id is required", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли уже чат
	existingRoom, err := h.directChatService.FindExistingDirectChat(userID, req.UserID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Failed to check existing chat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	isNewChat := existingRoom == nil

	room, err := h.directChatService.GetOrCreateDirectChat(userID, req.UserID)
	if err != nil {
		http.Error(w, "Failed to get/create direct chat: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Получаем информацию о втором участнике
	otherParticipant, err := h.directChatService.GetOtherParticipantInfo(room.ID, userID)
	if err != nil {
		http.Error(w, "Failed to get other participant info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Отправляем WebSocket событие только если чат был создан новый
	if isNewChat && h.hub != nil {
		participants := room.Participants
		if len(participants) == 2 {
			h.hub.SendDirectChatCreatedEvent(room.ID, participants[0], participants[1], &websocket.UserShortInfo{
				ID:       otherParticipant.ID,
				Username: otherParticipant.Username,
			})
		}
	}

	response := map[string]interface{}{
		"room":              room,
		"other_participant": otherParticipant,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetUserDirectChats возвращает все личные чаты пользователя
func (h *Handler) GetUserDirectChats(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	rooms, err := h.roomService.GetDirectChatsForUser(userID)
	if err != nil {
		http.Error(w, "Failed to get direct chats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Для каждой комнаты добавляем информацию о втором участнике
	type DirectChatResponse struct {
		Room             *models.Room `json:"room"`
		OtherParticipant *models.User `json:"other_participant"`
	}

	responses := make([]DirectChatResponse, 0)
	for _, room := range rooms {
		otherParticipant, err := h.directChatService.GetOtherParticipantInfo(room.ID, userID)
		if err != nil {
			continue
		}

		responses = append(responses, DirectChatResponse{
			Room:             &room,
			OtherParticipant: otherParticipant,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// isValidRoomType проверяет, является ли тип комнаты допустимым
func isValidRoomType(roomType string) bool {
	validTypes := []string{"direct", "group", "channel", "public"}
	for _, t := range validTypes {
		if strings.EqualFold(roomType, t) {
			return true
		}
	}
	return false
}

// UpdateMessageStatus handles message status updates (delivered, read)
func (h *Handler) UpdateMessageStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	var req struct {
		MessageID string               `json:"message_id"`
		Status    models.MessageStatus `json:"status"`
		RoomID    string               `json:"room_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.MessageID == "" || req.RoomID == "" {
		http.Error(w, "message_id and room_id are required", http.StatusBadRequest)
		return
	}

	// Validate status
	switch req.Status {
	case models.MessageStatusDelivered, models.MessageStatusRead:
		// Valid statuses
	default:
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	err := h.roomService.UpdateMessageStatus(req.MessageID, userID, req.RoomID, req.Status)
	if err != nil {
		http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// MarkRoomMessagesAsRead marks all messages in a room as read
func (h *Handler) MarkRoomMessagesAsRead(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	err := h.roomService.MarkMessagesAsRead(roomID, userID)
	if err != nil {
		http.Error(w, "Failed to mark messages as read: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast status update via WebSocket
	now := time.Now()
	statusMsg := websocket.RoomMessage{
		Type:      "message_status",
		RoomID:    roomID,
		SenderID:  userID,
		Timestamp: now,
		Data: map[string]interface{}{
			"action": "read_all",
		},
	}
	msgBytes, _ := json.Marshal(statusMsg)
	h.hub.Broadcast <- msgBytes

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "marked_read"})
}

// DeleteMessage soft deletes a message for the sender
func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	messageID := vars["messageId"]

	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusUnauthorized)
		return
	}

	err := h.roomService.SoftDeleteMessage(messageID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Message not found or not authorized to delete", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast deletion via WebSocket
	now := time.Now()
	statusMsg := websocket.RoomMessage{
		Type:      "message_deleted",
		RoomID:    "", // Will be looked up by clients
		SenderID:  userID,
		Timestamp: now,
		Data: map[string]interface{}{
			"message_id": messageID,
		},
	}
	msgBytes, _ := json.Marshal(statusMsg)
	h.hub.Broadcast <- msgBytes

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
