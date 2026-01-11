package server

import (
	"discord-bot-template/backend/handlers"
	"discord-bot-template/backend/middleware"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coder/websocket"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(s.db.DB())
	auditLogger := middleware.NewAuditLogger(s.db.DB())
	permissionChecker := middleware.NewPermissionChecker(s.db.DB())

	// Auth routes (public)
	mux.HandleFunc("GET /api/auth/discord/login", authHandler.DiscordLogin)
	mux.HandleFunc("GET /api/auth/discord/callback", authHandler.DiscordCallback)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.HandleFunc("GET /api/me", authHandler.GetCurrentUser)

	// Member routes (protected - requires moderator role)
	mux.HandleFunc("GET /api/members",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.GetMembers(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("GET /api/members/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.GetMemberByID(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("GET /api/members/{id}/stats",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.GetMemberStats(s.db.DB()),
			),
		),
	)

	// Moderation routes (protected - requires moderator role)
	mux.HandleFunc("POST /api/moderation/warns",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.CreateWarn(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("POST /api/moderation/notes",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.CreateNote(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("DELETE /api/moderation/warns/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.DeleteWarn(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("DELETE /api/moderation/notes/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.DeleteNote(s.db.DB()),
			),
		),
	)

	// Event routes (protected - all members can access)
	mux.HandleFunc("GET /api/events", handlers.GetEvents(s.db.DB()))
	mux.HandleFunc("GET /api/events/{id}", handlers.GetEventByID(s.db.DB()))
	mux.HandleFunc("POST /api/events", handlers.CreateEvent(s.db.DB()))
	mux.HandleFunc("PUT /api/events/{id}", handlers.UpdateEvent(s.db.DB()))
	mux.HandleFunc("DELETE /api/events/{id}", handlers.DeleteEvent(s.db.DB()))

	// Event Category routes (GET public, POST/PUT/DELETE admin only)
	mux.HandleFunc("GET /api/event-categories", handlers.GetEventCategories(s.db.DB()))
	mux.HandleFunc("POST /api/event-categories",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.CreateEventCategory(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("PUT /api/event-categories/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.UpdateEventCategory(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("DELETE /api/event-categories/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.DeleteEventCategory(s.db.DB()),
			),
		),
	)

	// Match routes (protected - all members can access)
	mux.HandleFunc("GET /api/matches", handlers.GetMatches(s.db.DB()))
	mux.HandleFunc("GET /api/matches/{id}", handlers.GetMatchByID(s.db.DB()))
	mux.HandleFunc("POST /api/matches", handlers.CreateMatch(s.db.DB()))
	mux.HandleFunc("PUT /api/matches/{id}", handlers.UpdateMatch(s.db.DB()))
	mux.HandleFunc("DELETE /api/matches/{id}", handlers.DeleteMatch(s.db.DB()))

	// Match Category routes (GET public, POST/PUT/DELETE admin only)
	mux.HandleFunc("GET /api/match-categories", handlers.GetMatchCategories(s.db.DB()))
	mux.HandleFunc("POST /api/match-categories",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.CreateMatchCategory(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("PUT /api/match-categories/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.UpdateMatchCategory(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("DELETE /api/match-categories/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.DeleteMatchCategory(s.db.DB()),
			),
		),
	)

	// Audit logs routes (protected - admin only)
	mux.HandleFunc("GET /api/audit-logs",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.GetAuditLogs(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("GET /api/audit-logs/user/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.GetUserAuditLogs(s.db.DB()),
			),
		),
	)

	// Discord Statistics routes (protected - moderator access)
	mux.HandleFunc("GET /api/discord/stats/current",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.GetCurrentStats(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("GET /api/discord/stats/historical",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.GetHistoricalStats(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("GET /api/discord/stats/range",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionModerator)(
				handlers.GetStatisticsInRange(s.db.DB()),
			),
		),
	)

	// Bot Settings routes (protected - admin only)
	mux.HandleFunc("GET /api/bot-settings", handlers.GetBotSettings(s.db.DB()))
	mux.HandleFunc("GET /api/discord/roles", handlers.GetDiscordRoles(s.db.DB()))
	mux.HandleFunc("GET /api/discord/channels", handlers.GetDiscordChannels(s.db.DB()))
	mux.HandleFunc("PUT /api/bot-settings/moderator-roles",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.UpdateModeratorRoles(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("PUT /api/bot-settings/moderation",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.UpdateModerationSettings(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("POST /api/bot-settings/create-voice",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.CreateOrUpdateCreateVoiceSetting(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("DELETE /api/bot-settings/create-voice/{channel_id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.DeleteCreateVoiceSetting(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("POST /api/bot-settings/purge",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.CreateOrUpdatePurgeSetting(s.db.DB()),
			),
		),
	)
	mux.HandleFunc("DELETE /api/bot-settings/purge/{channel_id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				handlers.DeletePurgeSetting(s.db.DB()),
			),
		),
	)

	// Health check
	mux.HandleFunc("/health", s.healthHandler)

	// Original routes for testing
	mux.HandleFunc("/", s.HelloWorldHandler)
	mux.HandleFunc("/websocket", s.websocketHandler)

	// Wrap with audit logging middleware, then CORS middleware
	handler := auditLogger.Middleware(mux)
	return s.corsMiddleware(handler)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

		// Default allowed origins if not set
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:5173,http://localhost:3000"
		}

		// Check if origin is allowed
		origins := strings.Split(allowedOrigins, ",")
		originAllowed := false
		for _, allowedOrigin := range origins {
			if strings.TrimSpace(allowedOrigin) == origin {
				originAllowed = true
				break
			}
		}

		// Set CORS headers with specific origin if allowed
		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
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
