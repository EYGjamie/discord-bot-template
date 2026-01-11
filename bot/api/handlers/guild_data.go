package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

// RoleResponse repräsentiert eine Discord-Rolle mit allen relevanten Daten
type RoleResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    int    `json:"color"`
	ColorHex string `json:"color_hex"`
	Position int    `json:"position"`
}

// ChannelResponse repräsentiert einen Discord-Channel mit allen relevanten Daten
type ChannelResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     int    `json:"type"` // 0 = Text, 2 = Voice, 4 = Category
	Position int    `json:"position"`
}

// GuildRolesHandler gibt alle Rollen einer Guild zurück
func GuildRolesHandler(session *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "GUILD_ID not configured", http.StatusInternalServerError)
			return
		}

		// Hole Guild-Daten von Discord
		guild, err := session.State.Guild(guildID)
		if err != nil {
			// Wenn nicht im State, hole von API
			guild, err = session.Guild(guildID)
			if err != nil {
				log.Printf("Error fetching guild: %v", err)
				http.Error(w, "Failed to fetch guild data", http.StatusInternalServerError)
				return
			}
		}

		// Konvertiere Rollen zu Response-Format
		roles := make([]RoleResponse, 0, len(guild.Roles))
		for _, role := range guild.Roles {
			// Konvertiere Farbe zu Hex
			colorHex := "#" + strconv.FormatInt(int64(role.Color), 16)
			if role.Color == 0 {
				colorHex = "#99aab5" // Discord default color
			}

			roles = append(roles, RoleResponse{
				ID:       role.ID,
				Name:     role.Name,
				Color:    role.Color,
				ColorHex: colorHex,
				Position: role.Position,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"roles": roles,
		})
	}
}

// GuildChannelsHandler gibt alle Channels einer Guild zurück
func GuildChannelsHandler(session *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "GUILD_ID not configured", http.StatusInternalServerError)
			return
		}

		// Query Parameter für Channel-Typ-Filter
		channelTypeFilter := r.URL.Query().Get("type") // "text", "voice", or empty for all

		// Hole Guild-Channels von Discord
		channels, err := session.GuildChannels(guildID)
		if err != nil {
			log.Printf("Error fetching channels: %v", err)
			http.Error(w, "Failed to fetch channels", http.StatusInternalServerError)
			return
		}

		// Filter und konvertiere Channels
		result := make([]ChannelResponse, 0)
		for _, channel := range channels {
			// Überspringe Kategorien wenn nicht explizit gewünscht
			if channel.Type == discordgo.ChannelTypeGuildCategory {
				continue
			}

			// Filter nach Typ wenn angegeben
			if channelTypeFilter != "" {
				if channelTypeFilter == "text" && channel.Type != discordgo.ChannelTypeGuildText {
					continue
				}
				if channelTypeFilter == "voice" && channel.Type != discordgo.ChannelTypeGuildVoice {
					continue
				}
			}

			result = append(result, ChannelResponse{
				ID:       channel.ID,
				Name:     channel.Name,
				Type:     int(channel.Type),
				Position: channel.Position,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"channels": result,
		})
	}
}
