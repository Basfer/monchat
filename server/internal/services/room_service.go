package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/matrix-messenger/server/internal/models"
)

type RoomService struct {
	DB *sql.DB
}

func NewRoomService(db *sql.DB) *RoomService {
	return &RoomService{DB: db}
}

func (s *RoomService) CreateRoom(name, description string, roomType models.RoomType, createdBy string) (*models.Room, error) {
	id := uuid.New().String()
	now := time.Now()

	var namePtr, descPtr *string
	if name != "" {
		namePtr = &name
	}
	if description != "" {
		descPtr = &description
	}

	res, err := s.DB.Exec(
		"INSERT INTO rooms (id, name, description, type, created_by, is_private, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		id, namePtr, descPtr, roomType, createdBy, false, now, now,
	)
	if err != nil {
		return nil, err
	}
	_ = res

	// Добавляем создателя как администратора
	_, err = s.DB.Exec(
		"INSERT INTO room_members (room_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4)",
		id, createdBy, models.RoleCreator, now,
	)
	if err != nil {
		return nil, err
	}

	return &models.Room{
		ID:          id,
		Name:        namePtr,
		Description: descPtr,
		Type:        roomType,
		CreatedBy:   createdBy,
		IsPrivate:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetUserRooms возвращает список комнат пользователя с опциональной фильтрацией по типу
func (s *RoomService) GetUserRooms(userID string, roomType *string) ([]models.Room, error) {
	var rows *sql.Rows
	var err error

	if roomType != nil && *roomType != "" {
		rows, err = s.DB.Query(`
			SELECT r.id, r.name, r.description, r.type, r.created_by, r.is_private, r.created_at, r.updated_at,
			       (SELECT COUNT(*) FROM room_members rm WHERE rm.room_id = r.id) as member_count
			FROM rooms r
			JOIN room_members rm ON r.id = rm.room_id
			WHERE rm.user_id = $1 AND r.type = $2
			ORDER BY r.updated_at DESC
		`, userID, *roomType)
	} else {
		rows, err = s.DB.Query(`
			SELECT r.id, r.name, r.description, r.type, r.created_by, r.is_private, r.created_at, r.updated_at,
			       (SELECT COUNT(*) FROM room_members rm WHERE rm.room_id = r.id) as member_count
			FROM rooms r
			JOIN room_members rm ON r.id = rm.room_id
			WHERE rm.user_id = $1
			ORDER BY r.updated_at DESC
		`, userID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]models.Room, 0)
	for rows.Next() {
		var r models.Room
		var mbCount int
		err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Type, &r.CreatedBy, &r.IsPrivate, &r.CreatedAt, &r.UpdatedAt, &mbCount)
		if err != nil {
			return nil, err
		}
		r.MemberCount = mbCount
		rooms = append(rooms, r)
	}
	return rooms, nil
}

// GetDirectChatsForUser returns all direct chats for a user with the other participant's info
func (s *RoomService) GetDirectChatsForUser(userID string) ([]models.Room, error) {
	query := `
		SELECT r.id, r.name, r.description, r.type, r.created_by, r.is_private, r.created_at, r.updated_at,
		       (SELECT COUNT(*) FROM room_members rm WHERE rm.room_id = r.id) as member_count
		FROM rooms r
		JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1 AND r.type = 'direct'
		ORDER BY r.updated_at DESC
	`

	rows, err := s.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]models.Room, 0)
	for rows.Next() {
		var r models.Room
		var mbCount int
		err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Type, &r.CreatedBy, &r.IsPrivate, &r.CreatedAt, &r.UpdatedAt, &mbCount)
		if err != nil {
			return nil, err
		}
		r.MemberCount = mbCount
		rooms = append(rooms, r)
	}
	return rooms, nil
}

func (s *RoomService) AddMember(roomID, userID string, role models.RoomRole) error {
	_, err := s.DB.Exec(
		"INSERT INTO room_members (room_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
		roomID, userID, role, time.Now(),
	)
	return err
}

func (s *RoomService) GetMessages(roomID string, limit int, start string) ([]models.Message, error) {
	query := `
		SELECT m.id, m.room_id, m.sender_id, u.username as sender_name, m.content, m.type, m.status, m.created_at
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.room_id = $1 AND m.deleted = false
		ORDER BY m.created_at DESC
		LIMIT $2
	`
	if start != "" {
		query += " AND m.id < $3"
		rows, err := s.DB.Query(query, roomID, limit, start)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanMessages(rows)
	}

	rows, err := s.DB.Query(query, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMessages(rows)
}

func scanMessages(rows *sql.Rows) ([]models.Message, error) {
	var msgs []models.Message
	for rows.Next() {
		var m models.Message
		var senderName string
		var status string
		err := rows.Scan(&m.ID, &m.RoomID, &m.SenderID, &senderName, &m.Content, &m.Type, &status, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		m.SenderName = senderName
		m.Status = models.MessageStatus(status)
		msgs = append(msgs, m)
	}
	return msgs, nil
}

// UpdateMessageStatus updates the status of a message for a specific user
func (s *RoomService) UpdateMessageStatus(messageID, userID, roomID string, status models.MessageStatus) error {
	// Check if user is a member of the room
	var isMember bool
	err := s.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)
	`, roomID, userID).Scan(&isMember)
	if err != nil {
		return err
	}
	if !isMember {
		return fmt.Errorf("user is not a member of the room")
	}

	_, err = s.DB.Exec(`
		INSERT INTO message_read_status (message_id, user_id, status, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (message_id, user_id)
		DO UPDATE SET status = EXCLUDED.status, updated_at = NOW()
	`, messageID, userID, string(status))
	return err
}

// MarkMessagesAsRead marks all messages in a room as read for a user
func (s *RoomService) MarkMessagesAsRead(roomID, userID string) error {
	rows, err := s.DB.Query(`
		SELECT id FROM messages WHERE room_id = $1 AND deleted = false
	`, roomID)
	if err != nil {
		return err
	}
	defer rows.Close()

	messageIDs := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		messageIDs = append(messageIDs, id)
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO message_read_status (message_id, user_id, status, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (message_id, user_id)
		DO UPDATE SET status = EXCLUDED.status, updated_at = NOW()
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range messageIDs {
		_, err := stmt.Exec(id, userID, string(models.MessageStatusRead))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SoftDeleteMessage marks a message as deleted for the sender
func (s *RoomService) SoftDeleteMessage(messageID, userID string) error {
	var id string
	err := s.DB.QueryRow(`
		UPDATE messages SET deleted = true WHERE id = $1 AND sender_id = $2 AND deleted = false
		RETURNING id
	`, messageID, userID).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}
