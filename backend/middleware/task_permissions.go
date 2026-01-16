package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"discord-bot-template/shared/database/tables"
)

// TaskPermissionChecker provides methods to check task permissions
type TaskPermissionChecker struct {
	DB *sql.DB
}

// BoardPermissionChecker provides methods to check board permissions
type BoardPermissionChecker struct {
	DB *sql.DB
}

// GetUserTaskPermission returns the highest permission level a user has for a task
func (c *TaskPermissionChecker) GetUserTaskPermission(taskID int, userID string, userRoles []string) (tables.PermissionLevel, error) {
	// Get the task's group
	var groupID *int
	err := c.DB.QueryRow(`SELECT group_id FROM tasks WHERE id = $1`, taskID).Scan(&groupID)
	if err != nil {
		return tables.PermissionNone, err
	}

	// If no group, allow read_content by default (you can change this logic)
	if groupID == nil {
		return tables.PermissionReadContent, nil
	}

	return c.GetUserGroupPermission(*groupID, userID, userRoles)
}

// GetUserGroupPermission returns the highest permission level a user has for a group
func (c *TaskPermissionChecker) GetUserGroupPermission(groupID int, userID string, userRoles []string) (tables.PermissionLevel, error) {
	// Permission hierarchy: none < existence < read_title < read_content < edit < delete
	permissionOrder := map[tables.PermissionLevel]int{
		tables.PermissionNone:        0,
		tables.PermissionExistence:   1,
		tables.PermissionReadTitle:   2,
		tables.PermissionReadContent: 3,
		tables.PermissionEdit:        4,
		tables.PermissionDelete:      5,
	}

	highestPermission := tables.PermissionNone
	highestValue := 0

	// Check user-specific permission
	var userPerm tables.PermissionLevel
	err := c.DB.QueryRow(`
		SELECT permission FROM task_group_permissions
		WHERE group_id = $1 AND user_id = $2
	`, groupID, userID).Scan(&userPerm)

	if err == nil {
		if permissionOrder[userPerm] > highestValue {
			highestPermission = userPerm
			highestValue = permissionOrder[userPerm]
		}
	}

	// Check role-based permissions
	if len(userRoles) > 0 {
		placeholders := make([]string, len(userRoles))
		args := make([]interface{}, len(userRoles)+1)
		args[0] = groupID

		for i, role := range userRoles {
			placeholders[i] = "$" + strconv.Itoa(i+2)
			args[i+1] = role
		}

		query := `
			SELECT permission FROM task_group_permissions
			WHERE group_id = $1 AND role_id IN (` + strings.Join(placeholders, ",") + `)
		`

		rows, err := c.DB.Query(query, args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var rolePerm tables.PermissionLevel
				if err := rows.Scan(&rolePerm); err == nil {
					if permissionOrder[rolePerm] > highestValue {
						highestPermission = rolePerm
						highestValue = permissionOrder[rolePerm]
					}
				}
			}
		}
	}

	return highestPermission, nil
}

// RequireTaskPermission middleware ensures user has required permission level for a task
func (c *TaskPermissionChecker) RequireTaskPermission(required tables.PermissionLevel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			taskIDStr := r.PathValue("id")
			if taskIDStr == "" {
				taskIDStr = r.PathValue("taskId")
			}

			taskID, err := strconv.Atoi(taskIDStr)
			if err != nil {
				http.Error(w, "Invalid task ID", http.StatusBadRequest)
				return
			}

			userID := GetUserIDFromContext(r.Context())
			if userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get user roles from context (should be set by auth middleware)
			var userRoles []string
			if roles := r.Context().Value("user_roles"); roles != nil {
				userRoles = roles.([]string)
			}

			permission, err := c.GetUserTaskPermission(taskID, userID, userRoles)
			if err != nil {
				http.Error(w, "Error checking permissions", http.StatusInternalServerError)
				return
			}

			permissionOrder := map[tables.PermissionLevel]int{
				tables.PermissionNone:        0,
				tables.PermissionExistence:   1,
				tables.PermissionReadTitle:   2,
				tables.PermissionReadContent: 3,
				tables.PermissionEdit:        4,
				tables.PermissionDelete:      5,
			}

			if permissionOrder[permission] < permissionOrder[required] {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Add permission level to context for handlers to use
			ctx := context.WithValue(r.Context(), "task_permission", permission)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FilterTasksByPermission filters tasks based on user's permission level
func (c *TaskPermissionChecker) FilterTasksByPermission(tasks []tables.Task, userID string, userRoles []string) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)

	for _, task := range tasks {
		permission := tables.PermissionReadContent
		var err error

		if task.GroupID != nil {
			permission, err = c.GetUserGroupPermission(*task.GroupID, userID, userRoles)
			if err != nil {
				continue
			}
		}

		// Skip tasks with no permission
		if permission == tables.PermissionNone {
			continue
		}

		taskMap := make(map[string]interface{})
		taskMap["id"] = task.ID
		taskMap["board_id"] = task.BoardID
		taskMap["status"] = task.Status
		taskMap["position"] = task.Position
		taskMap["created_at"] = task.CreatedAt
		taskMap["permission"] = permission

		// Always include assignee and due date (existence level)
		if task.AssigneeID != nil {
			taskMap["assignee_id"] = *task.AssigneeID
		}
		if task.DueDate != nil {
			taskMap["due_date"] = *task.DueDate
		}

		// Add title if permission allows
		if permission >= tables.PermissionReadTitle {
			taskMap["title"] = task.Title
			taskMap["group_id"] = task.GroupID
		}

		// Add content if permission allows
		if permission >= tables.PermissionReadContent {
			taskMap["description"] = task.Description
			taskMap["tags"] = task.Tags
			taskMap["created_by"] = task.CreatedBy
			taskMap["updated_at"] = task.UpdatedAt
		}

		result = append(result, taskMap)
	}

	return result, nil
}

