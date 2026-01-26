package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"discord-bot-template/backend/services"
	"discord-bot-template/shared/database"
)

type Server struct {
	port int

	db        database.Service
	scheduler *services.DiscordStatsScheduler
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		db: database.New(),
	}

	// Initialize database tables (idempotent - safe to call multiple times)
	db := NewServer.db.DB()
	if err := database.InitializeTables(db); err != nil {
		log.Printf("Warning: Failed to initialize database tables: %v", err)
	}

	// Initialisiere und starte den Discord Stats Scheduler
	NewServer.scheduler = services.NewDiscordStatsScheduler(NewServer.db.DB())
	NewServer.scheduler.Start()

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}

// Shutdown stoppt den Server und alle Services
func (s *Server) Shutdown() {
	if s.scheduler != nil {
		s.scheduler.Stop()
	}
}
