package commands

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"discord-bot-template/internal/bot/settings"
	"discord-bot-template/internal/database/tables"
	cmdutils "discord-bot-template/internal/shared/utils/commands"

	"github.com/bwmarrin/discordgo"
)

// SetupUserInfoCommand registriert den /userinfo Command
func SetupUserInfoCommand(s *discordgo.Session, guildID string) error {
	command := &discordgo.ApplicationCommand{
		Name:        "userinfo",
		Description: "Zeigt detaillierte Informationen über einen User",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionUser,
				Name:        "user",
				Description: "Der User dessen Informationen angezeigt werden sollen",
				Required:    true,
			},
		},
	}

	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, command)
	return err
}

// HandleUserInfoCommand behandelt den /userinfo Command
func HandleUserInfoCommand(s *discordgo.Session, i *discordgo.InteractionCreate, db *sql.DB) {
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

	// Extrahiere User
	options := i.ApplicationCommandData().Options
	var targetUser *discordgo.User

	for _, opt := range options {
		if opt.Name == "user" {
			targetUser = opt.UserValue(s)
		}
	}

	if targetUser == nil {
		cmdutils.RespondError(s, i, "User nicht gefunden.")
		return
	}

	// Hole User aus Datenbank
	dbUser, err := tables.GetUserByID(db, targetUser.ID)
	if err != nil {
		log.Printf("Error fetching user from database: %v", err)
	}

	// Hole Message Count
	messageCount, err := tables.CountUserMessages(db, targetUser.ID)
	if err != nil {
		log.Printf("Error counting messages: %v", err)
		messageCount = 0
	}

	// Hole Voice Stats
	totalVoiceTime, err := tables.GetTotalVoiceTimeByUser(db, targetUser.ID)
	if err != nil {
		log.Printf("Error fetching total voice time: %v", err)
		totalVoiceTime = 0
	}

	// Hole meistbesuchten Voice Channel
	mostUsedChannel, mostUsedTime, err := tables.GetMostUsedVoiceChannel(db, targetUser.ID)
	if err != nil {
		log.Printf("Error fetching most used voice channel: %v", err)
		mostUsedChannel = "N/A"
		mostUsedTime = 0
	}

	// Hole Warns
	warnCount, err := tables.CountModerationLogsByType(db, i.GuildID, targetUser.ID, tables.ModerationTypeWarn)
	if err != nil {
		log.Printf("Error counting warns: %v", err)
		warnCount = 0
	}

	// Hole Notes
	noteCount, err := tables.CountModerationLogsByType(db, i.GuildID, targetUser.ID, tables.ModerationTypeNote)
	if err != nil {
		log.Printf("Error counting notes: %v", err)
		noteCount = 0
	}

	// Hole alle Moderation Logs
	modLogs, err := tables.GetModerationLogsByUser(db, i.GuildID, targetUser.ID)
	if err != nil {
		log.Printf("Error fetching moderation logs: %v", err)
		modLogs = []tables.UserModerationLog{}
	}

	// Erstelle Embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📋 User Info: %s", targetUser.Username),
		Description: fmt.Sprintf("<@%s>", targetUser.ID),
		Color:       0x5865F2,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: targetUser.AvatarURL("256"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "👤 User ID",
				Value:  targetUser.ID,
				Inline: true,
			},
			{
				Name:   "📅 Account erstellt",
				Value:  cmdutils.FormatDiscordTimestamp(targetUser.ID),
				Inline: true,
			},
		},
	}

	// User Info aus DB
	if dbUser != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📝 Display Name",
			Value:  dbUser.DisplayName,
			Inline: true,
		})
	}

	// Message Stats
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "💬 Nachrichten",
		Value:  fmt.Sprintf("%d", messageCount),
		Inline: true,
	})

	// Voice Stats
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "🎤 Voice Zeit (Gesamt)",
		Value:  cmdutils.FormatDuration(totalVoiceTime),
		Inline: true,
	})

	if mostUsedChannel != "N/A" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "🔊 Meistgenutzter Channel",
			Value:  fmt.Sprintf("<#%s> (%s)", mostUsedChannel, cmdutils.FormatDuration(mostUsedTime)),
			Inline: false,
		})
	}

	// Moderation Stats
	embed.Fields = append(embed.Fields,
		&discordgo.MessageEmbedField{
			Name:   "⚠️ Warnings",
			Value:  fmt.Sprintf("%d", warnCount),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "📝 Notes",
			Value:  fmt.Sprintf("%d", noteCount),
			Inline: true,
		},
	)

	// Füge letzte Moderation Logs hinzu (max 5)
	if len(modLogs) > 0 {
		modLogText := ""
		maxLogs := 5
		if len(modLogs) < maxLogs {
			maxLogs = len(modLogs)
		}

		for i := 0; i < maxLogs; i++ {
			log := modLogs[i]
			emoji := "⚠️"
			if log.Type == tables.ModerationTypeNote {
				emoji = "📝"
			}

			modLogText += fmt.Sprintf("%s **%s** - %s\n*von <@%s>* - %s\n\n",
				emoji,
				log.Type,
				log.Reason,
				log.ModeratorID,
				log.CreatedAt.Format("02.01.2006 15:04"),
			)
		}

		if modLogText != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "📜 Letzte Moderation Logs",
				Value:  modLogText,
				Inline: false,
			})
		}
	}

	embed.Timestamp = time.Now().Format(time.RFC3339)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// formatDuration formatiert Sekunden in ein lesbares Format
func formatDuration(seconds int64) string {
	if seconds == 0 {
		return "0m"
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// formatDiscordTimestamp extrahiert das Erstellungsdatum aus einer Discord Snowflake ID
func formatDiscordTimestamp(snowflake string) string {
	// Discord Snowflake Epoch: 1420070400000 (01.01.2015)
	// Snowflake format: (timestamp << 22) | (worker_id << 17) | (process_id << 12) | increment

	var id int64
	fmt.Sscanf(snowflake, "%d", &id)

	timestamp := (id >> 22) + 1420070400000
	t := time.Unix(timestamp/1000, 0)

	return t.Format("02.01.2006")
}
