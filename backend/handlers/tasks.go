package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"discord-bot-template/backend/middleware"
	"discord-bot-template/shared/database/tables"
)

type TasksHandler struct {
	DB *sql.DB
}

// FlexibleDate handles both date-only ("2026-01-25") and datetime formats
type FlexibleDate struct {
	time.Time
}

func (fd *FlexibleDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		return nil
	}

	// Try date-only format first
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		fd.Time = t
		return nil
	}

	// Try RFC3339 datetime format
	t, err = time.Parse(time.RFC3339, s)
	if err == nil {
		fd.Time = t
		return nil
	}

	return err
}

type CreateTaskRequest struct {
	BoardID     int               `json:"board_id"`
	GroupID     *int              `json:"group_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      tables.TaskStatus `json:"status"`
	AssigneeID  *string           `json:"assignee_id"`
	StartDate   *FlexibleDate     `json:"start_date"`
	DueDate     *FlexibleDate     `json:"due_date"`
	Tags        []string          `json:"tags"`
}

type UpdateTaskRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      tables.TaskStatus `json:"status"`
	Position    *int              `json:"position"`
	AssigneeID  *string           `json:"assignee_id"`
	StartDate   *FlexibleDate     `json:"start_date"`
	DueDate     *FlexibleDate     `json:"due_date"`
	Tags        []string          `json:"tags"`
}

type MoveTaskRequest struct {
	Status   tables.TaskStatus `json:"status"`
	Position int               `json:"position"`
}

// UserBoardPermission holds the user's permission flags for a board
type UserBoardPermission struct {
	CanViewBoard    bool
	CanViewTaskList bool
	CanViewTasks    bool
	CanEditTasks    bool
	CanEditBoard    bool
}

// getUserBoardPermission returns the user's permission flags for a board
func (h *TasksHandler) getUserBoardPermission(boardID int, userID string, userRoles []string) UserBoardPermission {
	// Default: no permissions
	perm := UserBoardPermission{}

	// Check if user is the creator of the board (automatic full access)
	var createdBy string
	err := h.DB.QueryRow(`SELECT created_by FROM boards WHERE id = $1`, boardID).Scan(&createdBy)
	if err == nil && createdBy == userID {
		return UserBoardPermission{
			CanViewBoard:    true,
			CanViewTaskList: true,
			CanViewTasks:    true,
			CanEditTasks:    true,
			CanEditBoard:    true,
		}
	}

	// Check if user is admin (admin has all permissions)
	adminRoleIDs := strings.Split(os.Getenv("ADMIN_ROLE_IDS"), ",")
	for _, adminRole := range adminRoleIDs {
		adminRole = strings.TrimSpace(adminRole)
		if adminRole == "" {
			continue
		}
		for _, userRole := range userRoles {
			if userRole == adminRole {
				return UserBoardPermission{
					CanViewBoard:    true,
					CanViewTaskList: true,
					CanViewTasks:    true,
					CanEditTasks:    true,
					CanEditBoard:    true,
				}
			}
		}
	}

	// Check if any permissions exist for this board
	var permissionCount int
	err = h.DB.QueryRow(`SELECT COUNT(*) FROM board_permissions WHERE board_id = $1`, boardID).Scan(&permissionCount)
	if err != nil || permissionCount == 0 {
		// No permissions set = allow all authenticated users
		return UserBoardPermission{
			CanViewBoard:    true,
			CanViewTaskList: true,
			CanViewTasks:    true,
			CanEditTasks:    true,
			CanEditBoard:    false,
		}
	}

	// Check user-specific permission
	var canViewBoard, canViewTaskList, canViewTasks, canEditTasks, canEditBoard sql.NullBool
	err = h.DB.QueryRow(`
		SELECT can_view_board, can_view_task_list, can_view_tasks, can_edit_tasks, can_edit_board
		FROM board_permissions
		WHERE board_id = $1 AND user_id = $2
	`, boardID, userID).Scan(&canViewBoard, &canViewTaskList, &canViewTasks, &canEditTasks, &canEditBoard)

	if err == nil {
		if canViewBoard.Valid && canViewBoard.Bool {
			perm.CanViewBoard = true
		}
		if canViewTaskList.Valid && canViewTaskList.Bool {
			perm.CanViewTaskList = true
		}
		if canViewTasks.Valid && canViewTasks.Bool {
			perm.CanViewTasks = true
		}
		if canEditTasks.Valid && canEditTasks.Bool {
			perm.CanEditTasks = true
		}
		if canEditBoard.Valid && canEditBoard.Bool {
			perm.CanEditBoard = true
		}
	}

	// Check role-based permissions (merge with user permissions)
	if len(userRoles) > 0 {
		placeholders := make([]string, len(userRoles))
		args := make([]interface{}, len(userRoles)+1)
		args[0] = boardID

		for i, role := range userRoles {
			placeholders[i] = "$" + strconv.Itoa(i+2)
			args[i+1] = role
		}

		query := `
			SELECT can_view_board, can_view_task_list, can_view_tasks, can_edit_tasks, can_edit_board
			FROM board_permissions
			WHERE board_id = $1 AND role_id IN (` + strings.Join(placeholders, ",") + `)
		`

		rows, err := h.DB.Query(query, args...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var cvb, cvtl, cvt, cet, ceb sql.NullBool
				if err := rows.Scan(&cvb, &cvtl, &cvt, &cet, &ceb); err == nil {
					if cvb.Valid && cvb.Bool {
						perm.CanViewBoard = true
					}
					if cvtl.Valid && cvtl.Bool {
						perm.CanViewTaskList = true
					}
					if cvt.Valid && cvt.Bool {
						perm.CanViewTasks = true
					}
					if cet.Valid && cet.Bool {
						perm.CanEditTasks = true
					}
					if ceb.Valid && ceb.Bool {
						perm.CanEditBoard = true
					}
				}
			}
		}
	}

	return perm
}

// GetBoardTasks returns all tasks for a board
func (h *TasksHandler) GetBoardTasks(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.PathValue("boardId"))
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	// Get user ID and roles from context
	userID := middleware.GetUserIDFromContext(r.Context())
	var userRoles []string
	if roles := r.Context().Value(middleware.UserRolesKey); roles != nil {
		if roleSlice, ok := roles.([]string); ok {
			userRoles = roleSlice
		}
	}

	// Get user's board-level permissions
	boardPerm := h.getUserBoardPermission(boardID, userID, userRoles)

	// Determine permission level based on board permissions
	var permissionLevel string
	if boardPerm.CanEditTasks {
		permissionLevel = "edit"
	} else if boardPerm.CanViewTasks {
		permissionLevel = "read_content"
	} else if boardPerm.CanViewTaskList {
		permissionLevel = "read_title"
	} else if boardPerm.CanViewBoard {
		permissionLevel = "existence"
	} else {
		permissionLevel = "none"
	}

	query := `
		SELECT id, board_id, group_id, title, description, status, position, 
		       assignee_id, start_date, due_date, tags, created_by, created_at, updated_at
		FROM tasks
		WHERE board_id = $1
		ORDER BY status, position, created_at
	`

	rows, err := h.DB.Query(query, boardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var task tables.Task
		err := rows.Scan(&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
			&task.Status, &task.Position, &task.AssigneeID, &task.StartDate, &task.DueDate, &task.Tags,
			&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Build task map based on permission level
		taskMap := map[string]interface{}{
			"id":         task.ID,
			"board_id":   task.BoardID,
			"status":     task.Status,
			"position":   task.Position,
			"created_at": task.CreatedAt,
			"permission": permissionLevel,
		}

		// Always include assignee and due date (existence level)
		if task.AssigneeID != nil {
			taskMap["assignee_id"] = *task.AssigneeID
		}
		if task.DueDate != nil {
			taskMap["due_date"] = *task.DueDate
		}
		if task.StartDate != nil {
			taskMap["start_date"] = *task.StartDate
		}

		// Add title if permission allows (read_title or higher)
		if boardPerm.CanViewTaskList || boardPerm.CanViewTasks || boardPerm.CanEditTasks {
			taskMap["title"] = task.Title
			taskMap["group_id"] = task.GroupID
		}

		// Add content if permission allows (read_content or higher)
		if boardPerm.CanViewTasks || boardPerm.CanEditTasks {
			taskMap["description"] = task.Description
			taskMap["tags"] = task.Tags
			taskMap["created_by"] = task.CreatedBy
			taskMap["updated_at"] = task.UpdatedAt
		}

		tasks = append(tasks, taskMap)
	}

	if tasks == nil {
		tasks = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTask returns a single task by ID
func (h *TasksHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, board_id, group_id, title, description, status, position,
		       assignee_id, start_date, due_date, tags, created_by, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	var task tables.Task
	err = h.DB.QueryRow(query, taskID).Scan(
		&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
		&task.Status, &task.Position, &task.AssigneeID, &task.StartDate, &task.DueDate, &task.Tags,
		&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply permission-based filtering
	permissionValue := r.Context().Value("task_permission")
	if permissionValue != nil {
		permission, ok := permissionValue.(tables.PermissionLevel)
		if ok {
			// Filter data based on permission level
			permissionOrder := map[tables.PermissionLevel]int{
				tables.PermissionNone:        0,
				tables.PermissionExistence:   1,
				tables.PermissionReadTitle:   2,
				tables.PermissionReadContent: 3,
				tables.PermissionEdit:        4,
				tables.PermissionDelete:      5,
			}

			currentLevel := permissionOrder[permission]

			// If less than ReadContent, remove description and other sensitive data
			if currentLevel < permissionOrder[tables.PermissionReadContent] {
				task.Description = ""
				task.Tags = tables.TagArray([]string{})
			}

			// If less than ReadTitle, remove title
			if currentLevel < permissionOrder[tables.PermissionReadTitle] {
				task.Title = "[Restricted]"
			}

			// If only Existence, remove most data except assignee and dates
			if currentLevel < permissionOrder[tables.PermissionReadTitle] {
				task.Status = ""
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// CreateTask creates a new task
func (h *TasksHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = tables.TaskStatusToDo
	}

	// Convert tags to TagArray
	tags := tables.TagArray(req.Tags)
	if tags == nil {
		tags = tables.TagArray([]string{})
	}

	query := `
		INSERT INTO tasks (board_id, group_id, title, description, status, assignee_id, 
		                   start_date, due_date, tags, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, board_id, group_id, title, description, status, position,
		          assignee_id, start_date, due_date, tags, created_by, created_at, updated_at
	`

	now := time.Now()

	// Convert FlexibleDate to *time.Time
	var startDate *time.Time
	if req.StartDate != nil && !req.StartDate.Time.IsZero() {
		startDate = &req.StartDate.Time
	}

	var dueDate *time.Time
	if req.DueDate != nil && !req.DueDate.Time.IsZero() {
		dueDate = &req.DueDate.Time
	}

	var task tables.Task
	err := h.DB.QueryRow(query, req.BoardID, req.GroupID, req.Title, req.Description,
		req.Status, req.AssigneeID, startDate, dueDate, tags, userID, now, now).Scan(
		&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
		&task.Status, &task.Position, &task.AssigneeID, &task.StartDate, &task.DueDate, &task.Tags,
		&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send notification if task is assigned to someone
	if req.AssigneeID != nil && *req.AssigneeID != "" {
		go SendTaskNotification("task_assignment", task.ID, *req.AssigneeID, userID, "")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// UpdateTask updates a task
func (h *TasksHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get old task data to detect changes
	var oldTask tables.Task
	err = h.DB.QueryRow(`
		SELECT id, board_id, group_id, title, description, status, position,
		       assignee_id, start_date, due_date, tags, created_by, created_at, updated_at
		FROM tasks WHERE id = $1
	`, taskID).Scan(
		&oldTask.ID, &oldTask.BoardID, &oldTask.GroupID, &oldTask.Title, &oldTask.Description,
		&oldTask.Status, &oldTask.Position, &oldTask.AssigneeID, &oldTask.StartDate, &oldTask.DueDate,
		&oldTask.Tags, &oldTask.CreatedBy, &oldTask.CreatedAt, &oldTask.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert tags to TagArray
	tags := tables.TagArray(req.Tags)
	if tags == nil {
		tags = tables.TagArray([]string{})
	}

	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, assignee_id = $4,
		    start_date = $5, due_date = $6, tags = $7, updated_at = $8
		WHERE id = $9
		RETURNING id, board_id, group_id, title, description, status, position,
		          assignee_id, start_date, due_date, tags, created_by, created_at, updated_at
	`

	// Convert FlexibleDate to *time.Time
	var startDate *time.Time
	if req.StartDate != nil && !req.StartDate.Time.IsZero() {
		startDate = &req.StartDate.Time
	}

	var dueDate *time.Time
	if req.DueDate != nil && !req.DueDate.Time.IsZero() {
		dueDate = &req.DueDate.Time
	}

	var task tables.Task
	err = h.DB.QueryRow(query, req.Title, req.Description, req.Status, req.AssigneeID,
		startDate, dueDate, tags, time.Now(), taskID).Scan(
		&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
		&task.Status, &task.Position, &task.AssigneeID, &task.StartDate, &task.DueDate, &task.Tags,
		&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build change description for notification
	var changes []string
	if oldTask.Title != task.Title {
		changes = append(changes, "Titel geändert")
	}
	if oldTask.Description != task.Description {
		changes = append(changes, "Beschreibung geändert")
	}
	if oldTask.Status != task.Status {
		changes = append(changes, fmt.Sprintf("Status: %s → %s", oldTask.Status, task.Status))
	}

	// Check if assignee changed
	oldAssignee := ""
	newAssignee := ""
	if oldTask.AssigneeID != nil {
		oldAssignee = *oldTask.AssigneeID
	}
	if task.AssigneeID != nil {
		newAssignee = *task.AssigneeID
	}

	// Check if due date changed
	oldDueDateStr := ""
	newDueDateStr := ""
	if oldTask.DueDate != nil {
		oldDueDateStr = oldTask.DueDate.Format("02.01.2006 15:04")
	}
	if task.DueDate != nil {
		newDueDateStr = task.DueDate.Format("02.01.2006 15:04")
	}

	// Get all assigned users for notifications (old and new assignee)
	notifyUsers := make(map[string]bool)
	if oldAssignee != "" {
		notifyUsers[oldAssignee] = true
	}
	if newAssignee != "" {
		notifyUsers[newAssignee] = true
	}

	// Convert map to slice
	var userIDs []string
	for userID := range notifyUsers {
		userIDs = append(userIDs, userID)
	}

	// Handle assignee changes
	if oldAssignee != newAssignee {
		if newAssignee != "" && oldAssignee == "" {
			// New assignment
			go SendTaskNotification("task_assignment", task.ID, newAssignee, userID, "")
		} else if newAssignee != "" && oldAssignee != "" && newAssignee != oldAssignee {
			// Reassignment - notify old assignee about unassignment, new assignee about assignment
			go SendTaskNotification("task_unassignment", task.ID, oldAssignee, userID, "")
			go SendTaskNotification("task_assignment", task.ID, newAssignee, userID, "")
		} else if newAssignee == "" && oldAssignee != "" {
			// Unassignment
			go SendTaskNotification("task_unassignment", task.ID, oldAssignee, userID, "")
		}
	}

	// Handle due date changes
	if oldDueDateStr != newDueDateStr && len(userIDs) > 0 {
		changes = append(changes, "Fälligkeitsdatum geändert")
		go SendTaskDueDateChangeNotification(task.ID, userID, oldDueDateStr, newDueDateStr, userIDs)
	}

	// Send update notification for other changes if there are any and task is assigned
	if len(changes) > 0 && len(userIDs) > 0 {
		changeDesc := strings.Join(changes, ", ")
		go SendTaskNotification("task_update", task.ID, "", userID, changeDesc)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// MoveTask moves a task to a different status/position
func (h *TasksHandler) MoveTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var req MoveTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		UPDATE tasks
		SET status = $1, position = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, board_id, group_id, title, description, status, position,
		          assignee_id, due_date, tags, created_by, created_at, updated_at
	`

	var task tables.Task
	err = h.DB.QueryRow(query, req.Status, req.Position, time.Now(), taskID).Scan(
		&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
		&task.Status, &task.Position, &task.AssigneeID, &task.DueDate, &task.Tags,
		&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// DeleteTask deletes a task
func (h *TasksHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM tasks WHERE id = $1`
	result, err := h.DB.Exec(query, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
