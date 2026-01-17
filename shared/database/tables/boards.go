package tables

import (
	"database/sql"
	"time"
)

// Board represents a Kanban board
type Board struct {
	ID          int       `json:"id" db:"id"`
	GuildID     string    `json:"guild_id" db:"guild_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	Position    int       `json:"position" db:"position"`
	CreatedBy   string    `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// BoardPermission defines who can see and create tasks on a board
type BoardPermission struct {
	ID              int       `json:"id" db:"id"`
	BoardID         int       `json:"board_id" db:"board_id"`
	RoleID          *string   `json:"role_id,omitempty" db:"role_id"`             // Discord role ID (null for user-specific)
	UserID          *string   `json:"user_id,omitempty" db:"user_id"`             // Discord user ID (null for role-based)
	CanViewBoard    bool      `json:"can_view_board" db:"can_view_board"`         // Can see the board exists
	CanViewTaskList bool      `json:"can_view_task_list" db:"can_view_task_list"` // Can see task titles in list
	CanViewTasks    bool      `json:"can_view_tasks" db:"can_view_tasks"`         // Can view full task details
	CanEditTasks    bool      `json:"can_edit_tasks" db:"can_edit_tasks"`         // Can create/edit/move tasks
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// CreateBoardsTable creates the boards and board_permissions tables
func CreateBoardsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS boards (
		id SERIAL PRIMARY KEY,
		guild_id VARCHAR(20) NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		color VARCHAR(7) DEFAULT '#6aa6ff',
		position INTEGER DEFAULT 0,
		created_by VARCHAR(20) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_boards_guild_id ON boards(guild_id);

	CREATE TABLE IF NOT EXISTS board_permissions (
		id SERIAL PRIMARY KEY,
		board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
		role_id VARCHAR(20),
		user_id VARCHAR(20),
		can_view BOOLEAN DEFAULT false,
		can_create BOOLEAN DEFAULT false,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		CONSTRAINT chk_board_permission_role_or_user CHECK (
			(role_id IS NOT NULL AND user_id IS NULL) OR
			(role_id IS NULL AND user_id IS NOT NULL)
		)
	);

	CREATE INDEX IF NOT EXISTS idx_board_permissions_board_id ON board_permissions(board_id);
	CREATE INDEX IF NOT EXISTS idx_board_permissions_role_id ON board_permissions(role_id);
	CREATE INDEX IF NOT EXISTS idx_board_permissions_user_id ON board_permissions(user_id);
	`

	_, err := db.Exec(query)
	return err
}
