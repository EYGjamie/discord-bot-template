package handlers

import (
	"database/sql"
	"discord-bot-template/shared/database/tables"
	"encoding/json"
	"net/http"
)

// GetDiscordRolesAndMembers returns all roles and members for dropdowns
func GetDiscordRolesAndMembers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get roles
		rolesQuery := `SELECT id, name, color, position FROM roles ORDER BY position DESC, name`
		rolesRows, err := db.Query(rolesQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rolesRows.Close()

		var roles []tables.Role
		for rolesRows.Next() {
			var role tables.Role
			err := rolesRows.Scan(&role.ID, &role.Name, &role.Color, &role.Position)
			if err != nil {
				continue
			}
			roles = append(roles, role)
		}

		// Get members with basic info
		membersQuery := `
			SELECT id, name, COALESCE(display_name, name) as display_name, avatar 
			FROM users 
			WHERE bot = false
			ORDER BY display_name
		`
		membersRows, err := db.Query(membersQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer membersRows.Close()

		type MemberInfo struct {
			UserID      string  `json:"user_id"`
			Username    string  `json:"username"`
			DisplayName string  `json:"display_name"`
			Avatar      *string `json:"avatar"`
		}

		var members []MemberInfo
		for membersRows.Next() {
			var member MemberInfo
			err := membersRows.Scan(&member.UserID, &member.Username, &member.DisplayName, &member.Avatar)
			if err != nil {
				continue
			}
			members = append(members, member)
		}

		if roles == nil {
			roles = []tables.Role{}
		}
		if members == nil {
			members = []MemberInfo{}
		}

		response := map[string]interface{}{
			"roles":   roles,
			"members": members,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// SearchMembers searches for members by username or nickname
func SearchMembers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			json.NewEncoder(w).Encode([]interface{}{})
			return
		}

		searchQuery := `
			SELECT id, name, COALESCE(display_name, name) as display_name, avatar 
			FROM users 
			WHERE bot = false AND (
				LOWER(name) LIKE LOWER($1) OR 
				LOWER(COALESCE(display_name, '')) LIKE LOWER($1) OR
				id = $2
			)
			ORDER BY display_name
			LIMIT 10
		`

		rows, err := db.Query(searchQuery, "%"+query+"%", query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type MemberInfo struct {
			UserID      string  `json:"user_id"`
			Username    string  `json:"username"`
			DisplayName string  `json:"display_name"`
			Avatar      *string `json:"avatar"`
		}

		var members []MemberInfo
		for rows.Next() {
			var member MemberInfo
			err := rows.Scan(&member.UserID, &member.Username, &member.DisplayName, &member.Avatar)
			if err != nil {
				continue
			}
			members = append(members, member)
		}

		if members == nil {
			members = []MemberInfo{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(members)
	}
}
