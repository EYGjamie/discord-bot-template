package tables

import (
	"database/sql"
	"time"
)

// NotificationSettings represents user notification preferences
type NotificationSettings struct {
	ID                       int       `json:"id" db:"id"`
	UserID                   string    `json:"user_id" db:"user_id"`                                       // Discord user ID
	GuildID                  string    `json:"guild_id" db:"guild_id"`                                     // Discord guild ID
	TaskNotificationsEnabled bool      `json:"task_notifications_enabled" db:"task_notifications_enabled"` // Master switch for all task notifications
	NotifyOnAssignment       bool      `json:"notify_on_assignment" db:"notify_on_assignment"`             // Notify when assigned to a task
	NotifyOnTaskUpdate       bool      `json:"notify_on_task_update" db:"notify_on_task_update"`           // Notify when someone edits assigned task
	NotifyOnComment          bool      `json:"notify_on_comment" db:"notify_on_comment"`                   // Notify when someone comments on assigned task
	NotifyOnDueDateChange    bool      `json:"notify_on_due_date_change" db:"notify_on_due_date_change"`   // Notify when due date changes
	NotifyOnUnassignment     bool      `json:"notify_on_unassignment" db:"notify_on_unassignment"`         // Notify when unassigned from a task
	NotifyOnChecklistItem    bool      `json:"notify_on_checklist_item" db:"notify_on_checklist_item"`     // Notify when checklist item is added
	CreatedAt                time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time `json:"updated_at" db:"updated_at"`
}

// BoardNotificationSettings represents user notification preferences per board
type BoardNotificationSettings struct {
	ID                   int       `json:"id" db:"id"`
	UserID               string    `json:"user_id" db:"user_id"`                             // Discord user ID
	BoardID              int       `json:"board_id" db:"board_id"`                           // Board ID
	NotificationsEnabled bool      `json:"notifications_enabled" db:"notifications_enabled"` // Board-specific notifications on/off
	NotifyOnAssignment   bool      `json:"notify_on_assignment" db:"notify_on_assignment"`   // Override global setting for this board
	NotifyOnTaskUpdate   bool      `json:"notify_on_task_update" db:"notify_on_task_update"` // Override global setting for this board
	NotifyOnComment      bool      `json:"notify_on_comment" db:"notify_on_comment"`         // Override global setting for this board
	NotifyOnDueDateChange bool     `json:"notify_on_due_date_change" db:"notify_on_due_date_change"` // Override global setting for this board
	NotifyOnUnassignment bool      `json:"notify_on_unassignment" db:"notify_on_unassignment"` // Override global setting for this board
	NotifyOnChecklistItem bool     `json:"notify_on_checklist_item" db:"notify_on_checklist_item"` // Override global setting for this board
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// CreateNotificationSettingsTable creates the notification settings tables
func CreateNotificationSettingsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS notification_settings (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		guild_id VARCHAR(20) NOT NULL,
		task_notifications_enabled BOOLEAN DEFAULT TRUE,
		notify_on_assignment BOOLEAN DEFAULT TRUE,
		notify_on_task_update BOOLEAN DEFAULT TRUE,
		notify_on_comment BOOLEAN DEFAULT TRUE,
		notify_on_due_date_change BOOLEAN DEFAULT TRUE,
		notify_on_unassignment BOOLEAN DEFAULT TRUE,
		notify_on_checklist_item BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE(user_id, guild_id)
	);

	CREATE INDEX IF NOT EXISTS idx_notification_settings_user_guild ON notification_settings(user_id, guild_id);

	CREATE TABLE IF NOT EXISTS board_notification_settings (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(20) NOT NULL,
		board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
		notifications_enabled BOOLEAN DEFAULT TRUE,
		notify_on_assignment BOOLEAN DEFAULT TRUE,
		notify_on_task_update BOOLEAN DEFAULT TRUE,
		notify_on_comment BOOLEAN DEFAULT TRUE,
		notify_on_due_date_change BOOLEAN DEFAULT TRUE,
		notify_on_unassignment BOOLEAN DEFAULT TRUE,
		notify_on_checklist_item BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE(user_id, board_id)
	);

	CREATE INDEX IF NOT EXISTS idx_board_notification_settings_user_board ON board_notification_settings(user_id, board_id);
	CREATE INDEX IF NOT EXISTS idx_board_notification_settings_board ON board_notification_settings(board_id);
	`

	_, err := db.Exec(query)
	return err
}
