package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/matrix-messenger/server/internal/api"
	"github.com/matrix-messenger/server/internal/database"
	"github.com/matrix-messenger/server/internal/websocket"
)

func main() {
	// Инициализация базы данных
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := database.NewPostgresDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Инициализация WebSocket хабa с базой данных
	hub := websocket.NewHub(db.DB)
	go hub.Run()

	// Создание HTTP маршрутизатора с gorilla/mux
	r := mux.NewRouter()

	// Регистрация API handlers
	api.RegisterRoutes(r, db.DB, hub)

	// Устанавливаем глобальный хаб для WebSocket
	websocket.SetHub(hub)

	// Запуск сервера
	port := ":8008"
	log.Printf("Matrix Messenger Server starting on port %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
