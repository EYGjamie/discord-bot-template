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

// BotSettingsResponse repräsentiert die Bot-Einstellungen für die API
type BotSettingsResponse struct {
	ModeratorRoles      []string                      `json:"moderator_roles"`
	ModerationChannel   string                        `json:"moderation_channel"`
	LogMessageEdits     bool                          `json:"log_message_edits"`
	LogMessageDeletes   bool                          `json:"log_message_deletes"`
	NotificationUsers   []string                      `json:"notification_users"`
	CreateVoiceSettings []*tables.CreateVoiceSetting  `json:"create_voice_settings"`
	PurgeSettings       []*tables.ChannelPurgeSetting `json:"purge_settings"`
}

// ModeratorRoleRequest für das Hinzufügen/Entfernen von Moderator-Rollen
type ModeratorRoleRequest struct {
	RoleID string `json:"role_id"`
}

// ModerationSettingsRequest für Moderation-Einstellungen
type ModerationSettingsRequest struct {
	ChannelID         *string `json:"channel_id,omitempty"`
	LogMessageEdits   *bool   `json:"log_message_edits,omitempty"`
	LogMessageDeletes *bool   `json:"log_message_deletes,omitempty"`
}

// CreateVoiceSettingRequest für Create-Voice-Einstellungen
type CreateVoiceSettingRequest struct {
	ChannelID        string `json:"channel_id"`
	DefaultUserLimit int    `json:"default_user_limit"`
	ControlChannelID string `json:"control_channel_id"`
}

// PurgeSettingRequest für Purge-Einstellungen
type PurgeSettingRequest struct {
	ChannelID    string `json:"channel_id"`
	ScheduleTime string `json:"schedule_time"` // Format: "HH:MM"
	Enabled      bool   `json:"enabled"`
}

// GetBotSettings holt alle Bot-Einstellungen
func GetBotSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			log.Println("GUILD_ID environment variable not set")
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		// Hole Moderator-Rollen
		modRoles, err := settings.GetModeratorRoles(db, guildID)
		if err != nil {
			log.Printf("Error fetching moderator roles: %v", err)
			modRoles = []string{}
		}

		// Hole Moderation-Kanal
		moderationChannelSetting, _ := tables.GetBotSetting(db, "moderation_channel_id_"+guildID)
		moderationChannel := ""
		if moderationChannelSetting != nil {
			moderationChannel = moderationChannelSetting.Value
		}

		// Hole Edit/Delete Logging Settings
		logEditsSetting, _ := tables.GetBotSetting(db, "log_message_edits_"+guildID)
		logEdits := false
		if logEditsSetting != nil && logEditsSetting.Value == "true" {
			logEdits = true
		}

		logDeletesSetting, _ := tables.GetBotSetting(db, "log_message_deletes_"+guildID)
		logDeletes := false
		if logDeletesSetting != nil && logDeletesSetting.Value == "true" {
			logDeletes = true
		}

		// Hole Notification Users
		notificationUsersSetting, _ := tables.GetBotSetting(db, "notification_users_"+guildID)
		notificationUsers := []string{}
		if notificationUsersSetting != nil && notificationUsersSetting.Value != "" {
			json.Unmarshal([]byte(notificationUsersSetting.Value), &notificationUsers)
		}

		// Hole Create-Voice-Settings
		createVoiceSettings, err := tables.GetCreateVoiceSettingsByGuildID(db, guildID)
		if err != nil {
			log.Printf("Error fetching create voice settings: %v", err)
			createVoiceSettings = []*tables.CreateVoiceSetting{}
		}

		// Hole Purge-Settings
		purgeSettings, err := tables.GetGuildPurgeSettings(db, guildID)
		if err != nil {
			log.Printf("Error fetching purge settings: %v", err)
			purgeSettings = []*tables.ChannelPurgeSetting{}
		}

		response := BotSettingsResponse{
			ModeratorRoles:      modRoles,
			ModerationChannel:   moderationChannel,
			LogMessageEdits:     logEdits,
			LogMessageDeletes:   logDeletes,
			NotificationUsers:   notificationUsers,
			CreateVoiceSettings: createVoiceSettings,
			PurgeSettings:       purgeSettings,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// UpdateModeratorRoles aktualisiert die Moderator-Rollen (Admin only)
func UpdateModeratorRoles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		var roles []string
		if err := json.NewDecoder(r.Body).Decode(&roles); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := settings.SetModeratorRoles(db, guildID, roles); err != nil {
			log.Printf("Error setting moderator roles: %v", err)
			http.Error(w, "Failed to update moderator roles", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Moderator roles updated successfully",
		})
	}
}

