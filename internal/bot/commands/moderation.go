package commands

import (
	"fmt"

	"discord-bot-template/internal/bot/settings"

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
		respondError(s, i, "Dieser Command kann nur auf einem Server ausgeführt werden")
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
		respondError(s, i, "Du benötigst Administrator-Rechte für diesen Command")
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
	}
}

func handleSetChannel(s *discordgo.Session, i *discordgo.InteractionCreate, settingsManager *settings.Manager, options []*discordgo.ApplicationCommandInteractionDataOption) {
	channelID := options[0].ChannelValue(s).ID

	err := settingsManager.SetString("moderation_channel_id", channelID, true)
	if err != nil {
		respondError(s, i, fmt.Sprintf("Fehler beim Speichern: %v", err))
		return
	}

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
	enabled := options[0].BoolValue()

	err := settingsManager.SetEnabled("log_message_edits", enabled)
	if err != nil {
		// Falls Setting nicht existiert, erstelle es
		err = settingsManager.SetBool("log_message_edits", true, enabled)
		if err != nil {
			respondError(s, i, fmt.Sprintf("Fehler beim Speichern: %v", err))
			return
		}
	}

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
	enabled := options[0].BoolValue()

	err := settingsManager.SetEnabled("log_message_deletes", enabled)
	if err != nil {
		// Falls Setting nicht existiert, erstelle es
		err = settingsManager.SetBool("log_message_deletes", true, enabled)
		if err != nil {
			respondError(s, i, fmt.Sprintf("Fehler beim Speichern: %v", err))
			return
		}
	}

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
	editsEnabled := settingsManager.IsEnabled("log_message_edits")
	deletesEnabled := settingsManager.IsEnabled("log_message_deletes")

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

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "❌ Fehler",
					Description: message,
					Color:       0xFF0000,
				},
			},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}
