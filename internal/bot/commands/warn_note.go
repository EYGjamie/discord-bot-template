package commands

import (
	"database/sql"
	"fmt"
	"log"

	"discord-bot-template/internal/bot/settings"
	"discord-bot-template/internal/database/tables"
	cmdutils "discord-bot-template/internal/shared/utils/commands"

	"github.com/bwmarrin/discordgo"
)

// SetupWarnCommand registriert den /warn Command
func SetupWarnCommand(s *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "warn",
		Description: "Gibt einem User eine Verwarnung",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Der User der verwarnt werden soll",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "reason",
				Description: "Grund für die Verwarnung",
				Required:    true,
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
	return err
}

// HandleWarnCommand behandelt den /warn Command
func HandleWarnCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	// Prüfe ob User Moderator ist
	member := i.Member
	if member == nil {
		cmdutils.RespondError(s, i, "Dieser Command kann nur auf einem Server ausgeführt werden.")
		return
	}

	isMod, err := settings.IsModerator(db, i.GuildID, member.Roles)
	if err != nil {
		log.Printf("Error checking moderator status: %v", err)
		cmdutils.RespondError(s, i, "Fehler beim Prüfen der Berechtigung.")
		return
	}

	if !isMod && !cmdutils.HasAdminPermission(member.Permissions) {
		cmdutils.RespondError(s, i, "Du hast keine Berechtigung diesen Command zu verwenden.")
		return
	}

	// Extrahiere Parameter
	options := i.ApplicationCommandData().Options
	var targetUser *discordgo.User
	var reason string

	for _, opt := range options {
		switch opt.Name {
		case "user":
			targetUser = opt.UserValue(s)
		case "reason":
			reason = opt.StringValue()
		}
	}

	if targetUser == nil {
		cmdutils.RespondError(s, i, "User nicht gefunden.")
		return
	}

	// Speichere Warn in Datenbank
	err = tables.InsertModerationLog(db, i.GuildID, targetUser.ID, i.Member.User.ID, tables.ModerationTypeWarn, reason)
	if err != nil {
		log.Printf("Error inserting warn: %v", err)
		cmdutils.RespondError(s, i, "Fehler beim Speichern der Verwarnung.")
		return
	}

	// Hole aktuelle Warn-Anzahl
	warnCount, err := tables.CountModerationLogsByType(db, i.GuildID, targetUser.ID, tables.ModerationTypeWarn)
	if err != nil {
		log.Printf("Error counting warns: %v", err)
		warnCount = 0
	}

	// Bestätigung
	embed := &discordgo.MessageEmbed{
		Title:       "⚠️ Verwarnung erteilt",
		Description: fmt.Sprintf("<@%s> wurde verwarnt.", targetUser.ID),
		Color:       0xff9900,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Grund",
				Value:  reason,
				Inline: false,
			},
			{
				Name:   "Moderator",
				Value:  fmt.Sprintf("<@%s>", i.Member.User.ID),
				Inline: true,
			},
			{
				Name:   "Gesamt Warns",
				Value:  fmt.Sprintf("%d", warnCount),
				Inline: true,
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// SetupNoteCommand registriert den /note Command
func SetupNoteCommand(s *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "note",
		Description: "Fügt eine Notiz zu einem User hinzu",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Der User zu dem eine Notiz hinzugefügt werden soll",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "note",
				Description: "Die Notiz",
				Required:    true,
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
	return err
}

// HandleNoteCommand behandelt den /note Command
func HandleNoteCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	// Prüfe ob User Moderator ist
	member := i.Member
	if member == nil {
		cmdutils.RespondError(s, i, "Dieser Command kann nur auf einem Server ausgeführt werden.")
		return
	}

	isMod, err := settings.IsModerator(db, i.GuildID, member.Roles)
	if err != nil {
		log.Printf("Error checking moderator status: %v", err)
		cmdutils.RespondError(s, i, "Fehler beim Prüfen der Berechtigung.")
		return
	}

	if !isMod && !cmdutils.HasAdminPermission(member.Permissions) {
		cmdutils.RespondError(s, i, "Du hast keine Berechtigung diesen Command zu verwenden.")
		return
	}

	// Extrahiere Parameter
	options := i.ApplicationCommandData().Options
	var targetUser *discordgo.User
	var note string

	for _, opt := range options {
		switch opt.Name {
		case "user":
			targetUser = opt.UserValue(s)
		case "note":
			note = opt.StringValue()
		}
	}

	if targetUser == nil {
		cmdutils.RespondError(s, i, "User nicht gefunden.")
		return
	}

	// Speichere Note in Datenbank
	err = tables.InsertModerationLog(db, i.GuildID, targetUser.ID, i.Member.User.ID, tables.ModerationTypeNote, note)
	if err != nil {
		log.Printf("Error inserting note: %v", err)
		cmdutils.RespondError(s, i, "Fehler beim Speichern der Notiz.")
		return
	}

	// Bestätigung
	embed := &discordgo.MessageEmbed{
		Title:       "📝 Notiz hinzugefügt",
		Description: fmt.Sprintf("Notiz zu <@%s> wurde hinzugefügt.", targetUser.ID),
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Notiz",
				Value:  note,
				Inline: false,
			},
			{
				Name:   "Moderator",
				Value:  fmt.Sprintf("<@%s>", i.Member.User.ID),
				Inline: true,
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}
