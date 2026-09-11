package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	// Создаем таблицы, если их нет
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS rooms (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255),
		description TEXT,
		type VARCHAR(50) NOT NULL DEFAULT 'group', -- direct, group, channel, public
		created_by UUID REFERENCES users(id),
		is_private BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS room_members (
		room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		role VARCHAR(50) NOT NULL DEFAULT 'member', -- creator, admin, member, subscriber
		joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		PRIMARY KEY (room_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		room_id UUID REFERENCES rooms(id) ON DELETE CASCADE,
		sender_id UUID REFERENCES users(id),
		content TEXT NOT NULL,
		type VARCHAR(50) NOT NULL DEFAULT 'm.text',
		status VARCHAR(50) NOT NULL DEFAULT 'new',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE,
		deleted BOOLEAN DEFAULT FALSE
	);

	-- Per-user message read status table
	CREATE TABLE IF NOT EXISTS message_read_status (
		message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
		user_id UUID REFERENCES users(id) ON DELETE CASCADE,
		status VARCHAR(50) NOT NULL DEFAULT 'read',
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		PRIMARY KEY (message_id, user_id)
	);

	-- Indexes for performance
	CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
	CREATE INDEX IF NOT EXISTS idx_rooms_type ON rooms(type);
	CREATE INDEX IF NOT EXISTS idx_room_members_user_id ON room_members(user_id);
	CREATE INDEX IF NOT EXISTS idx_room_members_room_id ON room_members(room_id);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`
	_, err := db.Exec(schema)
	return err
}
