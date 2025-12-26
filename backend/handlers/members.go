package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Member struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	GlobalName    string  `json:"global_name"`
	DisplayName   string  `json:"display_name"`
	Bot           bool    `json:"bot"`
	Avatar        string  `json:"avatar"`
	AvatarURL     string  `json:"avatar_url"`
	Mention       string  `json:"mention"`
	CreatedAt     string  `json:"created_at"`
	Nick          string  `json:"nick"`
	JoinedAt      string  `json:"joined_at"`
	TopRole       string  `json:"top_role"`
	TopRoleName   *string `json:"top_role_name,omitempty"`
	TopRoleColor  *string `json:"top_role_color,omitempty"`
	TimedOutUntil *string `json:"timed_out_until"`
	PremiumSince  *string `json:"premium_since"`
	UpdatedAt     string  `json:"updated_at"`
}

type MemberStats struct {
	TotalMessages   int             `json:"total_messages"`
	TotalVoiceTime  int             `json:"total_voice_time"`
	TopTextChannel  *TopChannel     `json:"top_text_channel,omitempty"`
	TopVoiceChannel *TopChannel     `json:"top_voice_channel,omitempty"`
	MutedDuration   int             `json:"muted_duration"`
	DeafenDuration  int             `json:"deafen_duration"`
	StreamDuration  int             `json:"stream_duration"`
	JoinCount       int             `json:"join_count"`
	TotalJoins      int             `json:"total_joins"`
	TotalLeaves     int             `json:"total_leaves"`
	TotalWarns      int             `json:"total_warns"`
	Roles           []Role          `json:"roles"`
	Warns           []ModerationLog `json:"warns"`
	Notes           []ModerationLog `json:"notes"`
}

type TopChannel struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MessageCount *int   `json:"message_count,omitempty"`
	Duration     *int   `json:"duration,omitempty"`
}

type Role struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type ModerationLog struct {
	ID            int64  `json:"id"`
	ModeratorID   string `json:"moderator_id"`
	ModeratorName string `json:"moderator_name"`
	Reason        string `json:"reason"`
	CreatedAt     string `json:"created_at"`
}

