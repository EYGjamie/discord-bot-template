package commands

import (
	"database/sql"
	"fmt"

	"discord-bot-template/shared/database/tables"
	cmdutils "discord-bot-template/bot/utils/commands"
	"discord-bot-template/bot/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// SetupCreateVoiceCommand registriert den /setupcreatevoice Command
func SetupCreateVoiceCommand(s *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "setupcreatevoice",
		Description: "Richtet einen Create-Voice-Channel ein",
		DefaultMemberPermissions: func() *int64 {
			var perms int64 = discordgo.PermissionManageChannels
			return &perms
		}(),
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "Der Voice Channel der als Create-Voice fungieren soll",
				Required:    true,
				ChannelTypes: []discordgo.ChannelType{
					discordgo.ChannelTypeGuildVoice,
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "max_users",
				Description: "Standard-Maximale Anzahl an Usern in erstellten Channels (0 = unbegrenzt)",
				Required:    false,
				MinValue:    func() *float64 { v := 0.0; return &v }(),
				MaxValue:    99,
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
	return err
}

// HandleSetupCreateVoiceCommand behandelt den /setupcreatevoice Command
func HandleSetupCreateVoiceCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	logger := logging.NewLogger(db, s, i.GuildID, "commands.setupcreatevoice")

	// Prüfe Admin-Berechtigung
	if i.Member == nil || (i.Member.Permissions&discordgo.PermissionManageChannels) == 0 {
		cmdutils.RespondError(s, i, "Du benötigst die Berechtigung 'Kanäle verwalten' um diesen Command zu nutzen.")
		return
	}

	options := i.ApplicationCommandData().Options
	channelID := options[0].ChannelValue(s).ID
	maxUsers := 0

	if len(options) > 1 {
		maxUsers = int(options[1].IntValue())
	}

	// Prüfe ob der Channel existiert und ein Voice Channel ist
	channel, err := s.Channel(channelID)
	if err != nil {
		logger.LogError("Channel nicht gefunden", err.Error(), "")
		cmdutils.RespondError(s, i, "Channel konnte nicht gefunden werden.")
		return
	}

	if channel.Type != discordgo.ChannelTypeGuildVoice {
		cmdutils.RespondError(s, i, "Der ausgewählte Channel muss ein Voice Channel sein.")
		return
	}

	// Finde oder erstelle einen Text-Channel für das Control Panel
	controlChannelID, err := getOrCreateControlChannel(s, i.GuildID, channelID, channel.Name)
	if err != nil {
		logger.LogError("Control Channel konnte nicht erstellt werden", err.Error(), "")
		cmdutils.RespondError(s, i, "Fehler beim Erstellen des Control Channels.")
		return
	}

	// Speichere die Einstellung in der Datenbank
	setting := &tables.CreateVoiceSetting{
		GuildID:          i.GuildID,
		ChannelID:        channelID,
		DefaultUserLimit: maxUsers,
		ControlChannelID: controlChannelID,
	}

	setting, err = tables.UpsertCreateVoiceSetting(db, setting)
	if err != nil {
		logger.LogError("Fehler beim Speichern der Einstellung", err.Error(), "")
		cmdutils.RespondError(s, i, "Fehler beim Speichern der Einstellungen.")
		return
	}

	// Poste das Control Panel im Control Channel
	err = postControlPanelMessage(s, db, controlChannelID, setting.ID)
	if err != nil {
		logger.LogError("Fehler beim Posten des Control Panels", err.Error(), "")
		// Nicht kritisch, fahren fort
	}

	// Erfolgreiche Antwort
	embed := &discordgo.MessageEmbed{
		Title:       "✅ Create Voice eingerichtet",
		Description: fmt.Sprintf("Der Channel <#%s> wurde als Create-Voice-Channel eingerichtet.", channelID),
		Color:       0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Standard Max Users",
				Value:  fmt.Sprintf("%d", maxUsers),
				Inline: true,
			},
			{
				Name:   "Control Channel",
				Value:  fmt.Sprintf("<#%s>", controlChannelID),
				Inline: true,
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

	logger.LogInfo("Create Voice eingerichtet", fmt.Sprintf("Channel: %s, Max Users: %d", channelID, maxUsers), false)
}

// getOrCreateControlChannel findet oder erstellt einen Control Channel
func getOrCreateControlChannel(s *discordgo.Session, guildID, voiceChannelID, voiceChannelName string) (string, error) {
	// Hole den Voice Channel um die Position zu bekommen
	voiceChannel, err := s.Channel(voiceChannelID)
	if err != nil {
		return "", err
	}

	// Suche nach einem existierenden Text Channel mit dem Namen
	controlChannelName := voiceChannelName + "-control"
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return "", err
	}

	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildText && ch.Name == controlChannelName && ch.ParentID == voiceChannel.ParentID {
			return ch.ID, nil
		}
	}

	// Erstelle einen neuen Text Channel
	newChannel, err := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:     controlChannelName,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: voiceChannel.ParentID,
		Position: voiceChannel.Position + 1,
	})

	if err != nil {
		return "", err
	}

	return newChannel.ID, nil
}

// postControlPanelMessage postet die Control Panel Message im Control Channel
func postControlPanelMessage(s *discordgo.Session, db *sql.DB, channelID string, settingID int64) error {
	embed := &discordgo.MessageEmbed{
		Title:       "🎙️ Voice Channel Kontrolle",
		Description: "Wenn du einen temporären Voice Channel erstellst, kannst du ihn hier verwalten:",
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "✏️ Name ändern",
				Value: "Ändere den Namen deines Channels",
			},
			{
				Name:  "👥 User Limit",
				Value: "Setze die maximale Anzahl an Usern",
			},
			{
				Name:  "👢 Kick",
				Value: "Kicke einen User aus deinem Channel",
			},
			{
				Name:  "🚫 Block",
				Value: "Blockiere einen User vom Betreten deines Channels",
			},
			{
				Name:  "✅ Unblock",
				Value: "Hebe die Blockierung eines Users auf",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Nur der Ersteller des Channels kann diese Funktionen nutzen",
		},
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "✏️ Name",
					Style:    discordgo.PrimaryButton,
					CustomID: "cv_rename",
				},
				discordgo.Button{
					Label:    "👥 Limit",
					Style:    discordgo.PrimaryButton,
					CustomID: "cv_limit",
				},
				discordgo.Button{
					Label:    "👢 Kick",
					Style:    discordgo.SecondaryButton,
					CustomID: "cv_kick",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🚫 Block",
					Style:    discordgo.DangerButton,
					CustomID: "cv_block",
				},
				discordgo.Button{
					Label:    "✅ Unblock",
					Style:    discordgo.SuccessButton,
					CustomID: "cv_unblock",
				},
			},
		},
	}

	msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})

	if err != nil {
		return err
	}

	// Speichere die Message-ID in der Datenbank
	_, err = db.Exec("UPDATE create_voice_settings SET control_message_id = $1 WHERE id = $2", msg.ID, settingID)
	return err
}