// GetUserBoardPermission checks if user has access to a board
func (b *BoardPermissionChecker) GetUserBoardPermission(boardID int, userID string, userRoles []string) (canView bool, canCreate bool, err error) {
	// Check if any permissions exist for this board
	var permissionCount int
	err = b.DB.QueryRow(`
		SELECT COUNT(*) FROM board_permissions WHERE board_id = $1
	`, boardID).Scan(&permissionCount)

	if err != nil {
		return false, false, err
	}

	// If no permissions are set for this board, allow all authenticated users
	if permissionCount == 0 {
		return true, true, nil
	}

	// Check user-specific permission
	var userCanView, userCanCreate sql.NullBool
	err = b.DB.QueryRow(`
		SELECT can_view, can_create FROM board_permissions
		WHERE board_id = $1 AND user_id = $2
	`, boardID, userID).Scan(&userCanView, &userCanCreate)

	if err == nil {
		return userCanView.Bool, userCanCreate.Bool, nil
	}

	// Check role-based permissions
	if len(userRoles) > 0 {
		placeholders := make([]string, len(userRoles))
		args := make([]interface{}, len(userRoles)+1)
		args[0] = boardID

		for i, role := range userRoles {
			placeholders[i] = "$" + strconv.Itoa(i+2)
			args[i+1] = role
		}

		query := `
			SELECT can_view, can_create FROM board_permissions
			WHERE board_id = $1 AND role_id IN (` + strings.Join(placeholders, ",") + `)
		`

		rows, err := b.DB.Query(query, args...)
		if err != nil {
			return false, false, err
		}
		defer rows.Close()

		canView = false
		canCreate = false
		for rows.Next() {
			var cv, cc bool
			if err := rows.Scan(&cv, &cc); err == nil {
				if cv {
					canView = true
				}
				if cc {
					canCreate = true
				}
			}
		}

		if canView || canCreate {
			return canView, canCreate, nil
		}
	}

	// Default: no access if permissions exist but user doesn't have any
	return false, false, nil
}

// RequireBoardView middleware ensures user can view the board
func (b *BoardPermissionChecker) RequireBoardView() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			boardIDStr := r.PathValue("id")
			if boardIDStr == "" {
				boardIDStr = r.PathValue("boardId")
			}

			boardID, err := strconv.Atoi(boardIDStr)
			if err != nil {
				http.Error(w, "Invalid board ID", http.StatusBadRequest)
				return
			}

			userID := GetUserIDFromContext(r.Context())
			if userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			var userRoles []string
			if roles := r.Context().Value("user_roles"); roles != nil {
				userRoles = roles.([]string)
			}

			canView, canCreate, err := b.GetUserBoardPermission(boardID, userID, userRoles)
			if err != nil || !canView {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), "can_create", canCreate)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireBoardCreate middleware ensures user can create tasks on the board
func (b *BoardPermissionChecker) RequireBoardCreate() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract board ID from request body for POST requests
			var reqBody struct {
				BoardID int `json:"board_id"`
			}

			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&reqBody); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			userID := GetUserIDFromContext(r.Context())
			if userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			var userRoles []string
			if roles := r.Context().Value("user_roles"); roles != nil {
				userRoles = roles.([]string)
			}

			_, canCreate, err := b.GetUserBoardPermission(reqBody.BoardID, userID, userRoles)
			if err != nil || !canCreate {
				http.Error(w, "You don't have permission to create tasks on this board", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
