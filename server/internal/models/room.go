package models

import (
	"errors"
	"time"
)

type RoomType string

const (
	RoomTypeDirect  RoomType = "direct"
	RoomTypeGroup   RoomType = "group"
	RoomTypeChannel RoomType = "channel"
	RoomTypePublic  RoomType = "public"
)

type RoomRole string

const (
	RoleCreator    RoomRole = "creator"
	RoleAdmin      RoomRole = "admin"
	RoleMember     RoomRole = "member"
	RoleSubscriber RoomRole = "subscriber"
)

type Room struct {
	ID          string    `json:"id"`
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Type        RoomType  `json:"type"`
	CreatedBy   string    `json:"created_by"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Metadata for UI
	MemberCount int      `json:"member_count,omitempty"`
	LastMessage *Message `json:"last_message,omitempty"`

	// Participants holds user IDs for direct chats explicitly
	Participants []string `json:"participants,omitempty"`
}

type RoomMember struct {
	RoomID   string    `json:"room_id"`
	UserID   string    `json:"user_id"`
	Role     RoomRole  `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

// IsDirect checks if the room is a direct chat
func (r *Room) IsDirect() bool {
	return r.Type == RoomTypeDirect
}

// GetOtherParticipant returns the ID of the other participant in a direct chat
func (r *Room) GetOtherParticipant(myUserID string) (string, error) {
	if !r.IsDirect() {
		return "", errors.New("room is not a direct chat")
	}

	for _, participantID := range r.Participants {
		if participantID != myUserID {
			return participantID, nil
		}
	}

	return "", errors.New("other participant not found")
}
