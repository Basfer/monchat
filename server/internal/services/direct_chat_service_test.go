package services

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/matrix-messenger/server/internal/models"
)

// TestDirectChatService_CannotCreateChatWithSelf tests that a user cannot create a direct chat with themselves
func TestDirectChatService_CannotCreateChatWithSelf(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	db, err := sql.Open("postgres", "host=localhost port=5432 user=monchat password=monchat123 dbname=monchat sslmode=disable")
	if err != nil {
		t.Skipf("Could not connect to database: %v", err)
	}
	defer db.Close()

	service := NewDirectChatService(db)

	// Create two test users
	user1 := createTestUserInline(t, db, "testuser1_"+t.Name())
	user2 := createTestUserInline(t, db, "testuser2_"+t.Name())
	defer deleteUserInline(t, db, user1.ID)
	defer deleteUserInline(t, db, user2.ID)

	// Try to create a direct chat with yourself
	_, err = service.GetOrCreateDirectChat(user1.ID, user1.ID)
	if err == nil {
		t.Error("Expected error when creating direct chat with yourself, got nil")
	}
}

// TestDirectChatService_CreateNewDirectChat tests creating a new direct chat between two users
func TestDirectChatService_CreateNewDirectChat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	db, err := sql.Open("postgres", "host=localhost port=5432 user=monchat password=monchat123 dbname=monchat sslmode=disable")
	if err != nil {
		t.Skipf("Could not connect to database: %v", err)
	}
	defer db.Close()

	service := NewDirectChatService(db)

	user1 := createTestUserInline(t, db, "testuser_dm1_"+t.Name())
	user2 := createTestUserInline(t, db, "testuser_dm2_"+t.Name())
	defer deleteUserInline(t, db, user1.ID)
	defer deleteUserInline(t, db, user2.ID)

	room, err := service.GetOrCreateDirectChat(user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("Failed to create direct chat: %v", err)
	}

	if room.Type != models.RoomTypeDirect {
		t.Errorf("Expected room type 'direct', got '%s'", room.Type)
	}

	if room.MemberCount != 2 {
		t.Errorf("Expected member count 2, got %d", room.MemberCount)
	}

	if len(room.Participants) != 2 {
		t.Errorf("Expected 2 participants, got %d", len(room.Participants))
	}
}

// TestDirectChatService_Deduplication tests that the same direct chat is not created twice
func TestDirectChatService_Deduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	db, err := sql.Open("postgres", "host=localhost port=5432 user=monchat password=monchat123 dbname=monchat sslmode=disable")
	if err != nil {
		t.Skipf("Could not connect to database: %v", err)
	}
	defer db.Close()

	service := NewDirectChatService(db)

	user1 := createTestUserInline(t, db, "testuser_dup1_"+t.Name())
	user2 := createTestUserInline(t, db, "testuser_dup2_"+t.Name())
	defer deleteUserInline(t, db, user1.ID)
	defer deleteUserInline(t, db, user2.ID)

	// Create first direct chat
	room1, err := service.GetOrCreateDirectChat(user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("Failed to create first direct chat: %v", err)
	}

	// Create second direct chat (should return existing)
	room2, err := service.GetOrCreateDirectChat(user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("Failed to get existing direct chat: %v", err)
	}

	if room1.ID != room2.ID {
		t.Errorf("Expected same room ID, got %s and %s", room1.ID, room2.ID)
	}
}

// TestDirectChatService_OrderIndependence tests that the order of user IDs doesn't matter
func TestDirectChatService_OrderIndependence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	db, err := sql.Open("postgres", "host=localhost port=5432 user=monchat password=monchat123 dbname=monchat sslmode=disable")
	if err != nil {
		t.Skipf("Could not connect to database: %v", err)
	}
	defer db.Close()

	service := NewDirectChatService(db)

	user1 := createTestUserInline(t, db, "testuser_ord1_"+t.Name())
	user2 := createTestUserInline(t, db, "testuser_ord2_"+t.Name())
	defer deleteUserInline(t, db, user1.ID)
	defer deleteUserInline(t, db, user2.ID)

	// Create chat with user1 first
	room1, err := service.GetOrCreateDirectChat(user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("Failed to create direct chat (user1 first): %v", err)
	}

	// Try to create chat with user2 first (should return same room)
	room2, err := service.GetOrCreateDirectChat(user2.ID, user1.ID)
	if err != nil {
		t.Fatalf("Failed to get existing direct chat (user2 first): %v", err)
	}

	if room1.ID != room2.ID {
		t.Errorf("Expected same room ID regardless of order, got %s and %s", room1.ID, room2.ID)
	}
}

// TestDirectChatService_GetOtherParticipantInfo tests getting the other participant's info
func TestDirectChatService_GetOtherParticipantInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	db, err := sql.Open("postgres", "host=localhost port=5432 user=monchat password=monchat123 dbname=monchat sslmode=disable")
	if err != nil {
		t.Skipf("Could not connect to database: %v", err)
	}
	defer db.Close()

	service := NewDirectChatService(db)

	user1 := createTestUserInline(t, db, "testuser_part1_"+t.Name())
	user2 := createTestUserInline(t, db, "testuser_part2_"+t.Name())
	defer deleteUserInline(t, db, user1.ID)
	defer deleteUserInline(t, db, user2.ID)

	room, err := service.GetOrCreateDirectChat(user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("Failed to create direct chat: %v", err)
	}

	// Get other participant info for user1 (should be user2)
	otherParticipant, err := service.GetOtherParticipantInfo(room.ID, user1.ID)
	if err != nil {
		t.Fatalf("Failed to get other participant info: %v", err)
	}

	if otherParticipant.ID != user2.ID {
		t.Errorf("Expected other participant ID %s, got %s", user2.ID, otherParticipant.ID)
	}

	if otherParticipant.Username != user2.Username {
		t.Errorf("Expected other participant username %s, got %s", user2.Username, otherParticipant.Username)
	}
}

// Helper functions

func createTestUserInline(t *testing.T, db *sql.DB, username string) *models.User {
	t.Helper()

	// Generate unique email
	email := username + "@test.com"
	passwordHash := "$2a$10$dummyhashfor_testing" // Dummy hash for test purposes

	var userID string
	err := db.QueryRow(
		"INSERT INTO users (id, username, email, password_hash, created_at, updated_at) VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW()) RETURNING id",
		username, email, passwordHash,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user %s: %v", username, err)
	}

	return &models.User{
		ID:       userID,
		Username: username,
		Email:    email,
	}
}

func deleteUserInline(t *testing.T, db *sql.DB, userID string) {
	t.Helper()
	// Delete room members first, then rooms, then user
	db.Exec("DELETE FROM room_members WHERE user_id = $1", userID)
	db.Exec("DELETE FROM messages WHERE sender_id = $1", userID)
	db.Exec("DELETE FROM rooms WHERE created_by = $1", userID)
	db.Exec("DELETE FROM users WHERE id = $1", userID)
}
