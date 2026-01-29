package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"discord-bot-template/shared/database/tables"
)

type BoardLabelsHandler struct {
	DB *sql.DB
}

type CreateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type UpdateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// GetBoardLabels returns all labels for a board
func (h *BoardLabelsHandler) GetBoardLabels(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, board_id, name, color, created_at
		FROM board_labels
		WHERE board_id = $1
		ORDER BY name
	`

	rows, err := h.DB.Query(query, boardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var labels []tables.BoardLabel
	for rows.Next() {
		var label tables.BoardLabel
		err := rows.Scan(&label.ID, &label.BoardID, &label.Name, &label.Color, &label.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		labels = append(labels, label)
	}

	if labels == nil {
		labels = []tables.BoardLabel{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(labels)
}

// CreateBoardLabel creates a new label for a board
func (h *BoardLabelsHandler) CreateBoardLabel(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	var req CreateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if req.Color == "" {
		req.Color = "blue"
	}

	// Check if label with same name already exists
	var existingID int
	err = h.DB.QueryRow(`SELECT id FROM board_labels WHERE board_id = $1 AND name = $2`, boardID, req.Name).Scan(&existingID)
	if err == nil {
		http.Error(w, "Label with this name already exists", http.StatusConflict)
		return
	}

	query := `
		INSERT INTO board_labels (board_id, name, color, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, board_id, name, color, created_at
	`

	var label tables.BoardLabel
	err = h.DB.QueryRow(query, boardID, req.Name, req.Color, time.Now()).Scan(
		&label.ID, &label.BoardID, &label.Name, &label.Color, &label.CreatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(label)
}

// UpdateBoardLabel updates an existing label
func (h *BoardLabelsHandler) UpdateBoardLabel(w http.ResponseWriter, r *http.Request) {
	labelID, err := strconv.Atoi(r.PathValue("labelId"))
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	var req UpdateLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		UPDATE board_labels
		SET name = $1, color = $2
		WHERE id = $3
		RETURNING id, board_id, name, color, created_at
	`

	var label tables.BoardLabel
	err = h.DB.QueryRow(query, req.Name, req.Color, labelID).Scan(
		&label.ID, &label.BoardID, &label.Name, &label.Color, &label.CreatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Label not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(label)
}

// DeleteBoardLabel deletes a label from a board
func (h *BoardLabelsHandler) DeleteBoardLabel(w http.ResponseWriter, r *http.Request) {
	labelID, err := strconv.Atoi(r.PathValue("labelId"))
	if err != nil {
		http.Error(w, "Invalid label ID", http.StatusBadRequest)
		return
	}

	// Get the label name before deleting (for updating tasks)
	var labelName string
	var boardID int
	err = h.DB.QueryRow(`SELECT name, board_id FROM board_labels WHERE id = $1`, labelID).Scan(&labelName, &boardID)
	if err == sql.ErrNoRows {
		http.Error(w, "Label not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete the label
	query := `DELETE FROM board_labels WHERE id = $1`
	result, err := h.DB.Exec(query, labelID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Label not found", http.StatusNotFound)
		return
	}

	// Remove the label from all tasks on this board
	// Tags are stored as JSON array, so we need to update them
	updateQuery := `
		UPDATE tasks 
		SET tags = (
			SELECT COALESCE(
				jsonb_agg(elem)::text, 
				'[]'
			)
			FROM jsonb_array_elements_text(tags::jsonb) AS elem 
			WHERE elem != $1
		)
		WHERE board_id = $2 AND tags::jsonb ? $1
	`
	h.DB.Exec(updateQuery, labelName, boardID)

	w.WriteHeader(http.StatusNoContent)
}
