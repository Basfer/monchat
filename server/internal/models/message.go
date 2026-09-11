package models

import (
	"time"
)

type MessageType string

const (
	MessageTypeText   MessageType = "m.text"
	MessageTypeEdited MessageType = "m.emote" // marker for edited message
)

type MessageStatus string

const (
	MessageStatusNew       MessageStatus = "new"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusDeleted   MessageStatus = "deleted"
)

type Message struct {
	ID         string        `json:"id"`
	RoomID     string        `json:"room_id"`
	SenderID   string        `json:"sender_id"`
	SenderName string        `json:"sender_name,omitempty"`
	Content    string        `json:"content"`
	Type       MessageType   `json:"type"`
	Status     MessageStatus `json:"status"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  *time.Time    `json:"updatedAt,omitempty"`
	Deleted    bool          `json:"deleted,omitempty"`
}
