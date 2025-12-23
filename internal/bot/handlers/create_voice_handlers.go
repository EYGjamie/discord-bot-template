package handlers

import (
	"database/sql"
	"fmt"
	"strconv"

	"discord-bot-template/internal/database/tables"
	"discord-bot-template/internal/shared/services/events/voice"
	cmdutils "discord-bot-template/internal/shared/utils/commands"
	"discord-bot-template/internal/shared/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// HandleCreateVoiceButtons behandelt Button-Interaktionen für Create Voice
func HandleCreateVoiceButtons(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	customID := i.MessageComponentData().CustomID
	logger := logging.NewLogger(db, s, i.GuildID, "handlers.create_voice_buttons")

	// Prüfe ob der User in einem temporären Voice Channel ist
	member := i.Member
	if member == nil {
		cmdutils.RespondError(s, i, "Fehler beim Abrufen der Member-Daten.")
		return
	}

	// Hole Voice State des Users
	voiceState, err := s.State.VoiceState(i.GuildID, member.User.ID)
	if err != nil || voiceState == nil || voiceState.ChannelID == "" {
		cmdutils.RespondError(s, i, "Du musst in einem Voice Channel sein um diese Funktion zu nutzen.")
		return
	}

	// Prüfe ob es ein temporärer Channel ist
	tempChannel, err := tables.GetTemporaryVoiceChannelByChannelID(db, voiceState.ChannelID)
	if err != nil {
		logger.LogError("Fehler beim Abrufen des temporären Channels", err.Error(), "")
		cmdutils.RespondError(s, i, "Fehler beim Abrufen der Channel-Daten.")
		return
	}

	if tempChannel == nil {
		cmdutils.RespondError(s, i, "Du bist nicht in einem temporären Voice Channel.")
		return
	}

	// Prüfe ob der User der Owner ist
	if tempChannel.OwnerID != member.User.ID {
		cmdutils.RespondError(s, i, "Nur der Ersteller des Channels kann diese Funktion nutzen.")
		return
	}

	// Handle verschiedene Button-Aktionen
	switch customID {
	case "cv_rename":
		handleRenameModal(s, i)
	case "cv_limit":
		handleLimitModal(s, i)
	case "cv_kick":
		handleKickModal(s, i, voiceState.ChannelID)
	case "cv_block":
		handleBlockModal(s, i, voiceState.ChannelID)
	case "cv_unblock":
		handleUnblockModal(s, i, db, voiceState.ChannelID)
	default:
		cmdutils.RespondError(s, i, "Unbekannte Aktion.")
	}
}

// handleRenameModal zeigt ein Modal zum Umbenennen des Channels
func handleRenameModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "cv_rename_modal",
			Title:    "Channel umbenennen",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "new_name",
							Label:       "Neuer Channel-Name",
							Style:       discordgo.TextInputShort,
							Placeholder: "Mein Channel",
							Required:    true,
							MaxLength:   100,
							MinLength:   1,
						},
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("Fehler beim Anzeigen des Rename-Modals: %v\n", err)
	}
}

// handleLimitModal zeigt ein Modal zum Ändern des User-Limits
func handleLimitModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "cv_limit_modal",
			Title:    "User-Limit ändern",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "new_limit",
							Label:       "Maximale Anzahl an Usern (0 = unbegrenzt)",
							Style:       discordgo.TextInputShort,
							Placeholder: "0",
							Required:    true,
							MaxLength:   2,
							MinLength:   1,
						},
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("Fehler beim Anzeigen des Limit-Modals: %v\n", err)
	}
}

