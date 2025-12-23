package moderation

import (
	"fmt"
	"time"

	"discord-bot-template/internal/bot/settings"

	"github.com/bwmarrin/discordgo"
)

// LogMessageEdit sendet ein Embed für editierte Nachrichten in den Moderations-Kanal
func LogMessageEdit(s *discordgo.Session, settings *settings.Manager, beforeMsg, afterMsg *discordgo.Message) {
	// Prüfe ob Message-Edit-Logging aktiviert ist
	if !settings.GetBool("log_message_edits", false) {
		return
	}

	// Hole Moderations-Channel-ID
	channelID := settings.GetString("moderation_channel_id", "")
	if channelID == "" {
		return
	}

	// Ignoriere Bot-Nachrichten
	if beforeMsg.Author != nil && beforeMsg.Author.Bot {
		return
	}

	// Erstelle Embed
	embed := &discordgo.MessageEmbed{
		Title:       "📝 Nachricht bearbeitet",
		Color:       0xFFA500, // Orange
		Description: fmt.Sprintf("Eine Nachricht wurde in <#%s> bearbeitet", beforeMsg.ChannelID),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📤 Vorher",
				Value:  truncateString(beforeMsg.Content, 1024),
				Inline: false,
			},
			{
				Name:   "📥 Nachher",
				Value:  truncateString(afterMsg.Content, 1024),
				Inline: false,
			},
			{
				Name:   "👤 Autor",
				Value:  fmt.Sprintf("<@%s>", beforeMsg.Author.ID),
				Inline: true,
			},
			{
				Name:   "🆔 Nachrichten-ID",
				Value:  beforeMsg.ID,
				Inline: true,
			},
			{
				Name:   "🔗 Link",
				Value:  fmt.Sprintf("[Zur Nachricht](https://discord.com/channels/%s/%s/%s)", beforeMsg.GuildID, beforeMsg.ChannelID, beforeMsg.ID),
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Message Edit Logger",
		},
	}

	// Füge Autor-Avatar hinzu wenn vorhanden
	if beforeMsg.Author != nil {
		embed.Author = &discordgo.MessageEmbedAuthor{
			Name:    beforeMsg.Author.Username,
			IconURL: beforeMsg.Author.AvatarURL(""),
		}
	}

	// Sende Embed in Moderations-Kanal
	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		// Fehler nicht loggen um Spam zu vermeiden wenn Channel nicht existiert
		return
	}
}

// LogMessageDelete sendet ein Embed für gelöschte Nachrichten in den Moderations-Kanal
func LogMessageDelete(s *discordgo.Session, settings *settings.Manager, msg *discordgo.Message) {
	// Prüfe ob Message-Delete-Logging aktiviert ist
	if !settings.GetBool("log_message_deletes", false) {
		return
	}

	// Hole Moderations-Channel-ID
	channelID := settings.GetString("moderation_channel_id", "")
	if channelID == "" {
		return
	}

	// Ignoriere Bot-Nachrichten
	if msg.Author != nil && msg.Author.Bot {
		return
	}

	// Erstelle Embed
	embed := &discordgo.MessageEmbed{
		Title:       "🗑️ Nachricht gelöscht",
		Color:       0xFF0000, // Rot
		Description: fmt.Sprintf("Eine Nachricht wurde in <#%s> gelöscht", msg.ChannelID),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "💬 Inhalt",
				Value:  truncateString(msg.Content, 1024),
				Inline: false,
			},
			{
				Name:   "👤 Autor",
				Value:  fmt.Sprintf("<@%s>", msg.Author.ID),
				Inline: true,
			},
			{
				Name:   "🆔 Nachrichten-ID",
				Value:  msg.ID,
				Inline: true,
			},
			{
				Name:   "🕒 Gelöscht um",
				Value:  time.Now().Format("15:04:05"),
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Message Delete Logger",
		},
	}

	// Füge Autor-Avatar hinzu wenn vorhanden
	if msg.Author != nil {
		embed.Author = &discordgo.MessageEmbedAuthor{
			Name:    msg.Author.Username,
			IconURL: msg.Author.AvatarURL(""),
		}
	}

	// Füge Attachments hinzu wenn vorhanden
	if len(msg.Attachments) > 0 {
		attachmentList := ""
		for i, att := range msg.Attachments {
			if i < 5 { // Maximal 5 Attachments anzeigen
				attachmentList += fmt.Sprintf("• [%s](%s)\n", att.Filename, att.URL)
			}
		}
		if attachmentList != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "📎 Anhänge",
				Value:  attachmentList,
				Inline: false,
			})
		}
	}

	// Sende Embed in Moderations-Kanal
	_, err := s.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		// Fehler nicht loggen um Spam zu vermeiden wenn Channel nicht existiert
		return
	}
}

// truncateString kürzt einen String auf maxLength und fügt "..." hinzu
func truncateString(s string, maxLength int) string {
	if s == "" {
		return "*Keine Nachricht*"
	}

	if len(s) <= maxLength {
		return s
	}

	return s[:maxLength-3] + "..."
}
