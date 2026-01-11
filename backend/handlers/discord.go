package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// RoleInfo repräsentiert eine Discord-Rolle mit Namen und Farbe
type RoleInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    int    `json:"color"`
	ColorHex string `json:"color_hex"`
	Position int    `json:"position"`
}

// ChannelInfo repräsentiert einen Discord-Channel mit Namen
type ChannelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"`
	Position int    `json:"position"`
}

// GetDiscordRoles gibt alle Rollen der Guild von der Bot-API zurück
func GetDiscordRoles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		botAPIURL := os.Getenv("BOT_API_URL")
		if botAPIURL == "" {
			botAPIURL = "http://localhost:8090"
		}

		// Hole Rollen von Bot-API
		resp, err := http.Get(botAPIURL + "/api/guild/roles")
		if err != nil {
			log.Printf("Error fetching roles from bot API: %v", err)
			http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Bot API returned status %d", resp.StatusCode)
			http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
			return
		}

		// Parse Response
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Error decoding roles response: %v", err)
			http.Error(w, "Failed to parse roles", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// GetDiscordChannels gibt alle Channels der Guild von der Bot-API zurück
func GetDiscordChannels(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		botAPIURL := os.Getenv("BOT_API_URL")
		if botAPIURL == "" {
			botAPIURL = "http://localhost:8090"
		}

		// Optional: Type-Filter durchreichen
		channelType := r.URL.Query().Get("type")
		url := botAPIURL + "/api/guild/channels"
		if channelType != "" {
			url += "?type=" + channelType
		}

		// Hole Channels von Bot-API
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching channels from bot API: %v", err)
			http.Error(w, "Failed to fetch channels", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Bot API returned status %d", resp.StatusCode)
			http.Error(w, "Failed to fetch channels", http.StatusInternalServerError)
			return
		}

		// Parse Response
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("Error decoding channels response: %v", err)
			http.Error(w, "Failed to parse channels", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
