package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Server struct {
	port       int
	server     *http.Server
	botSession *discordgo.Session
	botDB      *sql.DB
}

// NewServer erstellt einen neuen API-Server für den Bot
func NewServer(session *discordgo.Session, db *sql.DB) *Server {
	port, _ := strconv.Atoi(os.Getenv("BOT_API_PORT"))
	if port == 0 {
		port = 8090 // Default port für Bot-API
	}

	srv := &Server{
		port:       port,
		botSession: session,
		botDB:      db,
	}

	// Konfiguriere HTTP-Server
	srv.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      srv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return srv
}

// Start startet den API-Server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown fährt den API-Server sauber herunter
func (s *Server) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// GetAddr gibt die Server-Adresse zurück
func (s *Server) GetAddr() string {
	return s.server.Addr
}
