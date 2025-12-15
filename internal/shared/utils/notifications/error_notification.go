package notifications

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"discord-bot-template/internal/database/tables"

	"github.com/bwmarrin/discordgo"
)

// SendErrorNotification sendet eine Error-Notification an alle registrierten User
func SendErrorNotification(s *discordgo.Session, db *sql.DB, guildID, title, description string, fields []*discordgo.MessageEmbedField) error {
	// Hole alle User die Error-Notifications erhalten sollen
	users, err := tables.GetNotificationUsers(db, guildID, tables.NotificationTypeError)
	if err != nil {
		return fmt.Errorf("fehler beim Abrufen der Notification-User: %v", err)
	}

	if len(users) == 0 {
		log.Printf("Keine User für Error-Notifications in Guild %s registriert", guildID)
		return nil
	}

	// Erstelle Embed
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       0xe74c3c, // Rot für Error
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Error Notification",
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
			log.Printf("Fehler beim Senden der Error-Notification an User %s: %v", user.UserID, err)
			continue
		}

		successCount++
	}

	log.Printf("Error-Notification '%s' an %d/%d User gesendet", title, successCount, len(users))
	return nil
}

// SendErrorNotificationSimple sendet eine einfache Error-Notification ohne zusätzliche Felder
func SendErrorNotificationSimple(s *discordgo.Session, db *sql.DB, guildID, title, description string) error {
	return SendErrorNotification(s, db, guildID, title, description, nil)
}
