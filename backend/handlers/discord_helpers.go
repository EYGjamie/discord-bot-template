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
			SELECT user_id, username, discriminator, avatar, nickname 
			FROM members 
			ORDER BY COALESCE(nickname, username)
		`
		membersRows, err := db.Query(membersQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer membersRows.Close()

		type MemberInfo struct {
			UserID        string  `json:"user_id"`
			Username      string  `json:"username"`
			Discriminator string  `json:"discriminator"`
			Avatar        *string `json:"avatar"`
			Nickname      *string `json:"nickname"`
			DisplayName   string  `json:"display_name"`
		}

		var members []MemberInfo
		for membersRows.Next() {
			var member MemberInfo
			err := membersRows.Scan(&member.UserID, &member.Username, &member.Discriminator, &member.Avatar, &member.Nickname)
			if err != nil {
				continue
			}
			// Set display name
			if member.Nickname != nil && *member.Nickname != "" {
				member.DisplayName = *member.Nickname
			} else {
				member.DisplayName = member.Username
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
			SELECT user_id, username, discriminator, avatar, nickname 
			FROM members 
			WHERE LOWER(username) LIKE LOWER($1) OR LOWER(COALESCE(nickname, '')) LIKE LOWER($1)
			ORDER BY COALESCE(nickname, username)
			LIMIT 10
		`

		rows, err := db.Query(searchQuery, "%"+query+"%")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type MemberInfo struct {
			UserID        string  `json:"user_id"`
			Username      string  `json:"username"`
			Discriminator string  `json:"discriminator"`
			Avatar        *string `json:"avatar"`
			Nickname      *string `json:"nickname"`
			DisplayName   string  `json:"display_name"`
		}

		var members []MemberInfo
		for rows.Next() {
			var member MemberInfo
			err := rows.Scan(&member.UserID, &member.Username, &member.Discriminator, &member.Avatar, &member.Nickname)
			if err != nil {
				continue
			}
			// Set display name
			if member.Nickname != nil && *member.Nickname != "" {
				member.DisplayName = *member.Nickname
			} else {
				member.DisplayName = member.Username
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
