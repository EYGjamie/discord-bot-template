package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// DashboardStats contains the aggregated statistics for the dashboard
type DashboardStats struct {
	TotalMembers int `json:"total_members"`
	ActiveEvents int `json:"active_events"`
	OpenTasks    int `json:"open_tasks"`
	OverdueTasks int `json:"overdue_tasks"`
	MatchWinRate int `json:"match_win_rate"` // Static for now
}

// ActiveUser represents a user that was active recently
type ActiveUser struct {
	ID           string  `json:"id"`
	DisplayName  string  `json:"display_name"`
	Avatar       string  `json:"avatar"`
	AvatarURL    string  `json:"avatar_url"`
	TopRole      string  `json:"top_role"`
	TopRoleName  *string `json:"top_role_name"`
	TopRoleColor *string `json:"top_role_color"`
	LastActive   string  `json:"last_active"`
	IsOnline     bool    `json:"is_online"`
}

// RecentActivityItem represents a recent activity log item
type RecentActivityItem struct {
	ID           int64  `json:"id"`
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	UserAvatar   string `json:"user_avatar"`
	ActionType   string `json:"action_type"`
	ResourceType string `json:"resource_type"`
	ResourceID   string `json:"resource_id"`
	Timestamp    string `json:"timestamp"`
}

