package commands

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

// RespondError sendet eine Fehlermeldung als ephemeral Message
func RespondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "❌ " + message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// RespondSuccess sendet eine Erfolgs-Embed
func RespondSuccess(s *discordgo.Session, i *discordgo.InteractionCreate, title, description string) {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0x57F287,
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// RespondEmbed sendet ein Custom Embed
func RespondEmbed(s *discordgo.Session, i *discordgo.InteractionCreate, embed *discordgo.MessageEmbed) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// HasAdminPermission prüft ob der User Administrator-Rechte hat
func HasAdminPermission(permissions int64) bool {
	return permissions&discordgo.PermissionAdministrator != 0
}

// FormatDuration formatiert Sekunden in ein lesbares Format (z.B. "2h 30m")
func FormatDuration(seconds int64) string {
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

// FormatDiscordTimestamp extrahiert das Erstellungsdatum aus einer Discord Snowflake ID
func FormatDiscordTimestamp(snowflake string) string {
	// Discord Snowflake Epoch: 1420070400000 (01.01.2015)
	// Snowflake format: (timestamp << 22) | (worker_id << 17) | (process_id << 12) | increment

	var id int64
	fmt.Sscanf(snowflake, "%d", &id)

	timestamp := (id >> 22) + 1420070400000
	t := time.Unix(timestamp/1000, 0)

	return t.Format("02.01.2006")
}

// FormatDiscordTimestampDetailed extrahiert das Erstellungsdatum mit Uhrzeit aus einer Discord Snowflake ID
func FormatDiscordTimestampDetailed(snowflake string) string {
	var id int64
	fmt.Sscanf(snowflake, "%d", &id)

	timestamp := (id >> 22) + 1420070400000
	t := time.Unix(timestamp/1000, 0)

	return t.Format("02.01.2006 15:04:05")
}
