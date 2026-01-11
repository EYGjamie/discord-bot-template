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

type MatchCategoryRequest struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

type MatchCategoryResponse struct {
	ID        int64  `json:"id"`
	GuildID   string `json:"guild_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GetMatchCategories holt alle Match-Kategorien
func GetMatchCategories(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		categories, err := tables.GetMatchCategoriesByGuild(db, guildID)
		if err != nil {
			log.Printf("Error fetching match categories: %v", err)
			http.Error(w, "Failed to fetch match categories", http.StatusInternalServerError)
			return
		}

		response := []MatchCategoryResponse{}
		for _, cat := range categories {
			response = append(response, MatchCategoryResponse{
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

// CreateMatchCategory erstellt eine neue Match-Kategorie (Admin only)
func CreateMatchCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req MatchCategoryRequest
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

		category := &tables.MatchCategory{
			GuildID:   guildID,
			Name:      req.Name,
			Color:     req.Color,
			SortOrder: req.SortOrder,
		}

		if err := tables.InsertMatchCategory(db, category); err != nil {
			log.Printf("Error creating match category: %v", err)
			http.Error(w, "Failed to create match category", http.StatusInternalServerError)
			return
		}

		// Log the action
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Match Category Created",
			Message: "Match category '" + req.Name + "' created",
			Source:  "backend.match_categories",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"id":      category.ID,
			"message": "Match category created successfully",
		})
	}
}

// UpdateMatchCategory aktualisiert eine Match-Kategorie (Admin only)
func UpdateMatchCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID := r.PathValue("id")
		id, err := strconv.ParseInt(categoryID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		var req MatchCategoryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validierung
		if req.Name == "" || req.Color == "" {
			http.Error(w, "name and color are required", http.StatusBadRequest)
			return
		}

		category := &tables.MatchCategory{
			ID:        id,
			Name:      req.Name,
			Color:     req.Color,
			SortOrder: req.SortOrder,
		}

		if err := tables.UpdateMatchCategory(db, category); err != nil {
			log.Printf("Error updating match category: %v", err)
			http.Error(w, "Failed to update match category", http.StatusInternalServerError)
			return
		}

		// Log the action
		guildID := os.Getenv("GUILD_ID")
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Match Category Updated",
			Message: "Match category #" + categoryID + " updated",
			Source:  "backend.match_categories",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Match category updated successfully",
		})
	}
}

// DeleteMatchCategory löscht eine Match-Kategorie (Admin only)
func DeleteMatchCategory(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categoryID := r.PathValue("id")
		id, err := strconv.ParseInt(categoryID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		if err := tables.DeleteMatchCategory(db, id); err != nil {
			log.Printf("Error deleting match category: %v", err)
			http.Error(w, "Failed to delete match category", http.StatusInternalServerError)
			return
		}

		// Log the action
		guildID := os.Getenv("GUILD_ID")
		logEntry := &tables.Log{
			GuildID: guildID,
			Level:   tables.LogLevelInfo,
			Title:   "Match Category Deleted",
			Message: "Match category #" + categoryID + " deleted",
			Source:  "backend.match_categories",
		}
		tables.InsertLog(db, logEntry)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Match category deleted successfully",
		})
	}
}
