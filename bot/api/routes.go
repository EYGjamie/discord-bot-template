package api

import (
	"net/http"

	"discord-bot-template/bot/api/handlers"
)

// RegisterRoutes registriert alle API-Routen
func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health-Check Endpoint
	mux.HandleFunc("/health", handlers.HealthHandler(s.botSession, s.botDB))

	// Bot API Endpoints
	mux.HandleFunc("/api/guild/sync", handlers.GuildSyncHandler(s.botSession, s.botDB))

	// Wrap mit CORS Middleware
	return s.corsMiddleware(mux)
}