// handleKickModal zeigt ein Modal zum Kicken eines Users
func handleKickModal(s *discordgo.Session, i *discordgo.InteractionCreate, channelID string) {
	// Hole alle User im Channel
	_, err := s.Channel(channelID)
	if err != nil {
		cmdutils.RespondError(s, i, "Fehler beim Abrufen des Channels.")
		return
	}

	guild, err := s.State.Guild(i.GuildID)
	if err != nil {
		cmdutils.RespondError(s, i, "Fehler beim Abrufen der Guild.")
		return
	}

	// Erstelle eine Liste von Usern im Channel (außer dem Owner)
	var usersInChannel []*discordgo.Member
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == channelID && vs.UserID != i.Member.User.ID {
			member, err := s.GuildMember(i.GuildID, vs.UserID)
			if err == nil {
				usersInChannel = append(usersInChannel, member)
			}
		}
	}

	if len(usersInChannel) == 0 {
		cmdutils.RespondError(s, i, "Es sind keine anderen User in deinem Channel.")
		return
	}

	// Erstelle Select Menu mit Usern
	options := []discordgo.SelectMenuOption{}
	for _, member := range usersInChannel {
		options = append(options, discordgo.SelectMenuOption{
			Label: member.User.Username,
			Value: member.User.ID,
			Emoji: &discordgo.ComponentEmoji{Name: "👢"},
		})
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Wähle einen User zum Kicken:",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "cv_kick_select",
							Placeholder: "User auswählen",
							Options:     options,
						},
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("Fehler beim Anzeigen des Kick-Select: %v\n", err)
	}
}

// handleBlockModal zeigt ein Modal zum Blockieren eines Users
func handleBlockModal(s *discordgo.Session, i *discordgo.InteractionCreate, channelID string) {
	// Hole alle User im Channel
	guild, err := s.State.Guild(i.GuildID)
	if err != nil {
		cmdutils.RespondError(s, i, "Fehler beim Abrufen der Guild.")
		return
	}

	// Erstelle eine Liste von Usern im Channel (außer dem Owner)
	var usersInChannel []*discordgo.Member
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == channelID && vs.UserID != i.Member.User.ID {
			member, err := s.GuildMember(i.GuildID, vs.UserID)
			if err == nil {
				usersInChannel = append(usersInChannel, member)
			}
		}
	}

	if len(usersInChannel) == 0 {
		cmdutils.RespondError(s, i, "Es sind keine anderen User in deinem Channel.")
		return
	}

	// Erstelle Select Menu mit Usern
	options := []discordgo.SelectMenuOption{}
	for _, member := range usersInChannel {
		options = append(options, discordgo.SelectMenuOption{
			Label: member.User.Username,
			Value: member.User.ID,
			Emoji: &discordgo.ComponentEmoji{Name: "🚫"},
		})
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Wähle einen User zum Blockieren:",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "cv_block_select",
							Placeholder: "User auswählen",
							Options:     options,
						},
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("Fehler beim Anzeigen des Block-Select: %v\n", err)
	}
}

// handleUnblockModal zeigt ein Modal zum Entblockieren eines Users
func handleUnblockModal(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB, channelID string) {
	// Hole blockierte User
	tempChannel, err := tables.GetTemporaryVoiceChannelByChannelID(db, channelID)
	if err != nil || tempChannel == nil {
		cmdutils.RespondError(s, i, "Fehler beim Abrufen der Channel-Daten.")
		return
	}

	var blockedUsers []string
	err = voice.GetBlockedUsers(tempChannel, &blockedUsers)
	if err != nil || len(blockedUsers) == 0 {
		cmdutils.RespondError(s, i, "Es sind keine User blockiert.")
		return
	}

	// Erstelle Select Menu mit blockierten Usern
	options := []discordgo.SelectMenuOption{}
	for _, userID := range blockedUsers {
		user, err := s.User(userID)
		if err == nil {
			options = append(options, discordgo.SelectMenuOption{
				Label: user.Username,
				Value: user.ID,
				Emoji: &discordgo.ComponentEmoji{Name: "✅"},
			})
		}
	}

	if len(options) == 0 {
		cmdutils.RespondError(s, i, "Es sind keine User blockiert.")
		return
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Wähle einen User zum Entblockieren:",
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.SelectMenu{
							CustomID:    "cv_unblock_select",
							Placeholder: "User auswählen",
							Options:     options,
						},
					},
				},
			},
		},
	})

	if err != nil {
		fmt.Printf("Fehler beim Anzeigen des Unblock-Select: %v\n", err)
	}
}

