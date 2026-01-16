package tables

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// TaskStatus represents the status column of a task
type TaskStatus string

const (
	TaskStatusToDo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusReview     TaskStatus = "review"
	TaskStatusDone       TaskStatus = "done"
)

// TagArray is a custom type for handling JSON array of tags
type TagArray []string

// Scan implements the sql.Scanner interface
func (t *TagArray) Scan(value interface{}) error {
	if value == nil {
		*t = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*t = []string{}
			return nil
		}
		bytes = []byte(str)
	}

	return json.Unmarshal(bytes, t)
}

// Value implements the driver.Valuer interface
func (t TagArray) Value() (driver.Value, error) {
	if t == nil {
		return "[]", nil
	}
	return json.Marshal(t)
}

// MarshalJSON implements json.Marshaler
func (t TagArray) MarshalJSON() ([]byte, error) {
	if t == nil {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(t))
}

// Task represents a task in a Kanban board
type Task struct {
	ID          int        `json:"id" db:"id"`
	BoardID     int        `json:"board_id" db:"board_id"`
	GroupID     *int       `json:"group_id,omitempty" db:"group_id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Status      TaskStatus `json:"status" db:"status"`
	Position    int        `json:"position" db:"position"`
	AssigneeID  *string    `json:"assignee_id,omitempty" db:"assignee_id"` // Discord user ID
	DueDate     *time.Time `json:"due_date,omitempty" db:"due_date"`
	Tags        TagArray   `json:"tags" db:"tags"` // JSON array of strings
	CreatedBy   string     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// TaskGroup represents a permission group for tasks
type TaskGroup struct {
	ID          int       `json:"id" db:"id"`
	GuildID     string    `json:"guild_id" db:"guild_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// PermissionLevel defines what level of access a user has to a task
type PermissionLevel string

const (
	PermissionNone        PermissionLevel = "none"
	PermissionExistence   PermissionLevel = "existence"    // See assignee and due date only
	PermissionReadTitle   PermissionLevel = "read_title"   // + task title
	PermissionReadContent PermissionLevel = "read_content" // + task details
	PermissionEdit        PermissionLevel = "edit"         // + edit details and change status
	PermissionDelete      PermissionLevel = "delete"       // + delete tasks
)

// TaskGroupPermission defines permissions for a Discord role or user in a group
type TaskGroupPermission struct {
	ID         int             `json:"id" db:"id"`
	GroupID    int             `json:"group_id" db:"group_id"`
	RoleID     *string         `json:"role_id,omitempty" db:"role_id"` // Discord role ID (null for user-specific)
	UserID     *string         `json:"user_id,omitempty" db:"user_id"` // Discord user ID (null for role-based)
	Permission PermissionLevel `json:"permission" db:"permission"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// CreateTasksTable creates the task_groups, tasks, and task_group_permissions tables
func CreateTasksTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS task_groups (
		id SERIAL PRIMARY KEY,
		guild_id VARCHAR(20) NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		color VARCHAR(7) DEFAULT '#6aa6ff',
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_task_groups_guild_id ON task_groups(guild_id);

	CREATE TABLE IF NOT EXISTS tasks (
		id SERIAL PRIMARY KEY,
		board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
		group_id INTEGER REFERENCES task_groups(id) ON DELETE SET NULL,
		title VARCHAR(255) NOT NULL,
		description TEXT,
		status VARCHAR(20) NOT NULL DEFAULT 'todo',
		position INTEGER DEFAULT 0,
		assignee_id VARCHAR(20),
		due_date TIMESTAMP,
		tags JSONB DEFAULT '[]'::jsonb,
		created_by VARCHAR(20) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_board_id ON tasks(board_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_group_id ON tasks(group_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks(assignee_id);

	CREATE TABLE IF NOT EXISTS task_group_permissions (
		id SERIAL PRIMARY KEY,
		group_id INTEGER NOT NULL REFERENCES task_groups(id) ON DELETE CASCADE,
		role_id VARCHAR(20),
		user_id VARCHAR(20),
		permission VARCHAR(20) NOT NULL DEFAULT 'none',
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		CONSTRAINT chk_task_group_permission_role_or_user CHECK (
			(role_id IS NOT NULL AND user_id IS NULL) OR
			(role_id IS NULL AND user_id IS NOT NULL)
		)
	);

	CREATE INDEX IF NOT EXISTS idx_task_group_permissions_group_id ON task_group_permissions(group_id);
	CREATE INDEX IF NOT EXISTS idx_task_group_permissions_role_id ON task_group_permissions(role_id);
	CREATE INDEX IF NOT EXISTS idx_task_group_permissions_user_id ON task_group_permissions(user_id);
	`

	_, err := db.Exec(query)
	return err
}
