package handlers

import (
	"database/sql"
	"discord-bot-template/backend/middleware"
	"discord-bot-template/shared/database/tables"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type TaskChecklistHandler struct {
	DB *sql.DB
}

type CreateChecklistItemRequest struct {
	Text string `json:"text"`
}

type UpdateChecklistItemRequest struct {
	Text        *string `json:"text,omitempty"`
	IsCompleted *bool   `json:"is_completed,omitempty"`
	Position    *int    `json:"position,omitempty"`
}

// GetChecklistItems returns all checklist items for a task
func (h *TaskChecklistHandler) GetChecklistItems(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.PathValue("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, task_id, text, is_completed, position, created_at, updated_at
		FROM task_checklist_items
		WHERE task_id = $1
		ORDER BY position ASC, created_at ASC
	`

	rows, err := h.DB.Query(query, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []tables.TaskChecklistItem
	for rows.Next() {
		var item tables.TaskChecklistItem
		err := rows.Scan(&item.ID, &item.TaskID, &item.Text, &item.IsCompleted,
			&item.Position, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// CreateChecklistItem creates a new checklist item
func (h *TaskChecklistHandler) CreateChecklistItem(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.PathValue("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateChecklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Get the highest position for this task
	var maxPosition int
	err = h.DB.QueryRow(`SELECT COALESCE(MAX(position), -1) FROM task_checklist_items WHERE task_id = $1`, taskID).Scan(&maxPosition)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	query := `
		INSERT INTO task_checklist_items (task_id, text, is_completed, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, task_id, text, is_completed, position, created_at, updated_at
	`

	var item tables.TaskChecklistItem
	err = h.DB.QueryRow(query, taskID, req.Text, false, maxPosition+1, now, now).Scan(
		&item.ID, &item.TaskID, &item.Text, &item.IsCompleted,
		&item.Position, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get task assignee(s) to notify
	var assigneeID *string
	err = h.DB.QueryRow(`SELECT assignee_id FROM tasks WHERE id = $1`, taskID).Scan(&assigneeID)
	if err == nil && assigneeID != nil && *assigneeID != "" && *assigneeID != userID {
		// Notify assignee about new checklist item (don't notify if creator is the assignee)
		go SendTaskChecklistItemNotification(taskID, userID, req.Text, []string{*assigneeID})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// UpdateChecklistItem updates a checklist item
func (h *TaskChecklistHandler) UpdateChecklistItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := r.PathValue("id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var req UpdateChecklistItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Build dynamic update query
	query := `UPDATE task_checklist_items SET updated_at = $1`
	args := []interface{}{time.Now()}
	argCount := 2

	if req.Text != nil {
		query += `, text = $` + strconv.Itoa(argCount)
		args = append(args, *req.Text)
		argCount++
	}

	if req.IsCompleted != nil {
		query += `, is_completed = $` + strconv.Itoa(argCount)
		args = append(args, *req.IsCompleted)
		argCount++
	}

	if req.Position != nil {
		query += `, position = $` + strconv.Itoa(argCount)
		args = append(args, *req.Position)
		argCount++
	}

	query += ` WHERE id = $` + strconv.Itoa(argCount)
	args = append(args, itemID)
	query += ` RETURNING id, task_id, text, is_completed, position, created_at, updated_at`

	var item tables.TaskChecklistItem
	err = h.DB.QueryRow(query, args...).Scan(
		&item.ID, &item.TaskID, &item.Text, &item.IsCompleted,
		&item.Position, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// ToggleChecklistItem toggles the completion status of a checklist item
func (h *TaskChecklistHandler) ToggleChecklistItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := r.PathValue("id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE task_checklist_items
		SET is_completed = NOT is_completed, updated_at = $1
		WHERE id = $2
		RETURNING id, task_id, text, is_completed, position, created_at, updated_at
	`

	var item tables.TaskChecklistItem
	err = h.DB.QueryRow(query, time.Now(), itemID).Scan(
		&item.ID, &item.TaskID, &item.Text, &item.IsCompleted,
		&item.Position, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// DeleteChecklistItem deletes a checklist item
func (h *TaskChecklistHandler) DeleteChecklistItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := r.PathValue("id")
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	_, err = h.DB.Exec(`DELETE FROM task_checklist_items WHERE id = $1`, itemID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