// HandleCreateVoiceModals behandelt Modal-Submissions für Create Voice
func HandleCreateVoiceModals(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	customID := i.ModalSubmitData().CustomID
	logger := logging.NewLogger(db, s, i.GuildID, "handlers.create_voice_modals")

	// Hole Voice State des Users
	voiceState, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil || voiceState == nil || voiceState.ChannelID == "" {
		cmdutils.RespondError(s, i, "Du musst in einem Voice Channel sein um diese Funktion zu nutzen.")
		return
	}

	channelID := voiceState.ChannelID

	switch customID {
	case "cv_rename_modal":
		data := i.ModalSubmitData()
		newName := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

		err := voice.UpdateChannelName(s, db, channelID, newName, i.Member.User.ID)
		if err != nil {
			logger.LogError("Fehler beim Umbenennen", err.Error(), "")
			cmdutils.RespondError(s, i, "Fehler beim Umbenennen des Channels: "+err.Error())
			return
		}

		respondSuccess(s, i, fmt.Sprintf("✅ Channel wurde umbenannt zu **%s**", newName))

	case "cv_limit_modal":
		data := i.ModalSubmitData()
		limitStr := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

		newLimit, err := strconv.Atoi(limitStr)
		if err != nil || newLimit < 0 || newLimit > 99 {
			cmdutils.RespondError(s, i, "Bitte gib eine gültige Zahl zwischen 0 und 99 ein.")
			return
		}

		err = voice.UpdateChannelLimit(s, db, channelID, newLimit, i.Member.User.ID)
		if err != nil {
			logger.LogError("Fehler beim Ändern des Limits", err.Error(), "")
			cmdutils.RespondError(s, i, "Fehler beim Ändern des Limits: "+err.Error())
			return
		}

		limitText := fmt.Sprintf("%d", newLimit)
		if newLimit == 0 {
			limitText = "unbegrenzt"
		}
		respondSuccess(s, i, fmt.Sprintf("✅ User-Limit wurde auf **%s** gesetzt", limitText))
	}
}

// HandleCreateVoiceSelects behandelt Select-Menu-Interaktionen für Create Voice
func HandleCreateVoiceSelects(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	customID := i.MessageComponentData().CustomID
	logger := logging.NewLogger(db, s, i.GuildID, "handlers.create_voice_selects")

	values := i.MessageComponentData().Values
	if len(values) == 0 {
		cmdutils.RespondError(s, i, "Kein User ausgewählt.")
		return
	}

	targetUserID := values[0]

	// Hole Voice State des Users
	voiceState, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)
	if err != nil || voiceState == nil || voiceState.ChannelID == "" {
		cmdutils.RespondError(s, i, "Du musst in einem Voice Channel sein um diese Funktion zu nutzen.")
		return
	}

	channelID := voiceState.ChannelID

	switch customID {
	case "cv_kick_select":
		err := voice.KickUserFromChannel(s, db, i.GuildID, channelID, targetUserID, i.Member.User.ID)
		if err != nil {
			logger.LogError("Fehler beim Kicken", err.Error(), "")
			cmdutils.RespondError(s, i, "Fehler beim Kicken des Users: "+err.Error())
			return
		}

		targetUser, _ := s.User(targetUserID)
		userName := targetUserID
		if targetUser != nil {
			userName = targetUser.Username
		}
		respondSuccess(s, i, fmt.Sprintf("✅ **%s** wurde aus dem Channel gekickt", userName))

	case "cv_block_select":
		err := voice.BlockUserFromChannel(s, db, i.GuildID, channelID, targetUserID, i.Member.User.ID)
		if err != nil {
			logger.LogError("Fehler beim Blockieren", err.Error(), "")
			cmdutils.RespondError(s, i, "Fehler beim Blockieren des Users: "+err.Error())
			return
		}

		targetUser, _ := s.User(targetUserID)
		userName := targetUserID
		if targetUser != nil {
			userName = targetUser.Username
		}
		respondSuccess(s, i, fmt.Sprintf("✅ **%s** wurde blockiert", userName))

	case "cv_unblock_select":
		err := voice.UnblockUserFromChannel(s, db, i.GuildID, channelID, targetUserID, i.Member.User.ID)
		if err != nil {
			logger.LogError("Fehler beim Entblockieren", err.Error(), "")
			cmdutils.RespondError(s, i, "Fehler beim Entblockieren des Users: "+err.Error())
			return
		}

		targetUser, _ := s.User(targetUserID)
		userName := targetUserID
		if targetUser != nil {
			userName = targetUser.Username
		}
		respondSuccess(s, i, fmt.Sprintf("✅ **%s** wurde entblockiert", userName))
	}
}

// respondSuccess sendet eine Erfolgsantwort
func respondSuccess(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}
