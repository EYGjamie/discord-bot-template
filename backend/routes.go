package server

import (
	"discord-bot-template/backend/handlers"
	"discord-bot-template/backend/middleware"
	"discord-bot-template/shared/database/tables"
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
	mux.HandleFunc("GET /api/discord/roles-and-members", middleware.RequireAuth(handlers.GetDiscordRolesAndMembers(s.db.DB())))
	mux.HandleFunc("GET /api/discord/members/search", middleware.RequireAuth(handlers.SearchMembers(s.db.DB())))
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

	// Task Management routes
	boardsHandler := &handlers.BoardsHandler{DB: s.db.DB()}
	tasksHandler := &handlers.TasksHandler{DB: s.db.DB()}
	taskGroupsHandler := &handlers.TaskGroupsHandler{DB: s.db.DB()}
	notificationSettingsHandler := &handlers.NotificationSettingsHandler{DB: s.db.DB()}
	taskPermChecker := &middleware.TaskPermissionChecker{DB: s.db.DB()}
	boardPermChecker := &middleware.BoardPermissionChecker{DB: s.db.DB()}
	requireAuthWithDB := middleware.RequireAuthWithDB(s.db.DB())

	// Board routes (protected)
	mux.HandleFunc("GET /api/boards", requireAuthWithDB(http.HandlerFunc(boardsHandler.GetBoards)))
	mux.HandleFunc("GET /api/boards/{id}",
		requireAuthWithDB(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				boardPermChecker.RequireBoardView()(http.HandlerFunc(boardsHandler.GetBoard)).ServeHTTP(w, r)
			}),
		),
	)
	mux.HandleFunc("POST /api/boards",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(boardsHandler.CreateBoard),
			),
		),
	)
	mux.HandleFunc("PUT /api/boards/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(boardsHandler.UpdateBoard),
			),
		),
	)
	mux.HandleFunc("DELETE /api/boards/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(boardsHandler.DeleteBoard),
			),
		),
	)

	// Board permission routes (admin only)
	mux.HandleFunc("GET /api/boards/{id}/permissions",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(boardsHandler.GetBoardPermissions),
			),
		),
	)
	mux.HandleFunc("POST /api/boards/{id}/permissions",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(boardsHandler.SetBoardPermission),
			),
		),
	)
	mux.HandleFunc("DELETE /api/boards/{id}/permissions/{permissionId}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(boardsHandler.DeleteBoardPermission),
			),
		),
	)
	mux.HandleFunc("PUT /api/boards/{id}/permissions/{permissionId}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(boardsHandler.UpdateBoardPermission),
			),
		),
	)

	// Task routes (protected with granular permissions)
	mux.HandleFunc("GET /api/boards/{boardId}/tasks",
		requireAuthWithDB(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				boardPermChecker.RequireBoardView()(http.HandlerFunc(tasksHandler.GetBoardTasks)).ServeHTTP(w, r)
			}),
		),
	)
	mux.HandleFunc("GET /api/tasks/{id}",
		middleware.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				taskPermChecker.RequireTaskPermission(tables.PermissionReadContent)(http.HandlerFunc(tasksHandler.GetTask)).ServeHTTP(w, r)
			}),
		),
	)
	mux.HandleFunc("POST /api/tasks", middleware.RequireAuth(http.HandlerFunc(tasksHandler.CreateTask)))
	mux.HandleFunc("PUT /api/tasks/{id}",
		middleware.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				taskPermChecker.RequireTaskPermission(tables.PermissionEdit)(http.HandlerFunc(tasksHandler.UpdateTask)).ServeHTTP(w, r)
			}),
		),
	)
	mux.HandleFunc("PUT /api/tasks/{id}/move",
		middleware.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				taskPermChecker.RequireTaskPermission(tables.PermissionEdit)(http.HandlerFunc(tasksHandler.MoveTask)).ServeHTTP(w, r)
			}),
		),
	)
	mux.HandleFunc("DELETE /api/tasks/{id}",
		middleware.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				taskPermChecker.RequireTaskPermission(tables.PermissionDelete)(http.HandlerFunc(tasksHandler.DeleteTask)).ServeHTTP(w, r)
			}),
		),
	)

	// Task Comments routes (protected with task view permission)
	taskCommentsHandler := &handlers.TaskCommentsHandler{DB: s.db.DB()}
	mux.HandleFunc("GET /api/tasks/{taskId}/comments",
		middleware.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				taskPermChecker.RequireTaskPermission(tables.PermissionReadContent)(http.HandlerFunc(taskCommentsHandler.GetComments)).ServeHTTP(w, r)
			}),
		),
	)
	mux.HandleFunc("POST /api/tasks/{taskId}/comments",
		middleware.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				taskPermChecker.RequireTaskPermission(tables.PermissionEdit)(http.HandlerFunc(taskCommentsHandler.CreateComment)).ServeHTTP(w, r)
			}),
		),
	)
	mux.HandleFunc("PUT /api/comments/{id}",
		middleware.RequireAuth(http.HandlerFunc(taskCommentsHandler.UpdateComment)),
	)
	mux.HandleFunc("DELETE /api/comments/{id}",
		middleware.RequireAuth(http.HandlerFunc(taskCommentsHandler.DeleteComment)),
	)

	// Task Checklist routes (protected with task view permission)
	taskChecklistHandler := &handlers.TaskChecklistHandler{DB: s.db.DB()}
	mux.HandleFunc("GET /api/tasks/{taskId}/checklist",
		middleware.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				taskPermChecker.RequireTaskPermission(tables.PermissionReadContent)(http.HandlerFunc(taskChecklistHandler.GetChecklistItems)).ServeHTTP(w, r)
			}),
		),
	)
	mux.HandleFunc("POST /api/tasks/{taskId}/checklist",
		middleware.RequireAuth(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				taskPermChecker.RequireTaskPermission(tables.PermissionEdit)(http.HandlerFunc(taskChecklistHandler.CreateChecklistItem)).ServeHTTP(w, r)
			}),
		),
	)
	mux.HandleFunc("PUT /api/checklist/{id}",
		middleware.RequireAuth(http.HandlerFunc(taskChecklistHandler.UpdateChecklistItem)),
	)
	mux.HandleFunc("POST /api/checklist/{id}/toggle",
		middleware.RequireAuth(http.HandlerFunc(taskChecklistHandler.ToggleChecklistItem)),
	)
	mux.HandleFunc("DELETE /api/checklist/{id}",
		middleware.RequireAuth(http.HandlerFunc(taskChecklistHandler.DeleteChecklistItem)),
	)

	// Task group routes (admin only)
	mux.HandleFunc("GET /api/task-groups", middleware.RequireAuth(http.HandlerFunc(taskGroupsHandler.GetTaskGroups)))
	mux.HandleFunc("GET /api/task-groups/{id}", middleware.RequireAuth(http.HandlerFunc(taskGroupsHandler.GetTaskGroup)))
	mux.HandleFunc("POST /api/task-groups",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(taskGroupsHandler.CreateTaskGroup),
			),
		),
	)
	mux.HandleFunc("PUT /api/task-groups/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(taskGroupsHandler.UpdateTaskGroup),
			),
		),
	)
	mux.HandleFunc("DELETE /api/task-groups/{id}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(taskGroupsHandler.DeleteTaskGroup),
			),
		),
	)

	// Task group permission routes (admin only)
	mux.HandleFunc("GET /api/task-groups/{id}/permissions",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(taskGroupsHandler.GetTaskGroupPermissions),
			),
		),
	)
	mux.HandleFunc("POST /api/task-groups/{id}/permissions",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(taskGroupsHandler.SetTaskGroupPermission),
			),
		),
	)
	mux.HandleFunc("DELETE /api/task-groups/{id}/permissions/{permissionId}",
		middleware.RequireAuth(
			permissionChecker.RequirePermission(middleware.PermissionAdmin)(
				http.HandlerFunc(taskGroupsHandler.DeleteTaskGroupPermission),
			),
		),
	)

	// Notification Settings routes (user-specific, protected)
	mux.HandleFunc("GET /api/notification-settings",
		middleware.RequireAuth(http.HandlerFunc(notificationSettingsHandler.GetNotificationSettings)),
	)
	mux.HandleFunc("PUT /api/notification-settings",
		middleware.RequireAuth(http.HandlerFunc(notificationSettingsHandler.UpdateNotificationSettings)),
	)
	mux.HandleFunc("GET /api/notification-settings/boards/{boardId}",
		middleware.RequireAuth(http.HandlerFunc(notificationSettingsHandler.GetBoardNotificationSettings)),
	)
	mux.HandleFunc("PUT /api/notification-settings/boards/{boardId}",
		middleware.RequireAuth(http.HandlerFunc(notificationSettingsHandler.UpdateBoardNotificationSettings)),
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