// GetDashboardStats returns aggregated dashboard statistics
func GetDashboardStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := DashboardStats{
			MatchWinRate: 72, // Static for now
		}

		// Get total members count
		err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE bot = false`).Scan(&stats.TotalMembers)
		if err != nil {
			stats.TotalMembers = 0
		}

		// Get active events count (events starting today or in the future)
		err = db.QueryRow(`
			SELECT COUNT(*) FROM events 
			WHERE end_date >= CURRENT_DATE
		`).Scan(&stats.ActiveEvents)
		if err != nil {
			stats.ActiveEvents = 0
		}

		// Get open tasks count (tasks not in 'done' status)
		err = db.QueryRow(`
			SELECT COUNT(*) FROM tasks 
			WHERE status != 'done'
		`).Scan(&stats.OpenTasks)
		if err != nil {
			stats.OpenTasks = 0
		}

		// Get overdue tasks count
		err = db.QueryRow(`
			SELECT COUNT(*) FROM tasks 
			WHERE status != 'done' 
			AND due_date IS NOT NULL 
			AND due_date < CURRENT_DATE
		`).Scan(&stats.OverdueTasks)
		if err != nil {
			stats.OverdueTasks = 0
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}

// GetActiveUsers returns users who were active in the last N minutes, filled up to 10 with recently active users
func GetActiveUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Default to 15 minutes for "online" users
		minutesAgo := 15
		cutoffTime := time.Now().Add(-time.Duration(minutesAgo) * time.Minute)
		minUsers := 10

		// First, get users active in the last 15 minutes (online users)
		query := `
			SELECT DISTINCT ON (wal.user_id)
				wal.user_id,
				COALESCE(u.display_name, u.global_name, u.name, 'Unknown') as display_name,
				COALESCE(u.avatar, '') as avatar,
				COALESCE(u.avatar_url, '') as avatar_url,
				COALESCE(u.top_role, '') as top_role,
				r.name as top_role_name,
				CASE 
					WHEN r.color = 0 THEN NULL 
					ELSE CONCAT('#', LPAD(TO_HEX(r.color), 6, '0'))
				END as top_role_color,
				wal.created_at as last_active,
				true as is_online
			FROM webapp_audit_logs wal
			LEFT JOIN users u ON wal.user_id = u.id
			LEFT JOIN roles r ON u.top_role = r.id
			WHERE wal.user_id IS NOT NULL 
			AND wal.user_id != ''
			AND wal.created_at >= $1
			ORDER BY wal.user_id, wal.created_at DESC
		`

		rows, err := db.Query(query, cutoffTime)
		if err != nil {
			http.Error(w, "Failed to fetch active users: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []ActiveUser
		onlineUserIDs := make(map[string]bool)
		for rows.Next() {
			var user ActiveUser
			var lastActive time.Time
			err := rows.Scan(
				&user.ID,
				&user.DisplayName,
				&user.Avatar,
				&user.AvatarURL,
				&user.TopRole,
				&user.TopRoleName,
				&user.TopRoleColor,
				&lastActive,
				&user.IsOnline,
			)
			if err != nil {
				continue
			}
			user.LastActive = lastActive.Format(time.RFC3339)
			users = append(users, user)
			onlineUserIDs[user.ID] = true
		}

		// If we have less than minUsers, fill up with recently active users (not currently online)
		if len(users) < minUsers {
			needed := minUsers - len(users)

			// Get recently active users who are not in the current online list
			fillQuery := `
				SELECT DISTINCT ON (wal.user_id)
					wal.user_id,
					COALESCE(u.display_name, u.global_name, u.name, 'Unknown') as display_name,
					COALESCE(u.avatar, '') as avatar,
					COALESCE(u.avatar_url, '') as avatar_url,
					COALESCE(u.top_role, '') as top_role,
					r.name as top_role_name,
					CASE 
						WHEN r.color = 0 THEN NULL 
						ELSE CONCAT('#', LPAD(TO_HEX(r.color), 6, '0'))
					END as top_role_color,
					wal.created_at as last_active
				FROM webapp_audit_logs wal
				LEFT JOIN users u ON wal.user_id = u.id
				LEFT JOIN roles r ON u.top_role = r.id
				WHERE wal.user_id IS NOT NULL 
				AND wal.user_id != ''
				AND wal.created_at < $1
				ORDER BY wal.user_id, wal.created_at DESC
			`

			fillRows, err := db.Query(fillQuery, cutoffTime)
			if err == nil {
				defer fillRows.Close()

				type recentUser struct {
					user       ActiveUser
					lastActive time.Time
				}
				var recentUsers []recentUser

				for fillRows.Next() {
					var user ActiveUser
					var lastActive time.Time
					err := fillRows.Scan(
						&user.ID,
						&user.DisplayName,
						&user.Avatar,
						&user.AvatarURL,
						&user.TopRole,
						&user.TopRoleName,
						&user.TopRoleColor,
						&lastActive,
					)
					if err != nil {
						continue
					}
					// Skip if already in online list
					if onlineUserIDs[user.ID] {
						continue
					}
					user.LastActive = lastActive.Format(time.RFC3339)
					user.IsOnline = false
					recentUsers = append(recentUsers, recentUser{user: user, lastActive: lastActive})
				}

				// Sort by last active time (most recent first)
				for i := 0; i < len(recentUsers)-1; i++ {
					for j := i + 1; j < len(recentUsers); j++ {
						if recentUsers[j].lastActive.After(recentUsers[i].lastActive) {
							recentUsers[i], recentUsers[j] = recentUsers[j], recentUsers[i]
						}
					}
				}

				// Add up to needed users
				for i := 0; i < len(recentUsers) && i < needed; i++ {
					users = append(users, recentUsers[i].user)
				}
			}
		}

		if users == nil {
			users = []ActiveUser{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": users,
			"count": len(users),
		})
	}
}

// GetRecentActivity returns recent relevant activity from audit logs
func GetRecentActivity(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get recent relevant activities (exclude VIEW actions, focus on CREATE, UPDATE, DELETE)
		query := `
			SELECT 
				wal.id,
				wal.user_id,
				COALESCE(u.display_name, u.global_name, u.name, 'Unknown') as user_name,
				COALESCE(u.avatar, '') as user_avatar,
				wal.action_type,
				wal.resource_type,
				COALESCE(wal.resource_id, '') as resource_id,
				wal.created_at
			FROM webapp_audit_logs wal
			LEFT JOIN users u ON wal.user_id = u.id
			WHERE wal.user_id IS NOT NULL 
			AND wal.user_id != ''
			AND wal.action_type IN ('CREATE', 'UPDATE', 'DELETE', 'WARN_CREATE', 'NOTE_CREATE', 'LOGIN')
			AND wal.status_code < 400
			ORDER BY wal.created_at DESC
			LIMIT 25
		`

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Failed to fetch recent activity: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var activities []RecentActivityItem
		for rows.Next() {
			var activity RecentActivityItem
			var timestamp time.Time
			err := rows.Scan(
				&activity.ID,
				&activity.UserID,
				&activity.UserName,
				&activity.UserAvatar,
				&activity.ActionType,
				&activity.ResourceType,
				&activity.ResourceID,
				&timestamp,
			)
			if err != nil {
				continue
			}
			activity.Timestamp = timestamp.Format(time.RFC3339)
			activities = append(activities, activity)
		}

		if activities == nil {
			activities = []RecentActivityItem{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"activities": activities,
		})
	}
}