type MembersResponse struct {
	Members    []Member `json:"members"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	PerPage    int      `json:"per_page"`
	TotalPages int      `json:"total_pages"`
}

func GetMembers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		query := r.URL.Query()
		page, _ := strconv.Atoi(query.Get("page"))
		if page < 1 {
			page = 1
		}
		perPage, _ := strconv.Atoi(query.Get("per_page"))
		if perPage < 1 || perPage > 100 {
			perPage = 25
		}
		search := query.Get("search")
		roleFilter := query.Get("role")

		// Build SQL query
		baseQuery := `
			SELECT u.id, u.name, u.global_name, u.display_name, u.bot, u.avatar, u.avatar_url,
			       u.mention, u.created_at, u.nick, u.joined_at, u.top_role,
			       r.name as top_role_name, 
			       CASE WHEN r.color > 0 THEN '#' || LPAD(TO_HEX(r.color), 6, '0') ELSE NULL END as top_role_color,
			       u.timed_out_until, u.premium_since, u.updated_at
			FROM users u
			LEFT JOIN roles r ON u.top_role = r.id
			WHERE 1=1
		`

		countQuery := `
			SELECT COUNT(*)
			FROM users u
			LEFT JOIN roles r ON u.top_role = r.id
			WHERE 1=1
		`

		args := []interface{}{}
		argIndex := 1

		// Add search filter
		if search != "" {
			searchCondition := ` AND (
				LOWER(u.name) LIKE LOWER($` + strconv.Itoa(argIndex) + `) OR
				LOWER(u.display_name) LIKE LOWER($` + strconv.Itoa(argIndex) + `) OR
				LOWER(u.nick) LIKE LOWER($` + strconv.Itoa(argIndex) + `)
			)`
			baseQuery += searchCondition
			countQuery += searchCondition
			args = append(args, "%"+search+"%")
			argIndex++
		}

		// Add role filter
		if roleFilter != "" {
			roleCondition := ` AND r.name = $` + strconv.Itoa(argIndex)
			baseQuery += roleCondition
			countQuery += roleCondition
			args = append(args, roleFilter)
			argIndex++
		}

		// Get total count
		var total int
		err := db.QueryRow(countQuery, args...).Scan(&total)
		if err != nil {
			log.Printf("Error counting members: %v", err)
			http.Error(w, "Failed to count members", http.StatusInternalServerError)
			return
		}

		// Add pagination
		offset := (page - 1) * perPage
		baseQuery += ` ORDER BY u.joined_at DESC LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)
		args = append(args, perPage, offset)

		// Execute query
		rows, err := db.Query(baseQuery, args...)
		if err != nil {
			log.Printf("Error querying members: %v", err)
			http.Error(w, "Failed to fetch members", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		members := []Member{}
		for rows.Next() {
			var m Member
			err := rows.Scan(
				&m.ID, &m.Name, &m.GlobalName, &m.DisplayName, &m.Bot, &m.Avatar, &m.AvatarURL,
				&m.Mention, &m.CreatedAt, &m.Nick, &m.JoinedAt, &m.TopRole,
				&m.TopRoleName, &m.TopRoleColor, &m.TimedOutUntil, &m.PremiumSince, &m.UpdatedAt,
			)
			if err != nil {
				log.Printf("Error scanning member: %v", err)
				continue
			}
			members = append(members, m)
		}

		totalPages := (total + perPage - 1) / perPage

		response := MembersResponse{
			Members:    members,
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func GetMemberByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		query := `
			SELECT u.id, u.name, u.global_name, u.display_name, u.bot, u.avatar, u.avatar_url,
			       u.mention, u.created_at, u.nick, u.joined_at, u.top_role,
			       r.name as top_role_name,
			       CASE WHEN r.color > 0 THEN '#' || LPAD(TO_HEX(r.color), 6, '0') ELSE NULL END as top_role_color,
			       u.timed_out_until, u.premium_since, u.updated_at
			FROM users u
			LEFT JOIN roles r ON u.top_role = r.id
			WHERE u.id = $1
		`

		var m Member
		err := db.QueryRow(query, userID).Scan(
			&m.ID, &m.Name, &m.GlobalName, &m.DisplayName, &m.Bot, &m.Avatar, &m.AvatarURL,
			&m.Mention, &m.CreatedAt, &m.Nick, &m.JoinedAt, &m.TopRole,
			&m.TopRoleName, &m.TopRoleColor, &m.TimedOutUntil, &m.PremiumSince, &m.UpdatedAt,
		)
		if err == sql.ErrNoRows {
			http.Error(w, "Member not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Error fetching member: %v", err)
			http.Error(w, "Failed to fetch member", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

func GetMemberStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		stats := MemberStats{}

		// Get total messages
		err := db.QueryRow(`
			SELECT COALESCE(COUNT(*), 0)
			FROM user_messages_logs
			WHERE author_id = $1
		`, userID).Scan(&stats.TotalMessages)
		if err != nil {
			log.Printf("Error counting messages: %v", err)
		}

		// Get voice time stats
		err = db.QueryRow(`
			SELECT 
				COALESCE(SUM(total_duration), 0),
				COALESCE(SUM(muted_duration), 0),
				COALESCE(SUM(deafen_duration), 0),
				COALESCE(SUM(stream_duration), 0),
				COALESCE(COUNT(*), 0)
			FROM user_voice_logs
			WHERE user_id = $1
		`, userID).Scan(
			&stats.TotalVoiceTime,
			&stats.MutedDuration,
			&stats.DeafenDuration,
			&stats.StreamDuration,
			&stats.JoinCount,
		)
		if err != nil {
			log.Printf("Error fetching voice stats: %v", err)
		}

		// Get top text channel
		var topTextChannel TopChannel
		err = db.QueryRow(`
			SELECT c.id, c.name, COUNT(*) as message_count
			FROM user_messages_logs uml
			JOIN channels c ON uml.channel_id = c.id
			WHERE uml.author_id = $1
			GROUP BY c.id, c.name
			ORDER BY message_count DESC
			LIMIT 1
		`, userID).Scan(&topTextChannel.ID, &topTextChannel.Name, &topTextChannel.MessageCount)
		if err == nil {
			stats.TopTextChannel = &topTextChannel
		} else if err != sql.ErrNoRows {
			log.Printf("Error fetching top text channel: %v", err)
		}

		// Get top voice channel
		var topVoiceChannel TopChannel
		err = db.QueryRow(`
			SELECT c.id, c.name, SUM(uvl.total_duration) as total_duration
			FROM user_voice_logs uvl
			JOIN channels c ON uvl.channel_id = c.id
			WHERE uvl.user_id = $1
			GROUP BY c.id, c.name
			ORDER BY total_duration DESC
			LIMIT 1
		`, userID).Scan(&topVoiceChannel.ID, &topVoiceChannel.Name, &topVoiceChannel.Duration)
		if err == nil {
			stats.TopVoiceChannel = &topVoiceChannel
		} else if err != sql.ErrNoRows {
			log.Printf("Error fetching top voice channel: %v", err)
		}

		// Get total joins
		err = db.QueryRow(`
			SELECT COALESCE(COUNT(*), 0)
			FROM user_joins
			WHERE user_id = $1
		`, userID).Scan(&stats.TotalJoins)
		if err != nil {
			log.Printf("Error counting joins: %v", err)
		}

		// Get total leaves
		err = db.QueryRow(`
			SELECT COALESCE(COUNT(*), 0)
			FROM user_leaves
			WHERE user_id = $1
		`, userID).Scan(&stats.TotalLeaves)
		if err != nil {
			log.Printf("Error counting leaves: %v", err)
		}

		// Get total warns
		err = db.QueryRow(`
			SELECT COALESCE(COUNT(*), 0)
			FROM user_moderation_logs
			WHERE user_id = $1 AND type = 'WARN'
		`, userID).Scan(&stats.TotalWarns)
		if err != nil {
			log.Printf("Error counting warns: %v", err)
		}

		// Get user roles
		rows, err := db.Query(`
			SELECT r.id, r.name, 
			       CASE WHEN r.color > 0 THEN '#' || LPAD(TO_HEX(r.color), 6, '0') ELSE NULL END as color
			FROM user_roles ur
			JOIN roles r ON ur.role_id = r.id
			WHERE ur.user_id = $1
			ORDER BY r.position DESC
		`, userID)
		if err != nil {
			log.Printf("Error fetching roles: %v", err)
		} else {
			defer rows.Close()
			roles := []Role{}
			for rows.Next() {
				var role Role
				if err := rows.Scan(&role.ID, &role.Name, &role.Color); err != nil {
					log.Printf("Error scanning role: %v", err)
					continue
				}
				roles = append(roles, role)
			}
			stats.Roles = roles
		}

		// If no roles, initialize empty array
		if stats.Roles == nil {
			stats.Roles = []Role{}
		}

		// Get warns
		warnRows, err := db.Query(`
			SELECT uml.id, uml.moderator_id, COALESCE(u.display_name, u.name, 'Unknown'), uml.reason, uml.created_at
			FROM user_moderation_logs uml
			LEFT JOIN users u ON uml.moderator_id = u.id
			WHERE uml.user_id = $1 AND uml.type = 'WARN'
			ORDER BY uml.created_at DESC
		`, userID)
		if err != nil {
			log.Printf("Error fetching warns: %v", err)
			stats.Warns = []ModerationLog{}
		} else {
			defer warnRows.Close()
			warns := []ModerationLog{}
			for warnRows.Next() {
				var warn ModerationLog
				if err := warnRows.Scan(&warn.ID, &warn.ModeratorID, &warn.ModeratorName, &warn.Reason, &warn.CreatedAt); err != nil {
					log.Printf("Error scanning warn: %v", err)
					continue
				}
				warns = append(warns, warn)
			}
			stats.Warns = warns
		}

		// Get notes
		noteRows, err := db.Query(`
			SELECT uml.id, uml.moderator_id, COALESCE(u.display_name, u.name, 'Unknown'), uml.reason, uml.created_at
			FROM user_moderation_logs uml
			LEFT JOIN users u ON uml.moderator_id = u.id
			WHERE uml.user_id = $1 AND uml.type = 'NOTE'
			ORDER BY uml.created_at DESC
		`, userID)
		if err != nil {
			log.Printf("Error fetching notes: %v", err)
			stats.Notes = []ModerationLog{}
		} else {
			defer noteRows.Close()
			notes := []ModerationLog{}
			for noteRows.Next() {
				var note ModerationLog
				if err := noteRows.Scan(&note.ID, &note.ModeratorID, &note.ModeratorName, &note.Reason, &note.CreatedAt); err != nil {
					log.Printf("Error scanning note: %v", err)
					continue
				}
				notes = append(notes, note)
			}
			stats.Notes = notes
		}

		// If no warns/notes, initialize empty arrays
		if stats.Warns == nil {
			stats.Warns = []ModerationLog{}
		}
		if stats.Notes == nil {
			stats.Notes = []ModerationLog{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
