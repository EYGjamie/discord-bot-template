package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"discord-bot-template/bot/settings"
	"discord-bot-template/shared/database/tables"
)

// GetDiscordRoles gibt alle konfigurierten Moderator-Rollen aus der DB zurück
func GetDiscordRoles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "GUILD_ID not configured", http.StatusInternalServerError)
			return
		}

		roles, err := settings.GetModeratorRoles(db, guildID)
		if err != nil {
			log.Printf("Error fetching moderator roles: %v", err)
			http.Error(w, "Failed to fetch roles", http.StatusInternalServerError)
			return
		}

		// Konvertiere []string zu []map[string]string für bessere JSON-Struktur
		roleList := make([]map[string]string, 0, len(roles))
		for _, roleID := range roles {
			roleList = append(roleList, map[string]string{
				"id":   roleID,
				"name": "Moderator Role", // Name nicht verfügbar ohne Discord API
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"roles": roleList,
		})
	}
}

// GetDiscordChannels gibt alle konfigurierten Channels aus der DB zurück
func GetDiscordChannels(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "GUILD_ID not configured", http.StatusInternalServerError)
			return
		}

		channels := make([]map[string]interface{}, 0)

		// Moderation Channel
		moderationChannelSetting, err := tables.GetBotSetting(db, "moderation_channel_id_"+guildID)
		if err == nil && moderationChannelSetting.Value != "" {
			channels = append(channels, map[string]interface{}{
				"id":   moderationChannelSetting.Value,
				"name": "Moderation Log Channel",
				"type": "moderation",
			})
		}

		// Create Voice Channels
		createVoiceSettings, err := tables.GetCreateVoiceSettingsByGuildID(db, guildID)
		if err != nil {
			log.Printf("Error fetching create voice settings: %v", err)
		} else {
			for _, cv := range createVoiceSettings {
				channels = append(channels, map[string]interface{}{
					"id":   cv.ChannelID,
					"name": "Create Voice Channel",
					"type": "voice",
				})
				if cv.ControlChannelID != "" {
					channels = append(channels, map[string]interface{}{
						"id":   cv.ControlChannelID,
						"name": "Voice Control Panel",
						"type": "text",
					})
				}
			}
		}

		// Purge Channels
		purgeSettings, err := tables.GetGuildPurgeSettings(db, guildID)
		if err != nil {
			log.Printf("Error fetching purge settings: %v", err)
		} else {
			for _, purge := range purgeSettings {
				channels = append(channels, map[string]interface{}{
					"id":   purge.ChannelID,
					"name": "Auto-Purge Channel",
					"type": "text",
				})
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"channels": channels,
		})
	}
}
