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

type EventRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	Color       string   `json:"color"`
	Location    string   `json:"location"`
	Tags        []string `json:"tags"`
}

type EventResponse struct {
	ID            int64    `json:"id"`
	GuildID       string   `json:"guild_id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	StartDate     string   `json:"start_date"`
	EndDate       string   `json:"end_date"`
	StartTime     string   `json:"start_time"`
	EndTime       string   `json:"end_time"`
	Color         string   `json:"color"`
	Location      string   `json:"location"`
	Tags          []string `json:"tags"`
	CreatedBy     string   `json:"created_by"`
	CreatorName   string   `json:"creator_name"`
	CreatorAvatar string   `json:"creator_avatar"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// GetEvents holt alle Events
func GetEvents(db *sql.DB) http.HandlerFunc {
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
			SELECT e.id, e.guild_id, e.title, e.description, e.start_date, e.end_date,
			       e.start_time, e.end_time, e.color, e.location, COALESCE(e.tags, '[]'), e.created_by,
			       COALESCE(u.display_name, u.name, 'Unknown') as creator_name,
			       u.avatar,
			       e.created_at, e.updated_at
			FROM events e
			LEFT JOIN users u ON e.created_by = u.id
			WHERE e.guild_id = $1
		`

		args := []interface{}{guildID}

		if year != "" && month != "" {
			query += " AND ((EXTRACT(YEAR FROM e.start_date) = $2 AND EXTRACT(MONTH FROM e.start_date) = $3) OR (EXTRACT(YEAR FROM e.end_date) = $2 AND EXTRACT(MONTH FROM e.end_date) = $3))"
			args = append(args, year, month)
		}

		query += " ORDER BY e.start_date ASC, e.start_time ASC"

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("Error fetching events: %v", err)
			http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		events := []EventResponse{}
		for rows.Next() {
			var event EventResponse
			var startTime, endTime, tagsJSON, creatorAvatar sql.NullString

			err := rows.Scan(
				&event.ID, &event.GuildID, &event.Title, &event.Description,
				&event.StartDate, &event.EndDate, &startTime, &endTime, &event.Color,
				&event.Location, &tagsJSON, &event.CreatedBy, &event.CreatorName,
				&creatorAvatar, &event.CreatedAt, &event.UpdatedAt,
			)
			if err != nil {
				log.Printf("Error scanning event: %v", err)
				continue
			}

			if startTime.Valid {
				event.StartTime = startTime.String
			}
			if endTime.Valid {
				event.EndTime = endTime.String
			}
			if tagsJSON.Valid && tagsJSON.String != "" {
				json.Unmarshal([]byte(tagsJSON.String), &event.Tags)
			}
			if event.Tags == nil {
				event.Tags = []string{}
			}
			if creatorAvatar.Valid {
				event.CreatorAvatar = creatorAvatar.String
			}

			events = append(events, event)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"events": events,
		})
	}
}

// GetEventByID holt ein einzelnes Event
func GetEventByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")

		query := `
			SELECT e.id, e.guild_id, e.title, e.description, e.start_date, e.end_date,
			       e.start_time, e.end_time, e.color, e.location, COALESCE(e.tags, '[]'), e.created_by,
			       COALESCE(u.display_name, u.name, 'Unknown') as creator_name,
			       u.avatar,
			       e.created_at, e.updated_at
			FROM events e
			LEFT JOIN users u ON e.created_by = u.id
			WHERE e.id = $1
		`

		var event EventResponse
		var startTime, endTime, tagsJSON, creatorAvatar sql.NullString

		err := db.QueryRow(query, eventID).Scan(
			&event.ID, &event.GuildID, &event.Title, &event.Description,
			&event.StartDate, &event.EndDate, &startTime, &endTime, &event.Color,
			&event.Location, &tagsJSON, &event.CreatedBy, &event.CreatorName,
			&creatorAvatar, &event.CreatedAt, &event.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Error fetching event: %v", err)
			http.Error(w, "Failed to fetch event", http.StatusInternalServerError)
			return
		}

		if startTime.Valid {
			event.StartTime = startTime.String
		}
		if endTime.Valid {
			event.EndTime = endTime.String
		}
		if tagsJSON.Valid && tagsJSON.String != "" {
			json.Unmarshal([]byte(tagsJSON.String), &event.Tags)
		}
		if event.Tags == nil {
			event.Tags = []string{}
		}
		if creatorAvatar.Valid {
			event.CreatorAvatar = creatorAvatar.String
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(event)
	}
}

// CreateEvent erstellt ein neues Event
func CreateEvent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EventRequest
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

		// Convert tags to JSON
		tagsJSON := "[]"
		if len(req.Tags) > 0 {
			tagsBytes, _ := json.Marshal(req.Tags)
			tagsJSON = string(tagsBytes)
		}

		event := &tables.Event{
			GuildID:     guildID,
			Title:       req.Title,
			Description: req.Description,
			StartDate:   req.StartDate,
			EndDate:     req.EndDate,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Color:       req.Color,
			Location:    req.Location,
			Tags:        tagsJSON,
			CreatedBy:   userID,
		}

		if err := tables.InsertEvent(db, event); err != nil {
			log.Printf("Error creating event: %v", err)
			http.Error(w, "Failed to create event", http.StatusInternalServerError)
			return
		}

		// Log the action
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Event Created",
			Message: "Event '" + req.Title + "' created by user " + userID + " for dates " + req.StartDate + " to " + req.EndDate,
			Source:  "backend.events",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      event.ID,
			"message": "Event created successfully",
		})
	}
}

// UpdateEvent aktualisiert ein bestehendes Event
func UpdateEvent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")
		id, err := strconv.ParseInt(eventID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid event ID", http.StatusBadRequest)
			return
		}

		var req EventRequest
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

		// Check if event exists and get creator + check if user is admin
		var createdBy string
		var isAdmin bool
		err = db.QueryRow(`
			SELECT e.created_by, COALESCE(u.is_admin, false) as is_admin
			FROM events e
			LEFT JOIN users u ON u.id = $1
			WHERE e.id = $2
		`, userID, id).Scan(&createdBy, &isAdmin)

		if err == sql.ErrNoRows {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Error checking event permissions: %v", err)
			http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
			return
		}

		// Check permissions: own event or admin
		if createdBy != userID && !isAdmin {
			http.Error(w, "Forbidden: You can only edit your own events", http.StatusForbidden)
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

		// Convert tags to JSON
		tagsJSON := "[]"
		if len(req.Tags) > 0 {
			tagsBytes, _ := json.Marshal(req.Tags)
			tagsJSON = string(tagsBytes)
		}

		event := &tables.Event{
			ID:          id,
			Title:       req.Title,
			Description: req.Description,
			StartDate:   req.StartDate,
			EndDate:     req.EndDate,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Color:       req.Color,
			Location:    req.Location,
			Tags:        tagsJSON,
		}

		if err := tables.UpdateEvent(db, event); err != nil {
			log.Printf("Error updating event: %v", err)
			http.Error(w, "Failed to update event", http.StatusInternalServerError)
			return
		}

		// Log the action
		guildID := os.Getenv("GUILD_ID")
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Event Updated",
			Message: "Event #" + eventID + " updated by user " + userID,
			Source:  "backend.events",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Event updated successfully",
		})
	}
}

// DeleteEvent löscht ein Event
func DeleteEvent(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")
		id, err := strconv.ParseInt(eventID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid event ID", http.StatusBadRequest)
			return
		}

		// Get user ID from request
		userID := middleware.GetUserIDFromRequest(r)
		if userID == "" {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		// Check if event exists and get creator + check if user is admin
		var createdBy, title string
		var isAdmin bool
		err = db.QueryRow(`
			SELECT e.created_by, e.title, COALESCE(u.is_admin, false) as is_admin
			FROM events e
			LEFT JOIN users u ON u.id = $1
			WHERE e.id = $2
		`, userID, id).Scan(&createdBy, &title, &isAdmin)

		if err == sql.ErrNoRows {
			http.Error(w, "Event not found", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Error checking event permissions: %v", err)
			http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
			return
		}

		// Check permissions: own event or admin
		if createdBy != userID && !isAdmin {
			http.Error(w, "Forbidden: You can only delete your own events", http.StatusForbidden)
			return
		}

		if err := tables.DeleteEvent(db, id); err != nil {
			log.Printf("Error deleting event: %v", err)
			http.Error(w, "Failed to delete event", http.StatusInternalServerError)
			return
		}

		// Log the action
		guildID := os.Getenv("GUILD_ID")
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Event Deleted",
			Message: "Event '" + title + "' (ID: " + eventID + ") deleted by user " + userID,
			Source:  "backend.events",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Event deleted successfully",
		})
	}
}
