package services

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/matrix-messenger/server/internal/models"
)

type DirectChatService struct {
	DB *sql.DB
}

func NewDirectChatService(db *sql.DB) *DirectChatService {
	return &DirectChatService{DB: db}
}

// GetOrCreateDirectChat returns existing or creates a new direct chat between two users
func (s *DirectChatService) GetOrCreateDirectChat(userID1, userID2 string) (*models.Room, error) {
	if userID1 == userID2 {
		return nil, errors.New("cannot create direct chat with yourself")
	}

	existingRoom, err := s.FindExistingDirectChat(userID1, userID2)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if existingRoom != nil {
		return existingRoom, nil
	}

	return s.createDirectChat(userID1, userID2)
}

// FindExistingDirectChat finds an existing direct chat between two users
func (s *DirectChatService) FindExistingDirectChat(userID1, userID2 string) (*models.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.type, r.created_by, r.is_private, r.created_at, r.updated_at
		FROM rooms r
		JOIN room_members rm1 ON r.id = rm1.room_id AND rm1.user_id = $1
		JOIN room_members rm2 ON r.id = rm2.room_id AND rm2.user_id = $2
		WHERE r.type = 'direct'
		LIMIT 1
	`

	var r models.Room
	var mbCount int
	err := s.DB.QueryRow(query, userID1, userID2).Scan(
		&r.ID, &r.Name, &r.Description, &r.Type, &r.CreatedBy, &r.IsPrivate, &r.CreatedAt, &r.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	r.Participants = []string{userID1, userID2}

	mbCountQuery := `SELECT COUNT(*) FROM room_members WHERE room_id = $1`
	err = s.DB.QueryRow(mbCountQuery, r.ID).Scan(&mbCount)
	if err != nil {
		return nil, err
	}
	r.MemberCount = mbCount

	return &r, nil
}

// createDirectChat creates a new direct chat room between two users
func (s *DirectChatService) createDirectChat(userID1, userID2 string) (*models.Room, error) {
	id := uuid.New().String()
	now := time.Now()

	res, err := s.DB.Exec(
		"INSERT INTO rooms (id, name, description, type, created_by, is_private, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		id, nil, nil, models.RoomTypeDirect, userID1, true, now, now,
	)
	if err != nil {
		return nil, err
	}
	_ = res

	_, err = s.DB.Exec(
		"INSERT INTO room_members (room_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4), ($1, $5, $3, $4)",
		id, userID1, models.RoleMember, now, userID2,
	)
	if err != nil {
		return nil, err
	}

	room := &models.Room{
		ID:           id,
		Type:         models.RoomTypeDirect,
		CreatedBy:    userID1,
		IsPrivate:    true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Participants: []string{userID1, userID2},
		MemberCount:  2,
	}

	return room, nil
}

// GetDirectChatsWithUser returns all direct chats between two specific users
func (s *DirectChatService) GetDirectChatsWithUser(userID, otherUserID string) ([]models.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.type, r.created_by, r.is_private, r.created_at, r.updated_at
		FROM rooms r
		JOIN room_members rm ON r.id = rm.room_id
		WHERE r.type = 'direct'
		AND (rm.user_id = $1 OR rm.user_id = $2)
		ORDER BY r.updated_at DESC
	`

	rows, err := s.DB.Query(query, userID, otherUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]models.Room, 0)
	for rows.Next() {
		var r models.Room
		var mbCount int
		err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Type, &r.CreatedBy, &r.IsPrivate, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			return nil, err
		}

		r.Participants = []string{userID, otherUserID}

		mbCountQuery := `SELECT COUNT(*) FROM room_members WHERE room_id = $1`
		err = s.DB.QueryRow(mbCountQuery, r.ID).Scan(&mbCount)
		if err != nil {
			return nil, err
		}
		r.MemberCount = mbCount

		rooms = append(rooms, r)
	}

	return rooms, nil
}

// GetOtherParticipantInfo retrieves the username and email of the other participant in a direct chat
func (s *DirectChatService) GetOtherParticipantInfo(roomID, myUserID string) (*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.updated_at
		FROM users u
		JOIN room_members rm ON u.id = rm.user_id
		WHERE rm.room_id = $1
		AND rm.user_id != $2
		LIMIT 1
	`

	var user models.User
	err := s.DB.QueryRow(query, roomID, myUserID).Scan(
		&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
