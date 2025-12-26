package notifications

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"discord-bot-template/shared/database/tables"

	"github.com/bwmarrin/discordgo"
)

// SendInfoNotification sendet eine Info-Notification an alle registrierten User
func SendInfoNotification(s *discordgo.Session, db *sql.DB, guildID, title, description string, fields []*discordgo.MessageEmbedField) error {
	// Hole alle User die Info-Notifications erhalten sollen
	users, err := tables.GetNotificationUsers(db, guildID, tables.NotificationTypeInfo)
	if err != nil {
		return fmt.Errorf("fehler beim Abrufen der Notification-User: %v", err)
	}

	if len(users) == 0 {
		log.Printf("Keine User für Info-Notifications in Guild %s registriert", guildID)
		return nil
	}

	// Erstelle Embed
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0x3498db, // Blau für Info
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Info Notification",
		},
	}

	if len(fields) > 0 {
		embed.Fields = fields
	}

	// Sende DM an alle registrierten User
	successCount := 0
	for _, user := range users {
		channel, err := s.UserChannelCreate(user.UserID)
		if err != nil {
			log.Printf("Fehler beim Erstellen des DM-Kanals für User %s: %v", user.UserID, err)
			continue
		}

		_, err = s.ChannelMessageSendEmbed(channel.ID, embed)
		if err != nil {
			log.Printf("Fehler beim Senden der Info-Notification an User %s: %v", user.UserID, err)
			continue
		}

		successCount++
	}

	log.Printf("Info-Notification '%s' an %d/%d User gesendet", title, successCount, len(users))
	return nil
}

// SendInfoNotificationSimple sendet eine einfache Info-Notification ohne zusätzliche Felder
func SendInfoNotificationSimple(s *discordgo.Session, db *sql.DB, guildID, title, description string) error {
	return SendInfoNotification(s, db, guildID, title, description, nil)
}
