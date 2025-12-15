package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"discord-bot-template/internal/shared/utils/logging"

	"github.com/coder/websocket"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/", s.HelloWorldHandler)

	mux.HandleFunc("/health", s.healthHandler)

	mux.HandleFunc("/websocket", s.websocketHandler)

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

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
	logger := logging.NewLogger(s.db.DB(), nil, "", "api.hello")
	logger.LogInfo("API Request", fmt.Sprintf("GET / from %s", r.RemoteAddr), false)

	resp := map[string]string{"message": "Hello World"}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		logger.LogError("JSON Marshal Failed", fmt.Sprintf("Failed to marshal response: %v", err), "")
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(jsonResp); err != nil {
		log.Printf("Failed to write response: %v", err)
		logger.LogError("Response Write Failed", fmt.Sprintf("Failed to write response: %v", err), "")
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.NewLogger(s.db.DB(), nil, "", "api.health")

	healthData := s.db.Health()
	resp, err := json.Marshal(healthData)
	if err != nil {
		logger.LogError("Health Check Failed", fmt.Sprintf("Failed to marshal health response: %v", err), "")
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}

	// Logge wenn DB down ist
	if status, ok := healthData["status"]; ok && status == "down" {
		logger.LogError("Database Unhealthy", "Health check reports database is down", "")
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
		logger.LogError("Response Write Failed", fmt.Sprintf("Failed to write health response: %v", err), "")
	}
}

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	logger := logging.NewLogger(s.db.DB(), nil, "", "api.websocket")
	logger.LogInfo("WebSocket Connection", fmt.Sprintf("New connection from %s", r.RemoteAddr), false)

	socket, err := websocket.Accept(w, r, nil)
	if err != nil {
		logger.LogError("WebSocket Accept Failed", fmt.Sprintf("Failed to accept websocket: %v", err), "")
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
			logger.LogError("WebSocket Write Failed", fmt.Sprintf("Failed to write to socket: %v", err), "")
			break
		}
		time.Sleep(2 * time.Second)
	}
	logger.LogInfo("WebSocket Closed", fmt.Sprintf("Connection from %s closed", r.RemoteAddr), false)
}
