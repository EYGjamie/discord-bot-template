package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	guildinit "discord-bot-template/bot/init"

	"github.com/bwmarrin/discordgo"
)

// GuildSyncHandler führt eine vollständige Guild-Synchronisation durch
func GuildSyncHandler(botSession *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Prüfe ob Bot-Session und DB verfügbar sind
		if botSession == nil || db == nil {
			http.Error(w, "Bot not initialized", http.StatusServiceUnavailable)
			return
		}

		// Parse Request Body
		var req struct {
			GuildID string `json:"guild_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.GuildID == "" {
			http.Error(w, "guild_id is required", http.StatusBadRequest)
			return
		}

		// Führe FullGuildSync aus
		log.Printf("Bot API: Starting FullGuildSync for guild %s", req.GuildID)
		err := guildinit.FullGuildSync(botSession, req.GuildID, db)
		if err != nil {
			log.Printf("Bot API: FullGuildSync failed for guild %s: %v", req.GuildID, err)
			http.Error(w, fmt.Sprintf("Guild sync failed: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("Bot API: FullGuildSync completed successfully for guild %s", req.GuildID)

		// Erfolgreiche Response
		resp := map[string]interface{}{
			"success":  true,
			"message":  fmt.Sprintf("Guild %s synchronized successfully", req.GuildID),
			"guild_id": req.GuildID,
		}

		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(jsonResp); err != nil {
			log.Printf("Failed to write response: %v", err)
		}
	}
}
