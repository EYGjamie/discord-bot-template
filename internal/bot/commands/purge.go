package commands

import (
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"discord-bot-template/internal/database/tables"
	"discord-bot-template/internal/shared/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// SetupPurgeCommand registriert den /purge Command
func SetupPurgeCommand(s *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "purge",
		Description: "Löscht alle Nachrichten in einem Channel",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Der Channel, der geleert werden soll",
				Required:    true,
				ChannelTypes: []discordgo.ChannelType{
					discordgo.ChannelTypeGuildText,
					discordgo.ChannelTypeGuildNews,
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "limit",
				Description: "Maximale Anzahl der zu löschenden Nachrichten (Standard: alle)",
				Required:    false,
				MinValue:    float64Ptr(1),
				MaxValue:    1000,
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
	return err
}

// SetupPurgeScheduleCommand registriert den /purge-schedule Command
func SetupPurgeScheduleCommand(s *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "purge-schedule",
		Description: "Verwaltet geplante automatische Channel-Purges",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set",
				Description: "Richtet einen täglichen Purge ein",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "Der Channel, der täglich geleert werden soll",
						Required:    true,
						ChannelTypes: []discordgo.ChannelType{
							discordgo.ChannelTypeGuildText,
							discordgo.ChannelTypeGuildNews,
						},
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "time",
						Description: "Uhrzeit im Format HH:MM (24-Stunden-Format, z.B. 14:30)",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Entfernt einen geplanten Purge",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "Der Channel dessen Purge entfernt werden soll",
						Required:    true,
						ChannelTypes: []discordgo.ChannelType{
							discordgo.ChannelTypeGuildText,
							discordgo.ChannelTypeGuildNews,
						},
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "toggle",
				Description: "Aktiviert/Deaktiviert einen geplanten Purge",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "Der Channel dessen Purge aktiviert/deaktiviert werden soll",
						Required:    true,
						ChannelTypes: []discordgo.ChannelType{
							discordgo.ChannelTypeGuildText,
							discordgo.ChannelTypeGuildNews,
						},
					},
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Name:        "enabled",
						Description: "Aktivieren oder deaktivieren",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "Zeigt alle geplanten Purges für diesen Server",
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
	return err
}

// HandlePurgeCommand behandelt den /purge Command
func HandlePurgeCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	// Prüfe ob User Administrator-Berechtigung hat
	if !hasAdminPermission(i.Member) {
		respondWithError(s, i, "Du benötigst Administrator-Berechtigung um diesen Command zu verwenden!")
		return
	}

	options := i.ApplicationCommandData().Options
	channelOption := options[0].ChannelValue(s)

	limit := 0
	if len(options) > 1 {
		limit = int(options[1].IntValue())
	}

	// Bestätige die Interaktion sofort
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		logger := logging.NewLogger(db, s, i.GuildID, "commands.purge")
		logger.LogError("Purge Command Error", fmt.Sprintf("Failed to respond to interaction: %v", err), i.Member.User.ID)
		return
	}

	// Führe den Purge aus
	deletedCount, err := purgeChannel(s, channelOption.ID, limit)
	if err != nil {
		logger := logging.NewLogger(db, s, i.GuildID, "commands.purge")
		logger.LogError("Purge Command Error", fmt.Sprintf("Failed to purge channel %s: %v", channelOption.Name, err), i.Member.User.ID)
		
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(fmt.Sprintf("❌ Fehler beim Löschen der Nachrichten: %v", err)),
		})
		return
	}

	// Erfolgs-Nachricht
	logger := logging.NewLogger(db, s, i.GuildID, "commands.purge")
	logger.LogInfo("Channel Purged", fmt.Sprintf("User %s purged %d messages from channel %s", i.Member.User.Username, deletedCount, channelOption.Name), false)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: strPtr(fmt.Sprintf("✅ %d Nachrichten wurden aus %s gelöscht!", deletedCount, channelOption.Mention())),
	})
}

// HandlePurgeScheduleCommand behandelt den /purge-schedule Command
func HandlePurgeScheduleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	// Prüfe ob User Administrator-Berechtigung hat
	if !hasAdminPermission(i.Member) {
		respondWithError(s, i, "Du benötigst Administrator-Berechtigung um diesen Command zu verwenden!")
		return
	}

	options := i.ApplicationCommandData().Options
	subCommand := options[0].Name

	switch subCommand {
	case "set":
		handlePurgeScheduleSet(s, i, db, options[0].Options)
	case "remove":
		handlePurgeScheduleRemove(s, i, db, options[0].Options)
	case "toggle":
		handlePurgeScheduleToggle(s, i, db, options[0].Options)
	case "list":
		handlePurgeScheduleList(s, i, db)
	}
}

