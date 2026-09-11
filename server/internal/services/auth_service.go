package services

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/matrix-messenger/server/internal/models"
)

type AuthService struct {
	DB *sql.DB
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{DB: db}
}

func (s *AuthService) Register(username, email, password string) (*models.User, error) {
	// Проверка уникальности
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2", username, email).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("username or email already exists")
	}

	hash := hashPassword(password)
	id := uuid.New().String()
	now := time.Now()

	_, err = s.DB.Exec(
		"INSERT INTO users (id, username, email, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		id, username, email, hash, now, now,
	)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:           id,
		Username:     username,
		Email:        email,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (s *AuthService) Login(username, password string) (*models.User, error) {
	var user models.User
	var hash string
	
	err := s.DB.QueryRow("SELECT id, username, email, password_hash, created_at, updated_at FROM users WHERE username = $1", username).Scan(
		&user.ID, &user.Username, &user.Email, &hash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if hash != hashPassword(password) {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}