package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"discord-bot-template/backend/middleware"
	"discord-bot-template/shared/database/tables"
)

type MatchRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Color       string `json:"color"`
	Location    string `json:"location"`
	Guests      string `json:"guests"`
}

type MatchResponse struct {
	ID            int64  `json:"id"`
	GuildID       string `json:"guild_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Color         string `json:"color"`
	Location      string `json:"location"`
	Guests        string `json:"guests"`
	CreatedBy     string `json:"created_by"`
	CreatorName   string `json:"creator_name"`
	CreatorAvatar string `json:"creator_avatar"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// GetMatches holt alle Matches
func GetMatches(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Optional: Filter nach Monat/Jahr
		month := r.URL.Query().Get("month")
		year := r.URL.Query().Get("year")

		query := `
			SELECT m.id, m.guild_id, m.title, m.description, m.start_date, m.end_date,
			       m.start_time, m.end_time, m.color, m.location, m.guests, m.created_by,
			       COALESCE(u.display_name, u.name, 'Unknown') as creator_name,
			       u.avatar,
			       m.created_at, m.updated_at
			FROM matches m
			LEFT JOIN users u ON m.created_by = u.id
			WHERE m.guild_id = $1
		`

		args := []interface{}{guildID}

		if year != "" && month != "" {
			query += " AND ((EXTRACT(YEAR FROM m.start_date) = $2 AND EXTRACT(MONTH FROM m.start_date) = $3) OR (EXTRACT(YEAR FROM m.end_date) = $2 AND EXTRACT(MONTH FROM m.end_date) = $3))"
			args = append(args, year, month)
		}

		query += " ORDER BY m.start_date ASC, m.start_time ASC"

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("Error fetching matches: %v", err)
			http.Error(w, "Failed to fetch matches", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		matches := []MatchResponse{}
		for rows.Next() {
			var match MatchResponse
			var startTime, endTime, guests, creatorAvatar sql.NullString

			err := rows.Scan(
				&match.ID, &match.GuildID, &match.Title, &match.Description,
				&match.StartDate, &match.EndDate, &startTime, &endTime, &match.Color,
				&match.Location, &guests, &match.CreatedBy, &match.CreatorName,
				&creatorAvatar, &match.CreatedAt, &match.UpdatedAt,
			)
			if err != nil {
				log.Printf("Error scanning match: %v", err)
				continue
			}

			if startTime.Valid {
				match.StartTime = startTime.String
			}
			if endTime.Valid {
				match.EndTime = endTime.String
			}
			if guests.Valid {
				match.Guests = guests.String
			}
			if creatorAvatar.Valid {
				match.CreatorAvatar = creatorAvatar.String
			}

			matches = append(matches, match)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"matches": matches,
		})
	}
}

// GetMatchByID holt ein einzelnes Match
func GetMatchByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matchID := r.PathValue("id")

		query := `
			SELECT m.id, m.guild_id, m.title, m.description, m.start_date, m.end_date,
			       m.start_time, m.end_time, m.color, m.location, m.guests, m.created_by,
			       COALESCE(u.display_name, u.name, 'Unknown') as creator_name,
			       u.avatar,
			       m.created_at, m.updated_at
			FROM matches m
			LEFT JOIN users u ON m.created_by = u.id
			WHERE m.id = $1
		`

		var match MatchResponse
		var startTime, endTime, guests, creatorAvatar sql.NullString

		err := db.QueryRow(query, matchID).Scan(
			&match.ID, &match.GuildID, &match.Title, &match.Description,
			&match.StartDate, &match.EndDate, &startTime, &endTime, &match.Color,
			&match.Location, &guests, &match.CreatedBy, &match.CreatorName,
			&creatorAvatar, &match.CreatedAt, &match.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Match not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Error fetching match: %v", err)
			http.Error(w, "Failed to fetch match", http.StatusInternalServerError)
			return
		}

		if startTime.Valid {
			match.StartTime = startTime.String
		}
		if endTime.Valid {
			match.EndTime = endTime.String
		}
		if guests.Valid {
			match.Guests = guests.String
		}
		if creatorAvatar.Valid {
			match.CreatorAvatar = creatorAvatar.String
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(match)
	}
}

