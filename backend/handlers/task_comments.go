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

type TaskCommentsHandler struct {
	DB *sql.DB
}

type CreateCommentRequest struct {
	Text string `json:"text"`
}

type UpdateCommentRequest struct {
	Text string `json:"text"`
}

// GetComments returns all comments for a task
func (h *TaskCommentsHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	taskIDStr := r.PathValue("taskId")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, task_id, user_id, user_name, user_avatar, text, created_at, updated_at
		FROM task_comments
		WHERE task_id = $1
		ORDER BY created_at ASC
	`

	rows, err := h.DB.Query(query, taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []tables.TaskComment
	for rows.Next() {
		var comment tables.TaskComment
		err := rows.Scan(&comment.ID, &comment.TaskID, &comment.UserID, &comment.UserName,
			&comment.UserAvatar, &comment.Text, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		comments = append(comments, comment)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}

// CreateComment creates a new comment on a task
func (h *TaskCommentsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
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

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Get user info
	var userName, userAvatar string
	err = h.DB.QueryRow(`
		SELECT COALESCE(display_name, name), COALESCE(avatar_url, '')
		FROM users WHERE id = $1
	`, userID).Scan(&userName, &userAvatar)
	if err != nil {
		// Fallback if user not found
		userName = userID
		userAvatar = ""
	}

	now := time.Now()
	query := `
		INSERT INTO task_comments (task_id, user_id, user_name, user_avatar, text, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, task_id, user_id, user_name, user_avatar, text, created_at, updated_at
	`

	var comment tables.TaskComment
	err = h.DB.QueryRow(query, taskID, userID, userName, userAvatar, req.Text, now, now).Scan(
		&comment.ID, &comment.TaskID, &comment.UserID, &comment.UserName,
		&comment.UserAvatar, &comment.Text, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)
}

// UpdateComment updates a comment
func (h *TaskCommentsHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user owns this comment
	var ownerID string
	err = h.DB.QueryRow(`SELECT user_id FROM task_comments WHERE id = $1`, commentID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		http.Error(w, "You can only edit your own comments", http.StatusForbidden)
		return
	}

	query := `
		UPDATE task_comments
		SET text = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, task_id, user_id, user_name, user_avatar, text, created_at, updated_at
	`

	var comment tables.TaskComment
	err = h.DB.QueryRow(query, req.Text, time.Now(), commentID).Scan(
		&comment.ID, &comment.TaskID, &comment.UserID, &comment.UserName,
		&comment.UserAvatar, &comment.Text, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

// DeleteComment deletes a comment
func (h *TaskCommentsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user owns this comment
	var ownerID string
	err = h.DB.QueryRow(`SELECT user_id FROM task_comments WHERE id = $1`, commentID).Scan(&ownerID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		http.Error(w, "You can only delete your own comments", http.StatusForbidden)
		return
	}

	_, err = h.DB.Exec(`DELETE FROM task_comments WHERE id = $1`, commentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
