package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"discord-bot-template/shared/database/tables"

	"github.com/gorilla/mux"
)

type TaskGroupsHandler struct {
	DB *sql.DB
}

type CreateTaskGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type UpdateTaskGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type TaskGroupPermissionRequest struct {
	RoleID     *string                `json:"role_id"`
	UserID     *string                `json:"user_id"`
	Permission tables.PermissionLevel `json:"permission"`
}

// GetTaskGroups returns all task groups for a guild
func (h *TaskGroupsHandler) GetTaskGroups(w http.ResponseWriter, r *http.Request) {
	guildID := os.Getenv("GUILD_ID")

	query := `
		SELECT id, guild_id, name, description, color, created_at, updated_at
		FROM task_groups
		WHERE guild_id = $1
		ORDER BY name
	`

	rows, err := h.DB.Query(query, guildID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []tables.TaskGroup
	for rows.Next() {
		var group tables.TaskGroup
		err := rows.Scan(&group.ID, &group.GuildID, &group.Name, &group.Description,
			&group.Color, &group.CreatedAt, &group.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		groups = append(groups, group)
	}

	if groups == nil {
		groups = []tables.TaskGroup{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// GetTaskGroup returns a single task group by ID
func (h *TaskGroupsHandler) GetTaskGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	guildID := os.Getenv("GUILD_ID")

	query := `
		SELECT id, guild_id, name, description, color, created_at, updated_at
		FROM task_groups
		WHERE id = $1 AND guild_id = $2
	`

	var group tables.TaskGroup
	err = h.DB.QueryRow(query, groupID, guildID).Scan(
		&group.ID, &group.GuildID, &group.Name, &group.Description,
		&group.Color, &group.CreatedAt, &group.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// CreateTaskGroup creates a new task group
func (h *TaskGroupsHandler) CreateTaskGroup(w http.ResponseWriter, r *http.Request) {
	guildID := os.Getenv("GUILD_ID")

	var req CreateTaskGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if req.Color == "" {
		req.Color = "#39d98a"
	}

	query := `
		INSERT INTO task_groups (guild_id, name, description, color, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, guild_id, name, description, color, created_at, updated_at
	`

	now := time.Now()
	var group tables.TaskGroup
	err := h.DB.QueryRow(query, guildID, req.Name, req.Description, req.Color, now, now).Scan(
		&group.ID, &group.GuildID, &group.Name, &group.Description,
		&group.Color, &group.CreatedAt, &group.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

// UpdateTaskGroup updates a task group
func (h *TaskGroupsHandler) UpdateTaskGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	guildID := os.Getenv("GUILD_ID")

	var req UpdateTaskGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		UPDATE task_groups
		SET name = $1, description = $2, color = $3, updated_at = $4
		WHERE id = $5 AND guild_id = $6
		RETURNING id, guild_id, name, description, color, created_at, updated_at
	`

	var group tables.TaskGroup
	err = h.DB.QueryRow(query, req.Name, req.Description, req.Color, time.Now(), groupID, guildID).Scan(
		&group.ID, &group.GuildID, &group.Name, &group.Description,
		&group.Color, &group.CreatedAt, &group.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(group)
}

// DeleteTaskGroup deletes a task group
func (h *TaskGroupsHandler) DeleteTaskGroup(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	guildID := os.Getenv("GUILD_ID")

	query := `DELETE FROM task_groups WHERE id = $1 AND guild_id = $2`
	result, err := h.DB.Exec(query, groupID, guildID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTaskGroupPermissions returns all permissions for a task group
func (h *TaskGroupsHandler) GetTaskGroupPermissions(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	query := `
		SELECT id, group_id, role_id, user_id, permission, created_at
		FROM task_group_permissions
		WHERE group_id = $1
		ORDER BY created_at
	`

	rows, err := h.DB.Query(query, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var permissions []tables.TaskGroupPermission
	for rows.Next() {
		var perm tables.TaskGroupPermission
		err := rows.Scan(&perm.ID, &perm.GroupID, &perm.RoleID, &perm.UserID,
			&perm.Permission, &perm.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		permissions = append(permissions, perm)
	}

	if permissions == nil {
		permissions = []tables.TaskGroupPermission{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(permissions)
}

// SetTaskGroupPermission creates or updates a task group permission
func (h *TaskGroupsHandler) SetTaskGroupPermission(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var req TaskGroupPermissionRequest
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
	checkQuery := `SELECT id FROM task_group_permissions WHERE group_id = $1 AND `
	if req.RoleID != nil {
		checkQuery += `role_id = $2`
		err = h.DB.QueryRow(checkQuery, groupID, *req.RoleID).Scan(&existingID)
	} else {
		checkQuery += `user_id = $2`
		err = h.DB.QueryRow(checkQuery, groupID, *req.UserID).Scan(&existingID)
	}

	var perm tables.TaskGroupPermission
	switch err {
	case sql.ErrNoRows:
		// Create new permission
		query := `
			INSERT INTO task_group_permissions (group_id, role_id, user_id, permission, created_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, group_id, role_id, user_id, permission, created_at
		`
		err = h.DB.QueryRow(query, groupID, req.RoleID, req.UserID, req.Permission, time.Now()).Scan(
			&perm.ID, &perm.GroupID, &perm.RoleID, &perm.UserID, &perm.Permission, &perm.CreatedAt,
		)
	case nil:
		// Update existing permission
		query := `
			UPDATE task_group_permissions
			SET permission = $1
			WHERE id = $2
			RETURNING id, group_id, role_id, user_id, permission, created_at
		`
		err = h.DB.QueryRow(query, req.Permission, existingID).Scan(
			&perm.ID, &perm.GroupID, &perm.RoleID, &perm.UserID, &perm.Permission, &perm.CreatedAt,
		)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(perm)
}

// DeleteTaskGroupPermission deletes a task group permission
func (h *TaskGroupsHandler) DeleteTaskGroupPermission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	permissionID, err := strconv.Atoi(vars["permissionId"])
	if err != nil {
		http.Error(w, "Invalid permission ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM task_group_permissions WHERE id = $1`
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