// CreateMatch erstellt ein neues Match
func CreateMatch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req MatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Get user ID from request
		userID := middleware.GetUserIDFromRequest(r)
		if userID == "" {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		// Validierung
		if req.Title == "" || req.StartDate == "" {
			http.Error(w, "title and start_date are required", http.StatusBadRequest)
			return
		}

		// Wenn end_date nicht gesetzt, auf start_date setzen
		if req.EndDate == "" {
			req.EndDate = req.StartDate
		}

		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Default color wenn nicht gesetzt
		if req.Color == "" {
			req.Color = "#4285F4"
		}

		match := &tables.Match{
			GuildID:     guildID,
			Title:       req.Title,
			Description: req.Description,
			StartDate:   req.StartDate,
			EndDate:     req.EndDate,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Color:       req.Color,
			Location:    req.Location,
			Guests:      req.Guests,
			CreatedBy:   userID,
		}

		if err := tables.InsertMatch(db, match); err != nil {
			log.Printf("Error creating match: %v", err)
			http.Error(w, "Failed to create match", http.StatusInternalServerError)
			return
		}

		// Log the action
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Match Created",
			Message: "Match '" + req.Title + "' created by user " + userID + " for dates " + req.StartDate + " to " + req.EndDate,
			Source:  "backend.matches",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      match.ID,
			"message": "Match created successfully",
		})
	}
}

// UpdateMatch aktualisiert ein bestehendes Match
func UpdateMatch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matchID := r.PathValue("id")
		id, err := strconv.ParseInt(matchID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid match ID", http.StatusBadRequest)
			return
		}

		var req MatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Get user ID from request
		userID := middleware.GetUserIDFromRequest(r)
		if userID == "" {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		// Check if match exists and get creator + check if user is admin
		var createdBy string
		var isAdmin bool
		err = db.QueryRow(`
			SELECT m.created_by, COALESCE(u.is_admin, false) as is_admin
			FROM matches m
			LEFT JOIN users u ON u.id = $1
			WHERE m.id = $2
		`, userID, id).Scan(&createdBy, &isAdmin)

		if err == sql.ErrNoRows {
			http.Error(w, "Match not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Error checking match permissions: %v", err)
			http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
			return
		}

		// Check permissions: own match or admin
		if createdBy != userID && !isAdmin {
			http.Error(w, "Forbidden: You can only edit your own matches", http.StatusForbidden)
			return
		}

		// Validierung
		if req.Title == "" || req.StartDate == "" {
			http.Error(w, "title and start_date are required", http.StatusBadRequest)
			return
		}

		// Wenn end_date nicht gesetzt, auf start_date setzen
		if req.EndDate == "" {
			req.EndDate = req.StartDate
		}

		match := &tables.Match{
			ID:          id,
			Title:       req.Title,
			Description: req.Description,
			StartDate:   req.StartDate,
			EndDate:     req.EndDate,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Color:       req.Color,
			Location:    req.Location,
			Guests:      req.Guests,
		}

		if err := tables.UpdateMatch(db, match); err != nil {
			log.Printf("Error updating match: %v", err)
			http.Error(w, "Failed to update match", http.StatusInternalServerError)
			return
		}

		// Log the action
		guildID := os.Getenv("GUILD_ID")
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Match Updated",
			Message: "Match #" + matchID + " updated by user " + userID,
			Source:  "backend.matches",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Match updated successfully",
		})
	}
}

// DeleteMatch löscht ein Match
func DeleteMatch(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		matchID := r.PathValue("id")
		id, err := strconv.ParseInt(matchID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid match ID", http.StatusBadRequest)
			return
		}

		// Get user ID from request
		userID := middleware.GetUserIDFromRequest(r)
		if userID == "" {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		// Check if match exists and get creator + check if user is admin
		var createdBy, title string
		var isAdmin bool
		err = db.QueryRow(`
			SELECT m.created_by, m.title, COALESCE(u.is_admin, false) as is_admin
			FROM matches m
			LEFT JOIN users u ON u.id = $1
			WHERE m.id = $2
		`, userID, id).Scan(&createdBy, &title, &isAdmin)

		if err == sql.ErrNoRows {
			http.Error(w, "Match not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Error checking match permissions: %v", err)
			http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
			return
		}

		// Check permissions: own match or admin
		if createdBy != userID && !isAdmin {
			http.Error(w, "Forbidden: You can only delete your own matches", http.StatusForbidden)
			return
		}

		if err := tables.DeleteMatch(db, id); err != nil {
			log.Printf("Error deleting match: %v", err)
			http.Error(w, "Failed to delete match", http.StatusInternalServerError)
			return
		}

		// Log the action
		guildID := os.Getenv("GUILD_ID")
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Match Deleted",
			Message: "Match '" + title + "' (ID: " + matchID + ") deleted by user " + userID,
			Source:  "backend.matches",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Match deleted successfully",
		})
	}
}
