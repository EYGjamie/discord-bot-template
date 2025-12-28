package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

// GuildMemberCountHandler gibt die Gesamtanzahl der Mitglieder einer Guild zurück
// Unterstützt auch Guilds mit 1000+ Mitgliedern durch Chunking
func GuildMemberCountHandler(botSession *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Prüfe ob Bot-Session verfügbar ist
		if botSession == nil {
			http.Error(w, "Bot not initialized", http.StatusServiceUnavailable)
			return
		}

		// Hole GuildID aus Query-Parameter
		guildID := r.URL.Query().Get("guild_id")
		if guildID == "" {
			http.Error(w, "guild_id parameter is required", http.StatusBadRequest)
			return
		}

		// Hole Guild-Informationen
		guild, err := botSession.State.Guild(guildID)
		if err != nil {
			// Falls nicht im State, versuche direkt von der API zu holen
			guild, err = botSession.Guild(guildID)
			if err != nil {
				log.Printf("Failed to get guild %s: %v", guildID, err)
				http.Error(w, "Guild not found or bot not in guild", http.StatusNotFound)
				return
			}
		}

		// Für Guilds mit vielen Mitgliedern: Request Guild Members
		if guild.MemberCount > len(guild.Members) {
			err = botSession.RequestGuildMembers(guildID, "", 0, "", false)
			if err != nil {
				log.Printf("Failed to request guild members for %s: %v", guildID, err)
			}
			// Hole Guild nochmal nach dem Request
			guild, _ = botSession.State.Guild(guildID)
		}

		resp := map[string]interface{}{
			"guild_id":          guildID,
			"member_count":      guild.MemberCount,
			"cached_members":    len(guild.Members),
			"approximate_count": guild.ApproximateMemberCount,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	}
}

// GuildRoleMemberCountHandler gibt die Anzahl der Mitglieder mit einer bestimmten Rolle zurück
func GuildRoleMemberCountHandler(botSession *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if botSession == nil {
			http.Error(w, "Bot not initialized", http.StatusServiceUnavailable)
			return
		}

		guildID := r.URL.Query().Get("guild_id")
		roleID := r.URL.Query().Get("role_id")

		if guildID == "" || roleID == "" {
			http.Error(w, "guild_id and role_id parameters are required", http.StatusBadRequest)
			return
		}

		// Hole Guild-Informationen
		guild, err := botSession.State.Guild(guildID)
		if err != nil {
			guild, err = botSession.Guild(guildID)
			if err != nil {
				log.Printf("Failed to get guild %s: %v", guildID, err)
				http.Error(w, "Guild not found or bot not in guild", http.StatusNotFound)
				return
			}
		}

		// Request Guild Members für große Guilds
		if guild.MemberCount > len(guild.Members) {
			err = botSession.RequestGuildMembers(guildID, "", 0, "", false)
			if err != nil {
				log.Printf("Failed to request guild members for %s: %v", guildID, err)
			}
			guild, _ = botSession.State.Guild(guildID)
		}

		// Zähle Mitglieder mit der Rolle
		count := 0
		var memberIDs []string
		for _, member := range guild.Members {
			for _, memberRole := range member.Roles {
				if memberRole == roleID {
					count++
					memberIDs = append(memberIDs, member.User.ID)
					break
				}
			}
		}

		// Verifiziere, dass die Rolle existiert
		roleExists := false
		var roleName string
		for _, role := range guild.Roles {
			if role.ID == roleID {
				roleExists = true
				roleName = role.Name
				break
			}
		}

		resp := map[string]interface{}{
			"guild_id":       guildID,
			"role_id":        roleID,
			"role_exists":    roleExists,
			"role_name":      roleName,
			"member_count":   count,
			"cached_members": len(guild.Members),
			"total_members":  guild.MemberCount,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	}
}

// GuildChannelCountHandler gibt die Anzahl der Channels einer Guild zurück (Text & Voice)
func GuildChannelCountHandler(botSession *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if botSession == nil {
			http.Error(w, "Bot not initialized", http.StatusServiceUnavailable)
			return
		}

		guildID := r.URL.Query().Get("guild_id")
		if guildID == "" {
			http.Error(w, "guild_id parameter is required", http.StatusBadRequest)
			return
		}

		// Hole Guild Channels
		channels, err := botSession.GuildChannels(guildID)
		if err != nil {
			log.Printf("Failed to get channels for guild %s: %v", guildID, err)
			http.Error(w, "Failed to get guild channels", http.StatusInternalServerError)
			return
		}

		// Zähle verschiedene Channel-Typen
		textChannels := 0
		voiceChannels := 0
		categoryChannels := 0
		otherChannels := 0

		for _, channel := range channels {
			switch channel.Type {
			case discordgo.ChannelTypeGuildText, discordgo.ChannelTypeGuildNews:
				textChannels++
			case discordgo.ChannelTypeGuildVoice, discordgo.ChannelTypeGuildStageVoice:
				voiceChannels++
			case discordgo.ChannelTypeGuildCategory:
				categoryChannels++
			default:
				otherChannels++
			}
		}

		resp := map[string]interface{}{
			"guild_id":          guildID,
			"total_channels":    len(channels),
			"text_channels":     textChannels,
			"voice_channels":    voiceChannels,
			"category_channels": categoryChannels,
			"other_channels":    otherChannels,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	}
}

// GuildVoiceUserCountHandler gibt die Anzahl der User in Voice Channels einer Guild zurück
func GuildVoiceUserCountHandler(botSession *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if botSession == nil {
			http.Error(w, "Bot not initialized", http.StatusServiceUnavailable)
			return
		}

		guildID := r.URL.Query().Get("guild_id")
		if guildID == "" {
			http.Error(w, "guild_id parameter is required", http.StatusBadRequest)
			return
		}

		// Hole Guild-Informationen
		guild, err := botSession.State.Guild(guildID)
		if err != nil {
			guild, err = botSession.Guild(guildID)
			if err != nil {
				log.Printf("Failed to get guild %s: %v", guildID, err)
				http.Error(w, "Guild not found or bot not in guild", http.StatusNotFound)
				return
			}
		}

		// Zähle User in Voice Channels
		totalVoiceUsers := 0
		voiceChannelDetails := make(map[string]interface{})

		for _, voiceState := range guild.VoiceStates {
			if voiceState.ChannelID != "" {
				totalVoiceUsers++

				// Hole Channel-Informationen
				if _, exists := voiceChannelDetails[voiceState.ChannelID]; !exists {
					channel, err := botSession.Channel(voiceState.ChannelID)
					if err == nil {
						voiceChannelDetails[voiceState.ChannelID] = map[string]interface{}{
							"channel_name": channel.Name,
							"user_count":   1,
						}
					} else {
						voiceChannelDetails[voiceState.ChannelID] = map[string]interface{}{
							"channel_name": "Unknown",
							"user_count":   1,
						}
					}
				} else {
					detail := voiceChannelDetails[voiceState.ChannelID].(map[string]interface{})
					detail["user_count"] = detail["user_count"].(int) + 1
					voiceChannelDetails[voiceState.ChannelID] = detail
				}
			}
		}

		resp := map[string]interface{}{
			"guild_id":              guildID,
			"total_voice_users":     totalVoiceUsers,
			"active_voice_channels": len(voiceChannelDetails),
			"channel_details":       voiceChannelDetails,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Failed to encode response: %v", err)
		}
	}
}
