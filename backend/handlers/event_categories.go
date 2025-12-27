package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"discord-bot-template/shared/database/tables"
)

type EventCategoryRequest struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

type EventCategoryResponse struct {
	ID        int64  `json:"id"`
	GuildID   string `json:"guild_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GetEventCategories holt alle Event-Kategorien
func GetEventCategories(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		categories, err := tables.GetEventCategoriesByGuild(db, guildID)
		if err != nil {
			log.Printf("Error fetching event categories: %v", err)
			http.Error(w, "Failed to fetch event categories", http.StatusInternalServerError)
			return
		}

		response := []EventCategoryResponse{}
		for _, cat := range categories {
			response = append(response, EventCategoryResponse{
				ID:        cat.ID,
				GuildID:   cat.GuildID,
				Name:      cat.Name,
				Color:     cat.Color,
				SortOrder: cat.SortOrder,
				CreatedAt: cat.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: cat.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"categories": response,
		})
	}
}

// CreateEventCategory erstellt eine neue Event-Kategorie (Admin only)
func CreateEventCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EventCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validierung
		if req.Name == "" || req.Color == "" {
			http.Error(w, "name and color are required", http.StatusBadRequest)
			return
		}

		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		category := &tables.EventCategory{
			GuildID:   guildID,
			Name:      req.Name,
			Color:     req.Color,
			SortOrder: req.SortOrder,
		}

		if err := tables.InsertEventCategory(db, category); err != nil {
			log.Printf("Error creating event category: %v", err)
			http.Error(w, "Failed to create event category", http.StatusInternalServerError)
			return
		}

		// Log the action
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Event Category Created",
			Message: "Event category '" + req.Name + "' created",
			Source:  "backend.event_categories",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      category.ID,
			"message": "Event category created successfully",
		})
	}
}

// UpdateEventCategory aktualisiert eine Event-Kategorie (Admin only)
func UpdateEventCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID := r.PathValue("id")
		id, err := strconv.ParseInt(categoryID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		var req EventCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validierung
		if req.Name == "" || req.Color == "" {
			http.Error(w, "name and color are required", http.StatusBadRequest)
			return
		}

		category := &tables.EventCategory{
			ID:        id,
			Name:      req.Name,
			Color:     req.Color,
			SortOrder: req.SortOrder,
		}

		if err := tables.UpdateEventCategory(db, category); err != nil {
			log.Printf("Error updating event category: %v", err)
			http.Error(w, "Failed to update event category", http.StatusInternalServerError)
			return
		}

		// Log the action
		guildID := os.Getenv("GUILD_ID")
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Event Category Updated",
			Message: "Event category #" + categoryID + " updated",
			Source:  "backend.event_categories",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Event category updated successfully",
		})
	}
}

// DeleteEventCategory löscht eine Event-Kategorie (Admin only)
func DeleteEventCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID := r.PathValue("id")
		id, err := strconv.ParseInt(categoryID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		if err := tables.DeleteEventCategory(db, id); err != nil {
			log.Printf("Error deleting event category: %v", err)
			http.Error(w, "Failed to delete event category", http.StatusInternalServerError)
			return
		}

		// Log the action
		guildID := os.Getenv("GUILD_ID")
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Event Category Deleted",
			Message: "Event category #" + categoryID + " deleted",
			Source:  "backend.event_categories",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Event category deleted successfully",
		})
	}
}
