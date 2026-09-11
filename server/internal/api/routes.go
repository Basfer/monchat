package api

import (
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/matrix-messenger/server/internal/websocket"
)

func RegisterRoutes(r *mux.Router, db *sql.DB, hub *websocket.Hub) {
	h := NewHandler(db, hub)

	// Auth routes
	r.HandleFunc("/api/v1/register", h.RegisterUser).Methods("POST")
	r.HandleFunc("/api/v1/login", h.LoginUser).Methods("POST")

	// Room routes
	r.HandleFunc("/api/v1/rooms", h.CreateRoom).Methods("POST")
	r.HandleFunc("/api/v1/rooms", h.GetUserRooms).Methods("GET")
	r.HandleFunc("/api/v1/rooms/{roomId}/messages", h.GetMessages).Methods("GET")
	r.HandleFunc("/api/v1/rooms/{roomId}/messages", h.SendMessage).Methods("POST")
	r.HandleFunc("/api/v1/rooms/{roomId}/messages/read", h.MarkRoomMessagesAsRead).Methods("POST")

	// Message status routes
	r.HandleFunc("/api/v1/messages/status", h.UpdateMessageStatus).Methods("POST")
	r.HandleFunc("/api/v1/messages/{messageId}/delete", h.DeleteMessage).Methods("DELETE")

	// Direct chat routes
	r.HandleFunc("/api/v1/users", h.SearchUsers).Methods("GET")
	r.HandleFunc("/api/v1/direct-chat", h.GetOrCreateDirectChat).Methods("POST")
	r.HandleFunc("/api/v1/direct-chats", h.GetUserDirectChats).Methods("GET")

	// WebSocket
	r.HandleFunc("/ws", websocket.HandleWebSocket)
}
