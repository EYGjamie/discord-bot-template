package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"discord-bot-template/backend/middleware"
	"discord-bot-template/shared/database/tables"
)

type NotificationSettingsHandler struct {
	DB *sql.DB
}

// UpdateNotificationSettingsRequest represents the request body for updating notification settings
type UpdateNotificationSettingsRequest struct {
	TaskNotificationsEnabled bool `json:"task_notifications_enabled"`
	NotifyOnAssignment       bool `json:"notify_on_assignment"`
	NotifyOnTaskUpdate       bool `json:"notify_on_task_update"`
	NotifyOnComment          bool `json:"notify_on_comment"`
	NotifyOnDueDateChange    bool `json:"notify_on_due_date_change"`
	NotifyOnUnassignment     bool `json:"notify_on_unassignment"`
	NotifyOnChecklistItem    bool `json:"notify_on_checklist_item"`
}

// UpdateBoardNotificationSettingsRequest represents the request body for updating board-specific notification settings
type UpdateBoardNotificationSettingsRequest struct {
	NotificationsEnabled  bool `json:"notifications_enabled"`
	NotifyOnAssignment    bool `json:"notify_on_assignment"`
	NotifyOnTaskUpdate    bool `json:"notify_on_task_update"`
	NotifyOnComment       bool `json:"notify_on_comment"`
	NotifyOnDueDateChange bool `json:"notify_on_due_date_change"`
	NotifyOnUnassignment  bool `json:"notify_on_unassignment"`
	NotifyOnChecklistItem bool `json:"notify_on_checklist_item"`
}

// NotificationSettingsResponse includes both global and board-specific settings
type NotificationSettingsResponse struct {
	Global tables.NotificationSettings        `json:"global"`
	Boards []tables.BoardNotificationSettings `json:"boards"`
}

