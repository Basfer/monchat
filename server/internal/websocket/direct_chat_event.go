package websocket

import (
	"encoding/json"
	"log"
)

// DirectChatCreatedEvent represents a WebSocket event when a new direct chat is created
type DirectChatCreatedEvent struct {
	Type             string        `json:"type"`
	RoomID           string        `json:"room_id"`
	Participant1ID   string        `json:"participant_1_id"`
	Participant2ID   string        `json:"participant_2_id"`
	OtherParticipant UserShortInfo `json:"other_participant"`
}

// UserShortInfo contains minimal user info for WS events
type UserShortInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// SendDirectChatCreatedToUser sends a direct_chat_created event to a specific user regardless of their current room
func (h *Hub) SendDirectChatCreatedToUser(userID string, roomID string, event DirectChatCreatedEvent) {
	eventData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling direct_chat_created event: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.clients[userID]; ok {
		select {
		case client.send <- eventData:
		default:
			log.Printf("Failed to send direct_chat_created to user %s", userID)
		}
	}
}

// SendDirectChatCreatedEvent sends a direct_chat_created event to both participants
func (h *Hub) SendDirectChatCreatedEvent(roomID, participant1ID, participant2ID string, otherParticipant *UserShortInfo) {
	event := DirectChatCreatedEvent{
		Type:             "direct_chat_created",
		RoomID:           roomID,
		Participant1ID:   participant1ID,
		Participant2ID:   participant2ID,
		OtherParticipant: *otherParticipant,
	}

	// Send to participant 1 if they are connected
	h.SendDirectChatCreatedToUser(participant1ID, roomID, event)

	// Send to participant 2 if they are connected
	h.SendDirectChatCreatedToUser(participant2ID, roomID, event)
}
