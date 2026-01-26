package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	RoleID          *string `json:"role_id"`
	UserID          *string `json:"user_id"`
	CanViewBoard    bool    `json:"can_view_board"`
	CanViewTaskList bool    `json:"can_view_task_list"`
	CanViewTasks    bool    `json:"can_view_tasks"`
	CanEditTasks    bool    `json:"can_edit_tasks"`
	CanEditBoard    bool    `json:"can_edit_board"`
}

// GetBoards returns all boards for a guild (filtered by user permissions)
func (h *BoardsHandler) GetBoards(w http.ResponseWriter, r *http.Request) {
	guildID := os.Getenv("GUILD_ID")
	userID := middleware.GetUserIDFromContext(r.Context())

	// Get user roles from context
	var userRoles []string
	if roles := r.Context().Value(middleware.UserRolesKey); roles != nil {
		userRoles = roles.([]string)
	}

	// Check if user is admin
	adminRoleIDs := strings.Split(os.Getenv("ADMIN_ROLE_IDS"), ",")
	isAdmin := false
	for _, adminRole := range adminRoleIDs {
		adminRole = strings.TrimSpace(adminRole)
		for _, userRole := range userRoles {
			if userRole == adminRole {
				isAdmin = true
				break
			}
		}
	}

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

		// If user is admin, include all boards
		if isAdmin {
			boards = append(boards, board)
			continue
		}

		// Check if user has permission to view this board
		canView, _, err := (&middleware.BoardPermissionChecker{DB: h.DB}).GetUserBoardPermission(board.ID, userID, userRoles)
		if err == nil && canView {
			boards = append(boards, board)
		}
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

	// Automatically grant full permissions to the creator (visible on Settings page)
	permissionQuery := `
		INSERT INTO board_permissions (board_id, user_id, can_view_board, can_view_task_list, can_view_tasks, can_edit_tasks, can_edit_board, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = h.DB.Exec(permissionQuery, board.ID, userID, true, true, true, true, true, now)
	if err != nil {
		// Log error but don't fail the request
		// The creator will still have permissions via the creator check
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
		SELECT 
			bp.id, bp.board_id, bp.role_id, bp.user_id, 
			bp.can_view_board, bp.can_view_task_list, bp.can_view_tasks, bp.can_edit_tasks, bp.can_edit_board,
			bp.created_at,
			r.name as role_name,
			u.name as user_name,
			COALESCE(u.display_name, u.name) as user_display_name
		FROM board_permissions bp
		LEFT JOIN roles r ON bp.role_id = r.id
		LEFT JOIN users u ON bp.user_id = u.id
		WHERE bp.board_id = $1
		ORDER BY bp.created_at
	`

	rows, err := h.DB.Query(query, boardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type PermissionResponse struct {
		ID              int       `json:"id"`
		BoardID         int       `json:"board_id"`
		RoleID          *string   `json:"role_id"`
		UserID          *string   `json:"user_id"`
		RoleName        *string   `json:"role_name"`
		UserName        *string   `json:"user_name"`
		UserDisplayName *string   `json:"user_display_name"`
		CanViewBoard    bool      `json:"can_view_board"`
		CanViewTaskList bool      `json:"can_view_task_list"`
		CanViewTasks    bool      `json:"can_view_tasks"`
		CanEditTasks    bool      `json:"can_edit_tasks"`
		CanEditBoard    bool      `json:"can_edit_board"`
		CreatedAt       time.Time `json:"created_at"`
	}

	var permissions []PermissionResponse
	for rows.Next() {
		var perm PermissionResponse
		err := rows.Scan(&perm.ID, &perm.BoardID, &perm.RoleID, &perm.UserID,
			&perm.CanViewBoard, &perm.CanViewTaskList, &perm.CanViewTasks, &perm.CanEditTasks, &perm.CanEditBoard, &perm.CreatedAt,
			&perm.RoleName, &perm.UserName, &perm.UserDisplayName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		permissions = append(permissions, perm)
	}

	if permissions == nil {
		permissions = []PermissionResponse{}
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
			INSERT INTO board_permissions (board_id, role_id, user_id, can_view_board, can_view_task_list, can_view_tasks, can_edit_tasks, can_edit_board, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, board_id, role_id, user_id, can_view_board, can_view_task_list, can_view_tasks, can_edit_tasks, can_edit_board, created_at
		`
		err = h.DB.QueryRow(query, boardID, req.RoleID, req.UserID, req.CanViewBoard, req.CanViewTaskList, req.CanViewTasks, req.CanEditTasks, req.CanEditBoard, time.Now()).Scan(
			&perm.ID, &perm.BoardID, &perm.RoleID, &perm.UserID, &perm.CanViewBoard, &perm.CanViewTaskList, &perm.CanViewTasks, &perm.CanEditTasks, &perm.CanEditBoard, &perm.CreatedAt,
		)
	case nil:
		// Update existing permission
		query := `
			UPDATE board_permissions
			SET can_view_board = $1, can_view_task_list = $2, can_view_tasks = $3, can_edit_tasks = $4, can_edit_board = $5
			WHERE id = $6
			RETURNING id, board_id, role_id, user_id, can_view_board, can_view_task_list, can_view_tasks, can_edit_tasks, can_edit_board, created_at
		`
		err = h.DB.QueryRow(query, req.CanViewBoard, req.CanViewTaskList, req.CanViewTasks, req.CanEditTasks, req.CanEditBoard, existingID).Scan(
			&perm.ID, &perm.BoardID, &perm.RoleID, &perm.UserID, &perm.CanViewBoard, &perm.CanViewTaskList, &perm.CanViewTasks, &perm.CanEditTasks, &perm.CanEditBoard, &perm.CreatedAt,
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

// UpdateBoardPermission updates an existing board permission
func (h *BoardsHandler) UpdateBoardPermission(w http.ResponseWriter, r *http.Request) {
	permissionID, err := strconv.Atoi(r.PathValue("permissionId"))
	if err != nil {
		http.Error(w, "Invalid permission ID", http.StatusBadRequest)
		return
	}

	var req BoardPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		UPDATE board_permissions
		SET can_view_board = $1, can_view_task_list = $2, can_view_tasks = $3, can_edit_tasks = $4, can_edit_board = $5
		WHERE id = $6
		RETURNING id
	`

	var id int
	err = h.DB.QueryRow(query, req.CanViewBoard, req.CanViewTaskList, req.CanViewTasks, req.CanEditTasks, req.CanEditBoard, permissionID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Permission not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "Permission updated successfully",
	})
}
