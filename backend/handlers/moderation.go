package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"discord-bot-template/backend/middleware"
	"discord-bot-template/shared/database/tables"
)

type CreateModerationRequest struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

type ModerationResponse struct {
	ID        int64  `json:"id"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// CreateWarn erstellt einen neuen Warn-Eintrag
func CreateWarn(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateModerationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Get moderator ID from request (authenticated user)
		moderatorID := middleware.GetUserIDFromRequest(r)
		if moderatorID == "" {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		// Validierung
		if req.UserID == "" || req.Reason == "" {
			http.Error(w, "user_id and reason are required", http.StatusBadRequest)
			return
		}

		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Warn in DB eintragen
		var warnID int64
		var createdAt time.Time
		err := db.QueryRow(`
			INSERT INTO user_moderation_logs (guild_id, user_id, moderator_id, type, reason, created_at)
			VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
			RETURNING id, created_at
		`, guildID, req.UserID, moderatorID, "WARN", req.Reason).Scan(&warnID, &createdAt)

		if err != nil {
			log.Printf("Error creating warn: %v", err)
			http.Error(w, "Failed to create warn", http.StatusInternalServerError)
			return
		}

		// Log in logs-Tabelle eintragen
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Warn Created",
			Message: "Warn created for user " + req.UserID + " by moderator " + moderatorID + ". Reason: " + req.Reason,
			Source:  "backend.moderation.warn",
		}
		if err := tables.InsertLog(db, logEntry); err != nil {
			log.Printf("Error logging warn creation: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ModerationResponse{
			ID:        warnID,
			Success:   true,
			Message:   "Warn created successfully",
			CreatedAt: createdAt.Format(time.RFC3339),
		})
	}
}

// CreateNote erstellt einen neuen Note-Eintrag
func CreateNote(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateModerationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Get moderator ID from request (authenticated user)
		moderatorID := middleware.GetUserIDFromRequest(r)
		if moderatorID == "" {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		// Validierung
		if req.UserID == "" || req.Reason == "" {
			http.Error(w, "user_id and reason are required", http.StatusBadRequest)
			return
		}

		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Note in DB eintragen
		var noteID int64
		var createdAt time.Time
		err := db.QueryRow(`
			INSERT INTO user_moderation_logs (guild_id, user_id, moderator_id, type, reason, created_at)
			VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
			RETURNING id, created_at
		`, guildID, req.UserID, moderatorID, "NOTE", req.Reason).Scan(&noteID, &createdAt)

		if err != nil {
			log.Printf("Error creating note: %v", err)
			http.Error(w, "Failed to create note", http.StatusInternalServerError)
			return
		}

		// Log in logs-Tabelle eintragen
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Note Created",
			Message: "Note created for user " + req.UserID + " by moderator " + moderatorID + ". Reason: " + req.Reason,
			Source:  "backend.moderation.note",
		}
		if err := tables.InsertLog(db, logEntry); err != nil {
			log.Printf("Error logging note creation: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ModerationResponse{
			ID:        noteID,
			Success:   true,
			Message:   "Note created successfully",
			CreatedAt: createdAt.Format(time.RFC3339),
		})
	}
}

// DeleteWarn löscht einen Warn-Eintrag
func DeleteWarn(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid warn ID", http.StatusBadRequest)
			return
		}

		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Zuerst die Warn-Daten abrufen für das Log
		var userID, moderatorID, reason string
		err = db.QueryRow(`
			SELECT user_id, moderator_id, reason
			FROM user_moderation_logs
			WHERE id = $1 AND type = 'WARN'
		`, id).Scan(&userID, &moderatorID, &reason)

		if err == sql.ErrNoRows {
			http.Error(w, "Warn not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Error fetching warn: %v", err)
			http.Error(w, "Failed to fetch warn", http.StatusInternalServerError)
			return
		}

		// Warn löschen
		result, err := db.Exec(`
			DELETE FROM user_moderation_logs
			WHERE id = $1 AND type = 'WARN'
		`, id)

		if err != nil {
			log.Printf("Error deleting warn: %v", err)
			http.Error(w, "Failed to delete warn", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Warn not found", http.StatusNotFound)
			return
		}

		// Log in logs-Tabelle eintragen
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Warn Deleted",
			Message: "Warn deleted for user " + userID + " (ID: " + idStr + "). Original reason: " + reason,
			Source:  "backend.moderation.warn",
		}
		if err := tables.InsertLog(db, logEntry); err != nil {
			log.Printf("Error logging warn deletion: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ModerationResponse{
			ID:      id,
			Success: true,
			Message: "Warn deleted successfully",
		})
	}
}

// DeleteNote löscht einen Note-Eintrag
func DeleteNote(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid note ID", http.StatusBadRequest)
			return
		}

		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Zuerst die Note-Daten abrufen für das Log
		var userID, moderatorID, reason string
		err = db.QueryRow(`
			SELECT user_id, moderator_id, reason
			FROM user_moderation_logs
			WHERE id = $1 AND type = 'NOTE'
		`, id).Scan(&userID, &moderatorID, &reason)

		if err == sql.ErrNoRows {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Error fetching note: %v", err)
			http.Error(w, "Failed to fetch note", http.StatusInternalServerError)
			return
		}

		// Note löschen
		result, err := db.Exec(`
			DELETE FROM user_moderation_logs
			WHERE id = $1 AND type = 'NOTE'
		`, id)

		if err != nil {
			log.Printf("Error deleting note: %v", err)
			http.Error(w, "Failed to delete note", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Note not found", http.StatusNotFound)
			return
		}

		// Log in logs-Tabelle eintragen
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Note Deleted",
			Message: "Note deleted for user " + userID + " (ID: " + idStr + "). Original reason: " + reason,
			Source:  "backend.moderation.note",
		}
		if err := tables.InsertLog(db, logEntry); err != nil {
			log.Printf("Error logging note deletion: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ModerationResponse{
			ID:      id,
			Success: true,
			Message: "Note deleted successfully",
		})
	}
}
