package commands

import (
	"fmt"

	"discord-bot-template/internal/bot/settings"
	"discord-bot-template/internal/database/tables"
	cmdutils "discord-bot-template/internal/shared/utils/commands"
	"discord-bot-template/internal/shared/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// SetupModerationCommand registriert den /moderation Command
func SetupModerationCommand(s *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "moderation",
		Description: "Moderations-Einstellungen verwalten",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set-channel",
				Description: "Setzt den Moderations-Log-Kanal",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "Der Kanal für Moderations-Logs",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "toggle-edits",
				Description: "Aktiviert/Deaktiviert das Logging von bearbeiteten Nachrichten",
				Options: []*discordgo.ApplicationCommandOption{
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
				Name:        "toggle-deletes",
				Description: "Aktiviert/Deaktiviert das Logging von gelöschten Nachrichten",
				Options: []*discordgo.ApplicationCommandOption{
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
				Name:        "status",
				Description: "Zeigt die aktuellen Moderations-Einstellungen",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add-notification-user",
				Description: "Fügt einen User zur Notification-Liste hinzu",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Der User der Notifications erhalten soll",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "type",
						Description: "Art der Notifications",
						Required:    true,
						Choices: []*discordgo.ApplicationCommandOptionChoice{
							{
								Name:  "Info",
								Value: "info",
							},
							{
								Name:  "Error",
								Value: "error",
							},
							{
								Name:  "Beide",
								Value: "both",
							},
						},
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove-notification-user",
				Description: "Entfernt einen User von der Notification-Liste",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionUser,
						Name:        "user",
						Description: "Der User der entfernt werden soll",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list-notification-users",
				Description: "Zeigt alle User die Notifications erhalten",
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
	return err
}

// HandleModerationCommand verarbeitet den /moderation Command
func HandleModerationCommand(s *discordgo.Session, i *discordgo.InteractionCreate, settingsManager *settings.Manager) {
	options := i.ApplicationCommandData().Options

	// Prüfe auf Admin-Rechte
	member := i.Member
	if member == nil {
		cmdutils.RespondError(s, i, "Dieser Command kann nur auf einem Server ausgeführt werden")
		return
	}

	hasAdmin := false
	for _, roleID := range member.Roles {
		role, err := s.State.Role(i.GuildID, roleID)
		if err == nil && (role.Permissions&discordgo.PermissionAdministrator) != 0 {
			hasAdmin = true
			break
		}
	}

	if !hasAdmin {
		cmdutils.RespondError(s, i, "Du benötigst Administrator-Rechte für diesen Command")
		return
	}

	switch options[0].Name {
	case "set-channel":
		handleSetChannel(s, i, settingsManager, options[0].Options)
	case "toggle-edits":
		handleToggleEdits(s, i, settingsManager, options[0].Options)
	case "toggle-deletes":
		handleToggleDeletes(s, i, settingsManager, options[0].Options)
	case "status":
		handleStatus(s, i, settingsManager)
	case "add-notification-user":
		handleAddNotificationUser(s, i, settingsManager, options[0].Options)
	case "remove-notification-user":
		handleRemoveNotificationUser(s, i, settingsManager, options[0].Options)
	case "list-notification-users":
		handleListNotificationUsers(s, i, settingsManager)
	}
}

func handleSetChannel(s *discordgo.Session, i *discordgo.InteractionCreate, settingsManager *settings.Manager, options []*discordgo.ApplicationCommandInteractionDataOption) {
	logger := logging.NewLogger(settingsManager.GetDB(), s, i.GuildID, "bot.commands.moderation")
	channelID := options[0].ChannelValue(s).ID

	err := settingsManager.SetString("moderation_channel_id", channelID, true)
	if err != nil {
		logger.LogError("Moderation Channel Set Failed", fmt.Sprintf("Failed to set moderation channel: %v", err), "")
		cmdutils.RespondError(s, i, fmt.Sprintf("Fehler beim Speichern: %v", err))
		return
	}
	logger.LogInfo("Moderation Channel Updated", fmt.Sprintf("Channel set to %s", channelID), false)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "✅ Moderations-Kanal gesetzt",
					Description: fmt.Sprintf("Moderations-Logs werden nun in <#%s> gepostet", channelID),
					Color:       0x00FF00,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleToggleEdits(s *discordgo.Session, i *discordgo.InteractionCreate, settingsManager *settings.Manager, options []*discordgo.ApplicationCommandInteractionDataOption) {
	logger := logging.NewLogger(settingsManager.GetDB(), s, i.GuildID, "bot.commands.moderation")
	enabled := options[0].BoolValue()

	// Speichere den Boolean-Wert direkt
	err := settingsManager.SetBool("log_message_edits", enabled, true)
	if err != nil {
		logger.LogError("Toggle Edits Failed", fmt.Sprintf("Failed to toggle message edits logging: %v", err), "")
		cmdutils.RespondError(s, i, fmt.Sprintf("Fehler beim Speichern: %v", err))
		return
	}
	logger.LogInfo("Message Edits Logging Toggled", fmt.Sprintf("Message edits logging set to: %v", enabled), false)

	status := "deaktiviert"
	if enabled {
		status = "aktiviert"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "✅ Einstellung aktualisiert",
					Description: fmt.Sprintf("Logging von bearbeiteten Nachrichten wurde **%s**", status),
					Color:       0x00FF00,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleToggleDeletes(s *discordgo.Session, i *discordgo.InteractionCreate, settingsManager *settings.Manager, options []*discordgo.ApplicationCommandInteractionDataOption) {
	logger := logging.NewLogger(settingsManager.GetDB(), s, i.GuildID, "bot.commands.moderation")
	enabled := options[0].BoolValue()

	// Speichere den Boolean-Wert direkt
	err := settingsManager.SetBool("log_message_deletes", enabled, true)
	if err != nil {
		logger.LogError("Toggle Deletes Failed", fmt.Sprintf("Failed to toggle message deletes logging: %v", err), "")
		cmdutils.RespondError(s, i, fmt.Sprintf("Fehler beim Speichern: %v", err))
		return
	}
	logger.LogInfo("Message Deletes Logging Toggled", fmt.Sprintf("Message deletes logging set to: %v", enabled), false)

	status := "deaktiviert"
	if enabled {
		status = "aktiviert"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "✅ Einstellung aktualisiert",
					Description: fmt.Sprintf("Logging von gelöschten Nachrichten wurde **%s**", status),
					Color:       0x00FF00,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleStatus(s *discordgo.Session, i *discordgo.InteractionCreate, settingsManager *settings.Manager) {
	channelID := settingsManager.GetString("moderation_channel_id", "Nicht gesetzt")
	editsEnabled := settingsManager.GetBool("log_message_edits", false)
	deletesEnabled := settingsManager.GetBool("log_message_deletes", false)

	channelDisplay := channelID
	if channelID != "Nicht gesetzt" {
		channelDisplay = fmt.Sprintf("<#%s>", channelID)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "⚙️ Moderations-Einstellungen",
					Color: 0x3498DB,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "📢 Moderations-Kanal",
							Value:  channelDisplay,
							Inline: false,
						},
						{
							Name:   "📝 Message Edits",
							Value:  formatStatus(editsEnabled),
							Inline: true,
						},
						{
							Name:   "🗑️ Message Deletes",
							Value:  formatStatus(deletesEnabled),
							Inline: true,
						},
					},
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func formatStatus(enabled bool) string {
	if enabled {
		return "✅ Aktiviert"
	}
	return "❌ Deaktiviert"
}

func handleAddNotificationUser(s *discordgo.Session, i *discordgo.InteractionCreate, settingsManager *settings.Manager, options []*discordgo.ApplicationCommandInteractionDataOption) {
	user := options[0].UserValue(s)
	notificationType := tables.NotificationType(options[1].StringValue())

	// Füge User zur Notification-Liste hinzu
	_, err := tables.AddNotificationUser(settingsManager.GetDB(), user.ID, i.GuildID, notificationType)
	if err != nil {
		cmdutils.RespondError(s, i, fmt.Sprintf("Fehler beim Hinzufügen des Users: %v", err))
		return
	}

	var typeText string
	switch notificationType {
	case tables.NotificationTypeInfo:
		typeText = "Info"
	case tables.NotificationTypeError:
		typeText = "Error"
	default:
		typeText = "Info und Error"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "✅ User hinzugefügt",
					Description: fmt.Sprintf("%s erhält ab jetzt **%s**-Notifications per DM.", user.Mention(), typeText),
					Color:       0x2ecc71,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleRemoveNotificationUser(s *discordgo.Session, i *discordgo.InteractionCreate, settingsManager *settings.Manager, options []*discordgo.ApplicationCommandInteractionDataOption) {
	user := options[0].UserValue(s)

	err := tables.RemoveNotificationUser(settingsManager.GetDB(), user.ID, i.GuildID)
	if err != nil {
		cmdutils.RespondError(s, i, fmt.Sprintf("Fehler beim Entfernen des Users: %v", err))
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "✅ User entfernt",
					Description: fmt.Sprintf("%s erhält keine Notifications mehr.", user.Mention()),
					Color:       0x2ecc71,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleListNotificationUsers(s *discordgo.Session, i *discordgo.InteractionCreate, settingsManager *settings.Manager) {
	users, err := tables.GetNotificationUsers(settingsManager.GetDB(), i.GuildID, "")
	if err != nil {
		cmdutils.RespondError(s, i, fmt.Sprintf("Fehler beim Abrufen der User: %v", err))
		return
	}

	if len(users) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{
					{
						Title:       "📋 Notification-User",
						Description: "Es sind keine User für Notifications registriert.",
						Color:       0x95a5a6,
					},
				},
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	description := ""
	for _, user := range users {
		var typeText string
		switch user.NotificationType {
		case tables.NotificationTypeInfo:
			typeText = "Info"
		case tables.NotificationTypeError:
			typeText = "Error"
		default:
			typeText = "Info & Error"
		}
		description += fmt.Sprintf("<@%s> - **%s**\n", user.UserID, typeText)
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "📋 Notification-User",
					Description: description,
					Color:       0x3498db,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}