// UpdateModerationSettings aktualisiert die Moderation-Einstellungen (Admin only)
func UpdateModerationSettings(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		var req ModerationSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Update Moderation Channel
		if req.ChannelID != nil {
			setting := &tables.BotSetting{
				Key:     "moderation_channel_id_" + guildID,
				Value:   *req.ChannelID,
				Type:    "string",
				Enabled: true,
			}
			if _, err := tables.UpsertBotSetting(db, setting); err != nil {
				log.Printf("Error updating moderation channel: %v", err)
				http.Error(w, "Failed to update moderation channel", http.StatusInternalServerError)
				return
			}
		}

		// Update Log Message Edits
		if req.LogMessageEdits != nil {
			value := "false"
			if *req.LogMessageEdits {
				value = "true"
			}
			setting := &tables.BotSetting{
				Key:     "log_message_edits_" + guildID,
				Value:   value,
				Type:    "bool",
				Enabled: true,
			}
			if _, err := tables.UpsertBotSetting(db, setting); err != nil {
				log.Printf("Error updating log message edits: %v", err)
				http.Error(w, "Failed to update log message edits", http.StatusInternalServerError)
				return
			}
		}

		// Update Log Message Deletes
		if req.LogMessageDeletes != nil {
			value := "false"
			if *req.LogMessageDeletes {
				value = "true"
			}
			setting := &tables.BotSetting{
				Key:     "log_message_deletes_" + guildID,
				Value:   value,
				Type:    "bool",
				Enabled: true,
			}
			if _, err := tables.UpsertBotSetting(db, setting); err != nil {
				log.Printf("Error updating log message deletes: %v", err)
				http.Error(w, "Failed to update log message deletes", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Moderation settings updated successfully",
		})
	}
}

// CreateOrUpdateCreateVoiceSetting erstellt oder aktualisiert eine Create-Voice-Einstellung (Admin only)
func CreateOrUpdateCreateVoiceSetting(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		var req CreateVoiceSettingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.ChannelID == "" {
			http.Error(w, "channel_id is required", http.StatusBadRequest)
			return
		}

		setting := &tables.CreateVoiceSetting{
			GuildID:          guildID,
			ChannelID:        req.ChannelID,
			DefaultUserLimit: req.DefaultUserLimit,
			ControlChannelID: req.ControlChannelID,
		}

		result, err := tables.UpsertCreateVoiceSetting(db, setting)
		if err != nil {
			log.Printf("Error creating/updating create voice setting: %v", err)
			http.Error(w, "Failed to create/update create voice setting", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Create voice setting saved successfully",
			"data":    result,
		})
	}
}

// DeleteCreateVoiceSetting löscht eine Create-Voice-Einstellung (Admin only)
func DeleteCreateVoiceSetting(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channelID := r.PathValue("channel_id")
		if channelID == "" {
			http.Error(w, "channel_id is required", http.StatusBadRequest)
			return
		}

		if err := tables.DeleteCreateVoiceSetting(db, channelID); err != nil {
			log.Printf("Error deleting create voice setting: %v", err)
			http.Error(w, "Failed to delete create voice setting", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Create voice setting deleted successfully",
		})
	}
}

// CreateOrUpdatePurgeSetting erstellt oder aktualisiert eine Purge-Einstellung (Admin only)
func CreateOrUpdatePurgeSetting(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		var req PurgeSettingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.ChannelID == "" || req.ScheduleTime == "" {
			http.Error(w, "channel_id and schedule_time are required", http.StatusBadRequest)
			return
		}

		setting := &tables.ChannelPurgeSetting{
			GuildID:      guildID,
			ChannelID:    req.ChannelID,
			ScheduleTime: req.ScheduleTime,
			Enabled:      req.Enabled,
		}

		result, err := tables.InsertChannelPurgeSetting(db, setting)
		if err != nil {
			log.Printf("Error creating/updating purge setting: %v", err)
			http.Error(w, "Failed to create/update purge setting", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Purge setting saved successfully",
			"data":    result,
		})
	}
}

// DeletePurgeSetting löscht eine Purge-Einstellung (Admin only)
func DeletePurgeSetting(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		channelID := r.PathValue("channel_id")

		if channelID == "" {
			http.Error(w, "channel_id is required", http.StatusBadRequest)
			return
		}

		if err := tables.DeleteChannelPurgeSetting(db, guildID, channelID); err != nil {
			log.Printf("Error deleting purge setting: %v", err)
			http.Error(w, "Failed to delete purge setting", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Purge setting deleted successfully",
		})
	}
}
