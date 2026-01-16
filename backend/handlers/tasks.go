package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"discord-bot-template/backend/middleware"
	"discord-bot-template/shared/database/tables"
)

type TasksHandler struct {
	DB *sql.DB
}

type CreateTaskRequest struct {
	BoardID     int               `json:"board_id"`
	GroupID     *int              `json:"group_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      tables.TaskStatus `json:"status"`
	AssigneeID  *string           `json:"assignee_id"`
	DueDate     *time.Time        `json:"due_date"`
	Tags        []string          `json:"tags"`
}

type UpdateTaskRequest struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      tables.TaskStatus `json:"status"`
	Position    *int              `json:"position"`
	AssigneeID  *string           `json:"assignee_id"`
	DueDate     *time.Time        `json:"due_date"`
	Tags        []string          `json:"tags"`
}

type MoveTaskRequest struct {
	Status   tables.TaskStatus `json:"status"`
	Position int               `json:"position"`
}

// GetBoardTasks returns all tasks for a board
func (h *TasksHandler) GetBoardTasks(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.PathValue("boardId"))
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, board_id, group_id, title, description, status, position, 
		       assignee_id, due_date, tags, created_by, created_at, updated_at
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

	var tasks []tables.Task
	for rows.Next() {
		var task tables.Task
		var tagsJSON string
		err := rows.Scan(&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
			&task.Status, &task.Position, &task.AssigneeID, &task.DueDate, &tagsJSON,
			&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		task.Tags = tagsJSON
		tasks = append(tasks, task)
	}

	if tasks == nil {
		tasks = []tables.Task{}
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
		       assignee_id, due_date, tags, created_by, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	var task tables.Task
	var tagsJSON string
	err = h.DB.QueryRow(query, taskID).Scan(
		&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
		&task.Status, &task.Position, &task.AssigneeID, &task.DueDate, &tagsJSON,
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

	task.Tags = tagsJSON

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

	// Serialize tags to JSON
	tagsJSON := "[]"
	if req.Tags != nil {
		tagsBytes, err := json.Marshal(req.Tags)
		if err == nil {
			tagsJSON = string(tagsBytes)
		}
	}

	query := `
		INSERT INTO tasks (board_id, group_id, title, description, status, assignee_id, 
		                   due_date, tags, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, board_id, group_id, title, description, status, position,
		          assignee_id, due_date, tags, created_by, created_at, updated_at
	`

	now := time.Now()
	var task tables.Task
	var tagsResult string
	err := h.DB.QueryRow(query, req.BoardID, req.GroupID, req.Title, req.Description,
		req.Status, req.AssigneeID, req.DueDate, tagsJSON, userID, now, now).Scan(
		&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
		&task.Status, &task.Position, &task.AssigneeID, &task.DueDate, &tagsResult,
		&task.CreatedBy, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	task.Tags = tagsResult

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

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Serialize tags to JSON
	tagsJSON := "[]"
	if req.Tags != nil {
		tagsBytes, err := json.Marshal(req.Tags)
		if err == nil {
			tagsJSON = string(tagsBytes)
		}
	}

	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, assignee_id = $4,
		    due_date = $5, tags = $6, updated_at = $7
		WHERE id = $8
		RETURNING id, board_id, group_id, title, description, status, position,
		          assignee_id, due_date, tags, created_by, created_at, updated_at
	`

	var task tables.Task
	var tagsResult string
	err = h.DB.QueryRow(query, req.Title, req.Description, req.Status, req.AssigneeID,
		req.DueDate, tagsJSON, time.Now(), taskID).Scan(
		&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
		&task.Status, &task.Position, &task.AssigneeID, &task.DueDate, &tagsResult,
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

	task.Tags = tagsResult

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
	var tagsResult string
	err = h.DB.QueryRow(query, req.Status, req.Position, time.Now(), taskID).Scan(
		&task.ID, &task.BoardID, &task.GroupID, &task.Title, &task.Description,
		&task.Status, &task.Position, &task.AssigneeID, &task.DueDate, &tagsResult,
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

	task.Tags = tagsResult

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
