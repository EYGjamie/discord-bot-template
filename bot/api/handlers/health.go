package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

// HealthHandler prüft den Status der Bot-API
func HealthHandler(botSession *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"status":        "ok",
			"service":       "bot-api",
			"bot_connected": botSession != nil && botSession.DataReady,
		}

		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(jsonResp); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}
}
