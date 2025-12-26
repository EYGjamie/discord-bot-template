package server

import (
	"discord-bot-template/backend/handlers"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Initialize auth handler
	authHandler := handlers.NewAuthHandler()

	// Auth routes (public)
	mux.HandleFunc("GET /api/auth/discord/login", authHandler.DiscordLogin)
	mux.HandleFunc("GET /api/auth/discord/callback", authHandler.DiscordCallback)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.HandleFunc("GET /api/me", authHandler.GetCurrentUser)

	// Member routes (protected)
	mux.HandleFunc("GET /api/members", handlers.GetMembers(s.db.DB()))
	mux.HandleFunc("GET /api/members/{id}", handlers.GetMemberByID(s.db.DB()))
	mux.HandleFunc("GET /api/members/{id}/stats", handlers.GetMemberStats(s.db.DB()))

	// Moderation routes (protected)
	mux.HandleFunc("POST /api/moderation/warns", handlers.CreateWarn(s.db.DB()))
	mux.HandleFunc("POST /api/moderation/notes", handlers.CreateNote(s.db.DB()))
	mux.HandleFunc("DELETE /api/moderation/warns/{id}", handlers.DeleteWarn(s.db.DB()))
	mux.HandleFunc("DELETE /api/moderation/notes/{id}", handlers.DeleteNote(s.db.DB()))

	// Health check
	mux.HandleFunc("/health", s.healthHandler)

	// Original routes for testing
	mux.HandleFunc("/", s.HelloWorldHandler)
	mux.HandleFunc("/websocket", s.websocketHandler)

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-User-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "Hello World"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	healthData := s.db.Health()
	resp, err := json.Marshal(healthData)
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	socket, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to open websocket", http.StatusInternalServerError)
		return
	}
	defer socket.Close(websocket.StatusGoingAway, "Server closing websocket")

	ctx := r.Context()
	socketCtx := socket.CloseRead(ctx)

	for {
		payload := fmt.Sprintf("server timestamp: %d", time.Now().UnixNano())
		if err := socket.Write(socketCtx, websocket.MessageText, []byte(payload)); err != nil {
			log.Printf("Failed to write to socket: %v", err)
			break
		}
		time.Sleep(2 * time.Second)
	}
}