// handlePurgeScheduleSet richtet einen geplanten Purge ein
func handlePurgeScheduleSet(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB, options []*discordgo.ApplicationCommandInteractionDataOption) {
	channelOption := options[0].ChannelValue(s)
	timeStr := options[1].StringValue()

	// Validiere Zeitformat
	timeRegex := regexp.MustCompile(`^([0-1][0-9]|2[0-3]):([0-5][0-9])$`)
	if !timeRegex.MatchString(timeStr) {
		respondWithError(s, i, "Ungültiges Zeitformat! Bitte verwende HH:MM im 24-Stunden-Format (z.B. 14:30)")
		return
	}

	// Speichere die Einstellung
	setting := &tables.ChannelPurgeSetting{
		GuildID:      i.GuildID,
		ChannelID:    channelOption.ID,
		ScheduleTime: timeStr,
		Enabled:      true,
	}

	_, err := tables.InsertChannelPurgeSetting(db, setting)
	if err != nil {
		logger := logging.NewLogger(db, s, i.GuildID, "commands.purge_schedule")
		logger.LogError("Purge Schedule Error", fmt.Sprintf("Failed to save purge schedule for channel %s: %v", channelOption.Name, err), i.Member.User.ID)
		respondWithError(s, i, fmt.Sprintf("Fehler beim Speichern der Einstellung: %v", err))
		return
	}

	// Erfolgs-Nachricht
	logger := logging.NewLogger(db, s, i.GuildID, "commands.purge_schedule")
	logger.LogInfo("Purge Schedule Set", fmt.Sprintf("User %s set daily purge for channel %s at %s", i.Member.User.Username, channelOption.Name, timeStr), false)

	respondWithSuccess(s, i, fmt.Sprintf("✅ Täglicher Purge für %s wurde auf %s Uhr eingerichtet!", channelOption.Mention(), timeStr))
}

// handlePurgeScheduleRemove entfernt einen geplanten Purge
func handlePurgeScheduleRemove(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB, options []*discordgo.ApplicationCommandInteractionDataOption) {
	channelOption := options[0].ChannelValue(s)

	err := tables.DeleteChannelPurgeSetting(db, i.GuildID, channelOption.ID)
	if err != nil {
		logger := logging.NewLogger(db, s, i.GuildID, "commands.purge_schedule")
		logger.LogError("Purge Schedule Error", fmt.Sprintf("Failed to remove purge schedule for channel %s: %v", channelOption.Name, err), i.Member.User.ID)
		respondWithError(s, i, fmt.Sprintf("Fehler beim Entfernen der Einstellung: %v", err))
		return
	}

	// Erfolgs-Nachricht
	logger := logging.NewLogger(db, s, i.GuildID, "commands.purge_schedule")
	logger.LogInfo("Purge Schedule Removed", fmt.Sprintf("User %s removed daily purge for channel %s", i.Member.User.Username, channelOption.Name), false)

	respondWithSuccess(s, i, fmt.Sprintf("✅ Geplanter Purge für %s wurde entfernt!", channelOption.Mention()))
}

// handlePurgeScheduleToggle aktiviert/deaktiviert einen geplanten Purge
func handlePurgeScheduleToggle(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB, options []*discordgo.ApplicationCommandInteractionDataOption) {
	channelOption := options[0].ChannelValue(s)
	enabled := options[1].BoolValue()

	err := tables.UpdatePurgeSettingEnabled(db, i.GuildID, channelOption.ID, enabled)
	if err != nil {
		logger := logging.NewLogger(db, s, i.GuildID, "commands.purge_schedule")
		logger.LogError("Purge Schedule Error", fmt.Sprintf("Failed to toggle purge schedule for channel %s: %v", channelOption.Name, err), i.Member.User.ID)
		respondWithError(s, i, fmt.Sprintf("Fehler beim Ändern der Einstellung: %v", err))
		return
	}

	// Erfolgs-Nachricht
	logger := logging.NewLogger(db, s, i.GuildID, "commands.purge_schedule")
	status := "deaktiviert"
	if enabled {
		status = "aktiviert"
	}
	logger.LogInfo("Purge Schedule Toggled", fmt.Sprintf("User %s %s daily purge for channel %s", i.Member.User.Username, status, channelOption.Name), false)

	statusEmoji := "❌"
	if enabled {
		statusEmoji = "✅"
	}
	respondWithSuccess(s, i, fmt.Sprintf("%s Geplanter Purge für %s wurde %s!", statusEmoji, channelOption.Mention(), status))
}