// GetNotificationSettings retrieves the notification settings for the authenticated user
func (h *NotificationSettingsHandler) GetNotificationSettings(w http.ResponseWriter, r *http.Request) {
	guildID := os.Getenv("GUILD_ID")
	userID := middleware.GetUserIDFromContext(r.Context())

	// Get or create global notification settings
	var globalSettings tables.NotificationSettings
	err := h.DB.QueryRow(`
		SELECT id, user_id, guild_id, task_notifications_enabled, notify_on_assignment, notify_on_task_update,
		       notify_on_comment, notify_on_due_date_change, notify_on_unassignment, notify_on_checklist_item,
		       created_at, updated_at
		FROM notification_settings
		WHERE user_id = $1 AND guild_id = $2
	`, userID, guildID).Scan(
		&globalSettings.ID,
		&globalSettings.UserID,
		&globalSettings.GuildID,
		&globalSettings.TaskNotificationsEnabled,
		&globalSettings.NotifyOnAssignment,
		&globalSettings.NotifyOnTaskUpdate,
		&globalSettings.NotifyOnComment,
		&globalSettings.NotifyOnDueDateChange,
		&globalSettings.NotifyOnUnassignment,
		&globalSettings.NotifyOnChecklistItem,
		&globalSettings.CreatedAt,
		&globalSettings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create default settings
		err = h.DB.QueryRow(`
			INSERT INTO notification_settings (user_id, guild_id, task_notifications_enabled, notify_on_assignment, 
			                                   notify_on_task_update, notify_on_comment, notify_on_due_date_change,
			                                   notify_on_unassignment, notify_on_checklist_item)
			VALUES ($1, $2, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE)
			RETURNING id, user_id, guild_id, task_notifications_enabled, notify_on_assignment, notify_on_task_update,
			          notify_on_comment, notify_on_due_date_change, notify_on_unassignment, notify_on_checklist_item,
			          created_at, updated_at
		`, userID, guildID).Scan(
			&globalSettings.ID,
			&globalSettings.UserID,
			&globalSettings.GuildID,
			&globalSettings.TaskNotificationsEnabled,
			&globalSettings.NotifyOnAssignment,
			&globalSettings.NotifyOnTaskUpdate,
			&globalSettings.NotifyOnComment,
			&globalSettings.NotifyOnDueDateChange,
			&globalSettings.NotifyOnUnassignment,
			&globalSettings.NotifyOnChecklistItem,
			&globalSettings.CreatedAt,
			&globalSettings.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to create notification settings", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Failed to get notification settings", http.StatusInternalServerError)
		return
	}

	// Get board-specific settings
	rows, err := h.DB.Query(`
		SELECT id, user_id, board_id, notifications_enabled, notify_on_assignment, notify_on_task_update,
		       notify_on_comment, notify_on_due_date_change, notify_on_unassignment, notify_on_checklist_item,
		       created_at, updated_at
		FROM board_notification_settings
		WHERE user_id = $1
	`, userID)
	if err != nil {
		http.Error(w, "Failed to get board notification settings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	boardSettings := []tables.BoardNotificationSettings{}
	for rows.Next() {
		var setting tables.BoardNotificationSettings
		err := rows.Scan(
			&setting.ID,
			&setting.UserID,
			&setting.BoardID,
			&setting.NotificationsEnabled,
			&setting.NotifyOnAssignment,
			&setting.NotifyOnTaskUpdate,
			&setting.NotifyOnComment,
			&setting.NotifyOnDueDateChange,
			&setting.NotifyOnUnassignment,
			&setting.NotifyOnChecklistItem,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to scan board notification settings", http.StatusInternalServerError)
			return
		}
		boardSettings = append(boardSettings, setting)
	}

	response := NotificationSettingsResponse{
		Global: globalSettings,
		Boards: boardSettings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateNotificationSettings updates the global notification settings for the authenticated user
func (h *NotificationSettingsHandler) UpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	guildID := os.Getenv("GUILD_ID")
	userID := middleware.GetUserIDFromContext(r.Context())

	var req UpdateNotificationSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var settings tables.NotificationSettings
	err := h.DB.QueryRow(`
		INSERT INTO notification_settings (user_id, guild_id, task_notifications_enabled, notify_on_assignment, 
		                                   notify_on_task_update, notify_on_comment, notify_on_due_date_change,
		                                   notify_on_unassignment, notify_on_checklist_item)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, guild_id)
		DO UPDATE SET
			task_notifications_enabled = EXCLUDED.task_notifications_enabled,
			notify_on_assignment = EXCLUDED.notify_on_assignment,
			notify_on_task_update = EXCLUDED.notify_on_task_update,
			notify_on_comment = EXCLUDED.notify_on_comment,
			notify_on_due_date_change = EXCLUDED.notify_on_due_date_change,
			notify_on_unassignment = EXCLUDED.notify_on_unassignment,
			notify_on_checklist_item = EXCLUDED.notify_on_checklist_item,
			updated_at = NOW()
		RETURNING id, user_id, guild_id, task_notifications_enabled, notify_on_assignment, notify_on_task_update,
		          notify_on_comment, notify_on_due_date_change, notify_on_unassignment, notify_on_checklist_item,
		          created_at, updated_at
	`, userID, guildID, req.TaskNotificationsEnabled, req.NotifyOnAssignment, req.NotifyOnTaskUpdate,
		req.NotifyOnComment, req.NotifyOnDueDateChange, req.NotifyOnUnassignment, req.NotifyOnChecklistItem).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.GuildID,
		&settings.TaskNotificationsEnabled,
		&settings.NotifyOnAssignment,
		&settings.NotifyOnTaskUpdate,
		&settings.NotifyOnComment,
		&settings.NotifyOnDueDateChange,
		&settings.NotifyOnUnassignment,
		&settings.NotifyOnChecklistItem,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "Failed to update notification settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// GetBoardNotificationSettings retrieves the notification settings for a specific board
func (h *NotificationSettingsHandler) GetBoardNotificationSettings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	boardIDStr := r.PathValue("boardId")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	var settings tables.BoardNotificationSettings
	err = h.DB.QueryRow(`
		SELECT id, user_id, board_id, notifications_enabled, notify_on_assignment, notify_on_task_update,
		       notify_on_comment, notify_on_due_date_change, notify_on_unassignment, notify_on_checklist_item,
		       created_at, updated_at
		FROM board_notification_settings
		WHERE user_id = $1 AND board_id = $2
	`, userID, boardID).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.BoardID,
		&settings.NotificationsEnabled,
		&settings.NotifyOnAssignment,
		&settings.NotifyOnTaskUpdate,
		&settings.NotifyOnComment,
		&settings.NotifyOnDueDateChange,
		&settings.NotifyOnUnassignment,
		&settings.NotifyOnChecklistItem,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// Create default settings
		err = h.DB.QueryRow(`
			INSERT INTO board_notification_settings (user_id, board_id, notifications_enabled, notify_on_assignment, 
			                                         notify_on_task_update, notify_on_comment, notify_on_due_date_change,
			                                         notify_on_unassignment, notify_on_checklist_item)
			VALUES ($1, $2, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE)
			RETURNING id, user_id, board_id, notifications_enabled, notify_on_assignment, notify_on_task_update,
			          notify_on_comment, notify_on_due_date_change, notify_on_unassignment, notify_on_checklist_item,
			          created_at, updated_at
		`, userID, boardID).Scan(
			&settings.ID,
			&settings.UserID,
			&settings.BoardID,
			&settings.NotificationsEnabled,
			&settings.NotifyOnAssignment,
			&settings.NotifyOnTaskUpdate,
			&settings.NotifyOnComment,
			&settings.NotifyOnDueDateChange,
			&settings.NotifyOnUnassignment,
			&settings.NotifyOnChecklistItem,
			&settings.CreatedAt,
			&settings.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to create board notification settings", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Failed to get board notification settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateBoardNotificationSettings updates the notification settings for a specific board
func (h *NotificationSettingsHandler) UpdateBoardNotificationSettings(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	boardIDStr := r.PathValue("boardId")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	var req UpdateBoardNotificationSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var settings tables.BoardNotificationSettings
	err = h.DB.QueryRow(`
		INSERT INTO board_notification_settings (user_id, board_id, notifications_enabled, notify_on_assignment, 
		                                         notify_on_task_update, notify_on_comment, notify_on_due_date_change,
		                                         notify_on_unassignment, notify_on_checklist_item)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, board_id)
		DO UPDATE SET
			notifications_enabled = EXCLUDED.notifications_enabled,
			notify_on_assignment = EXCLUDED.notify_on_assignment,
			notify_on_task_update = EXCLUDED.notify_on_task_update,
			notify_on_comment = EXCLUDED.notify_on_comment,
			notify_on_due_date_change = EXCLUDED.notify_on_due_date_change,
			notify_on_unassignment = EXCLUDED.notify_on_unassignment,
			notify_on_checklist_item = EXCLUDED.notify_on_checklist_item,
			updated_at = NOW()
		RETURNING id, user_id, board_id, notifications_enabled, notify_on_assignment, notify_on_task_update,
		          notify_on_comment, notify_on_due_date_change, notify_on_unassignment, notify_on_checklist_item,
		          created_at, updated_at
	`, userID, boardID, req.NotificationsEnabled, req.NotifyOnAssignment, req.NotifyOnTaskUpdate,
		req.NotifyOnComment, req.NotifyOnDueDateChange, req.NotifyOnUnassignment, req.NotifyOnChecklistItem).Scan(
		&settings.ID,
		&settings.UserID,
		&settings.BoardID,
		&settings.NotificationsEnabled,
		&settings.NotifyOnAssignment,
		&settings.NotifyOnTaskUpdate,
		&settings.NotifyOnComment,
		&settings.NotifyOnDueDateChange,
		&settings.NotifyOnUnassignment,
		&settings.NotifyOnChecklistItem,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "Failed to update board notification settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}
