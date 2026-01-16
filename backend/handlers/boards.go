package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"discord-bot-template/backend/middleware"
	"discord-bot-template/shared/database/tables"
)

type BoardsHandler struct {
	DB *sql.DB
}

type CreateBoardRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type UpdateBoardRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Position    *int   `json:"position"`
}

type BoardPermissionRequest struct {
	RoleID    *string `json:"role_id"`
	UserID    *string `json:"user_id"`
	CanView   bool    `json:"can_view"`
	CanCreate bool    `json:"can_create"`
}

// GetBoards returns all boards for a guild
func (h *BoardsHandler) GetBoards(w http.ResponseWriter, r *http.Request) {
	guildID := os.Getenv("GUILD_ID")

	query := `
		SELECT id, guild_id, name, description, color, position, created_by, created_at, updated_at
		FROM boards
		WHERE guild_id = $1
		ORDER BY position, created_at
	`

	rows, err := h.DB.Query(query, guildID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var boards []tables.Board
	for rows.Next() {
		var board tables.Board
		err := rows.Scan(&board.ID, &board.GuildID, &board.Name, &board.Description,
			&board.Color, &board.Position, &board.CreatedBy, &board.CreatedAt, &board.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		boards = append(boards, board)
	}

	if boards == nil {
		boards = []tables.Board{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boards)
}

// GetBoard returns a single board by ID
func (h *BoardsHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	guildID := os.Getenv("GUILD_ID")

	query := `
		SELECT id, guild_id, name, description, color, position, created_by, created_at, updated_at
		FROM boards
		WHERE id = $1 AND guild_id = $2
	`

	var board tables.Board
	err = h.DB.QueryRow(query, boardID, guildID).Scan(
		&board.ID, &board.GuildID, &board.Name, &board.Description,
		&board.Color, &board.Position, &board.CreatedBy, &board.CreatedAt, &board.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Board not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(board)
}

// CreateBoard creates a new board
func (h *BoardsHandler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	guildID := os.Getenv("GUILD_ID")
	userID := middleware.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if req.Color == "" {
		req.Color = "#6aa6ff"
	}

	query := `
		INSERT INTO boards (guild_id, name, description, color, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, guild_id, name, description, color, position, created_by, created_at, updated_at
	`

	now := time.Now()
	var board tables.Board
	err := h.DB.QueryRow(query, guildID, req.Name, req.Description, req.Color, userID, now, now).Scan(
		&board.ID, &board.GuildID, &board.Name, &board.Description,
		&board.Color, &board.Position, &board.CreatedBy, &board.CreatedAt, &board.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(board)
}

// UpdateBoard updates a board
func (h *BoardsHandler) UpdateBoard(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	guildID := os.Getenv("GUILD_ID")

	var req UpdateBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		UPDATE boards
		SET name = $1, description = $2, color = $3, updated_at = $4
		WHERE id = $5 AND guild_id = $6
		RETURNING id, guild_id, name, description, color, position, created_by, created_at, updated_at
	`

	var board tables.Board
	err = h.DB.QueryRow(query, req.Name, req.Description, req.Color, time.Now(), boardID, guildID).Scan(
		&board.ID, &board.GuildID, &board.Name, &board.Description,
		&board.Color, &board.Position, &board.CreatedBy, &board.CreatedAt, &board.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Board not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(board)
}

// DeleteBoard deletes a board
func (h *BoardsHandler) DeleteBoard(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	guildID := os.Getenv("GUILD_ID")

	query := `DELETE FROM boards WHERE id = $1 AND guild_id = $2`
	result, err := h.DB.Exec(query, boardID, guildID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Board not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetBoardPermissions returns all permissions for a board
func (h *BoardsHandler) GetBoardPermissions(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, board_id, role_id, user_id, can_view, can_create, created_at
		FROM board_permissions
		WHERE board_id = $1
		ORDER BY created_at
	`

	rows, err := h.DB.Query(query, boardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var permissions []tables.BoardPermission
	for rows.Next() {
		var perm tables.BoardPermission
		err := rows.Scan(&perm.ID, &perm.BoardID, &perm.RoleID, &perm.UserID,
			&perm.CanView, &perm.CanCreate, &perm.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		permissions = append(permissions, perm)
	}

	if permissions == nil {
		permissions = []tables.BoardPermission{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

// SetBoardPermission creates or updates a board permission
func (h *BoardsHandler) SetBoardPermission(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	var req BoardPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if (req.RoleID == nil && req.UserID == nil) || (req.RoleID != nil && req.UserID != nil) {
		http.Error(w, "Either role_id or user_id must be provided, but not both", http.StatusBadRequest)
		return
	}

	// Check if permission already exists
	var existingID int
	checkQuery := `SELECT id FROM board_permissions WHERE board_id = $1 AND `
	if req.RoleID != nil {
		checkQuery += `role_id = $2`
		err = h.DB.QueryRow(checkQuery, boardID, *req.RoleID).Scan(&existingID)
	} else {
		checkQuery += `user_id = $2`
		err = h.DB.QueryRow(checkQuery, boardID, *req.UserID).Scan(&existingID)
	}

	var perm tables.BoardPermission
	switch err {
	case sql.ErrNoRows:
		// Create new permission
		query := `
			INSERT INTO board_permissions (board_id, role_id, user_id, can_view, can_create, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, board_id, role_id, user_id, can_view, can_create, created_at
		`
		err = h.DB.QueryRow(query, boardID, req.RoleID, req.UserID, req.CanView, req.CanCreate, time.Now()).Scan(
			&perm.ID, &perm.BoardID, &perm.RoleID, &perm.UserID, &perm.CanView, &perm.CanCreate, &perm.CreatedAt,
		)
	case nil:
		// Update existing permission
		query := `
			UPDATE board_permissions
			SET can_view = $1, can_create = $2
			WHERE id = $3
			RETURNING id, board_id, role_id, user_id, can_view, can_create, created_at
		`
		err = h.DB.QueryRow(query, req.CanView, req.CanCreate, existingID).Scan(
			&perm.ID, &perm.BoardID, &perm.RoleID, &perm.UserID, &perm.CanView, &perm.CanCreate, &perm.CreatedAt,
		)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(perm)
}

// DeleteBoardPermission deletes a board permission
func (h *BoardsHandler) DeleteBoardPermission(w http.ResponseWriter, r *http.Request) {
	permissionID, err := strconv.Atoi(r.PathValue("permissionId"))
	if err != nil {
		http.Error(w, "Invalid permission ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM board_permissions WHERE id = $1`
	result, err := h.DB.Exec(query, permissionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Permission not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