// handlePurgeScheduleList zeigt alle geplanten Purges
func handlePurgeScheduleList(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	settings, err := tables.GetGuildPurgeSettings(db, i.GuildID)
	if err != nil {
		logger := logging.NewLogger(db, s, i.GuildID, "commands.purge_schedule")
		logger.LogError("Purge Schedule Error", fmt.Sprintf("Failed to get purge schedules: %v", err), i.Member.User.ID)
		respondWithError(s, i, fmt.Sprintf("Fehler beim Abrufen der Einstellungen: %v", err))
		return
	}

	if len(settings) == 0 {
		respondWithSuccess(s, i, "Es sind keine geplanten Purges eingerichtet.")
		return
	}

	// Erstelle Embed mit allen Einstellungen
	embed := &discordgo.MessageEmbed{
		Title:       "📅 Geplante Channel-Purges",
		Description: "Hier sind alle geplanten Purges für diesen Server:",
		Color:       0x3498db,
		Fields:      []*discordgo.MessageEmbedField{},
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	for _, setting := range settings {
		status := "✅ Aktiv"
		if !setting.Enabled {
			status = "❌ Deaktiviert"
		}

		lastRun := "Noch nie"
		if !setting.LastRun.IsZero() {
			lastRun = setting.LastRun.Format("02.01.2006 15:04")
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("<#%s>", setting.ChannelID),
			Value: fmt.Sprintf("**Uhrzeit:** %s\n**Status:** %s\n**Letzter Durchlauf:** %s",
				setting.ScheduleTime, status, lastRun),
			Inline: false,
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

// purgeChannel löscht alle Nachrichten in einem Channel
func purgeChannel(s *discordgo.Session, channelID string, limit int) (int, error) {
	deletedCount := 0
	var beforeID string

	for {
		// Hole Nachrichten in Batches von 100
		batchLimit := 100
		if limit > 0 && limit-deletedCount < 100 {
			batchLimit = limit - deletedCount
		}

		messages, err := s.ChannelMessages(channelID, batchLimit, beforeID, "", "")
		if err != nil {
			return deletedCount, err
		}

		if len(messages) == 0 {
			break
		}

		// Nachrichten nach Alter filtern (nur Nachrichten < 14 Tage können bulk-deleted werden)
		var bulkDeleteIDs []string
		var oldMessages []*discordgo.Message

		twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)

		for _, msg := range messages {
			if msg.Timestamp.After(twoWeeksAgo) {
				bulkDeleteIDs = append(bulkDeleteIDs, msg.ID)
			} else {
				oldMessages = append(oldMessages, msg)
			}
		}

		// Bulk-Delete für neue Nachrichten
		if len(bulkDeleteIDs) > 0 {
			if len(bulkDeleteIDs) == 1 {
				// Einzelne Nachricht löschen
				err = s.ChannelMessageDelete(channelID, bulkDeleteIDs[0])
				if err != nil {
					return deletedCount, err
				}
				deletedCount++
			} else {
				// Bulk-Delete für mehrere Nachrichten
				err = s.ChannelMessagesBulkDelete(channelID, bulkDeleteIDs)
				if err != nil {
					return deletedCount, err
				}
				deletedCount += len(bulkDeleteIDs)
			}
		}

		// Alte Nachrichten einzeln löschen
		for _, msg := range oldMessages {
			err = s.ChannelMessageDelete(channelID, msg.ID)
			if err != nil {
				return deletedCount, err
			}
			deletedCount++
			time.Sleep(100 * time.Millisecond) // Rate limiting
		}

		// Wenn wir ein Limit haben und erreicht haben, stop
		if limit > 0 && deletedCount >= limit {
			break
		}

		// Wenn keine weiteren Nachrichten vorhanden sind, stop
		if len(messages) < batchLimit {
			break
		}

		beforeID = messages[len(messages)-1].ID
		time.Sleep(500 * time.Millisecond) // Rate limiting zwischen Batches
	}

	return deletedCount, nil
}

// Helper-Funktionen
func hasAdminPermission(member *discordgo.Member) bool {
	if member.Permissions&discordgo.PermissionAdministrator != 0 {
		return true
	}
	return false
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("❌ %s", message),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func respondWithSuccess(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func strPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
