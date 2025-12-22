package commands

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"discord-bot-template/internal/bot/settings"
	cmdutils "discord-bot-template/internal/shared/utils/commands"

	"github.com/bwmarrin/discordgo"
)

// SetupSetModRoleCommand registriert den /setmodrole Command
func SetupSetModRoleCommand(s *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "setmodrole",
		Description: "Legt Moderator-Rollen fest (nur für Administratoren)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Fügt eine Moderator-Rolle hinzu",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "Die Rolle die hinzugefügt werden soll",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Entfernt eine Moderator-Rolle",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "Die Rolle die entfernt werden soll",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "Zeigt alle Moderator-Rollen an",
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
	return err
}

// HandleSetModRoleCommand behandelt den /setmodrole Command
func HandleSetModRoleCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	// Prüfe Admin-Berechtigung
	member := i.Member
	if member == nil {
		cmdutils.RespondError(s, i, "Dieser Command kann nur auf einem Server ausgeführt werden.")
		return
	}

	if !cmdutils.HasAdminPermission(member.Permissions) {
		cmdutils.RespondError(s, i, "Du benötigst Administrator-Rechte um diesen Command zu verwenden.")
		return
	}

	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		cmdutils.RespondError(s, i, "Keine Subcommand gefunden.")
		return
	}

	subCommand := options[0]

	switch subCommand.Name {
	case "add":
		handleAddModRole(s, i, db, subCommand.Options)
	case "remove":
		handleRemoveModRole(s, i, db, subCommand.Options)
	case "list":
		handleListModRoles(s, i, db)
	default:
		cmdutils.RespondError(s, i, "Unbekannter Subcommand.")
	}
}

func handleAddModRole(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB, options []*discordgo.ApplicationCommandInteractionDataOption) {
	var roleID string
	for _, opt := range options {
		if opt.Name == "role" {
			roleID = opt.RoleValue(s, i.GuildID).ID
		}
	}

	if roleID == "" {
		cmdutils.RespondError(s, i, "Rolle nicht gefunden.")
		return
	}

	// Hole aktuelle Moderator-Rollen
	modRoles, err := settings.GetModeratorRoles(db, i.GuildID)
	if err != nil {
		log.Printf("Error getting moderator roles: %v", err)
		cmdutils.RespondError(s, i, "Fehler beim Abrufen der Moderator-Rollen.")
		return
	}

	// Prüfe ob Rolle bereits vorhanden
	for _, r := range modRoles {
		if r == roleID {
			cmdutils.RespondError(s, i, "Diese Rolle ist bereits als Moderator-Rolle festgelegt.")
			return
		}
	}

	// Füge Rolle hinzu
	modRoles = append(modRoles, roleID)
	err = settings.SetModeratorRoles(db, i.GuildID, modRoles)
	if err != nil {
		log.Printf("Error setting moderator roles: %v", err)
		cmdutils.RespondError(s, i, "Fehler beim Speichern der Moderator-Rolle.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ Moderator-Rolle hinzugefügt",
		Description: fmt.Sprintf("<@&%s> wurde als Moderator-Rolle festgelegt.", roleID),
		Color:       0x57F287,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func handleRemoveModRole(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB, options []*discordgo.ApplicationCommandInteractionDataOption) {
	var roleID string
	for _, opt := range options {
		if opt.Name == "role" {
			roleID = opt.RoleValue(s, i.GuildID).ID
		}
	}

	if roleID == "" {
		cmdutils.RespondError(s, i, "Rolle nicht gefunden.")
		return
	}

	// Hole aktuelle Moderator-Rollen
	modRoles, err := settings.GetModeratorRoles(db, i.GuildID)
	if err != nil {
		log.Printf("Error getting moderator roles: %v", err)
		cmdutils.RespondError(s, i, "Fehler beim Abrufen der Moderator-Rollen.")
		return
	}

	// Entferne Rolle
	newRoles := []string{}
	found := false
	for _, r := range modRoles {
		if r != roleID {
			newRoles = append(newRoles, r)
		} else {
			found = true
		}
	}

	if !found {
		cmdutils.RespondError(s, i, "Diese Rolle ist keine Moderator-Rolle.")
		return
	}

	err = settings.SetModeratorRoles(db, i.GuildID, newRoles)
	if err != nil {
		log.Printf("Error setting moderator roles: %v", err)
		cmdutils.RespondError(s, i, "Fehler beim Speichern der Moderator-Rollen.")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title:       "✅ Moderator-Rolle entfernt",
		Description: fmt.Sprintf("<@&%s> ist keine Moderator-Rolle mehr.", roleID),
		Color:       0x57F287,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func handleListModRoles(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
	modRoles, err := settings.GetModeratorRoles(db, i.GuildID)
	if err != nil {
		log.Printf("Error getting moderator roles: %v", err)
		cmdutils.RespondError(s, i, "Fehler beim Abrufen der Moderator-Rollen.")
		return
	}

	if len(modRoles) == 0 {
		cmdutils.RespondError(s, i, "Keine Moderator-Rollen festgelegt.")
		return
	}

	// Erstelle Liste der Rollen
	roleList := []string{}
	for _, roleID := range modRoles {
		roleList = append(roleList, fmt.Sprintf("<@&%s>", roleID))
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🛡️ Moderator-Rollen",
		Description: strings.Join(roleList, "\n"),
		Color:       0x5865F2,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
