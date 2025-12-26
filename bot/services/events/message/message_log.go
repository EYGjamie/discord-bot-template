package message

import (
	"database/sql"
	"fmt"
	"log"

	"discord-bot-template/shared/database/tables"
	"discord-bot-template/bot/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// LogMessage dokumentiert alle Nachrichten in der Datenbank
func LogMessage(db *sql.DB, user_message *discordgo.MessageCreate) {
	// Prüfe ob der Author existiert (vermeidet NULL-Werte)
	if user_message.Author == nil {
		log.Println("Warnung: Nachricht ohne Author erhalten")
		return
	}

	logger := logging.NewLogger(db, nil, user_message.GuildID, "service.message_log")

	// Logge die Nachricht in der Datenbank
	err := tables.InsertUserMessageLog(db, user_message.Author.ID, user_message.ChannelID)
	if err != nil {
		log.Printf("Fehler beim Loggen der Nachricht: %v", err)
		logger.LogError("Message Log Insert Failed", fmt.Sprintf("Failed to log message from user %s: %v", user_message.Author.Username, err), "")
		return
	}

	log.Printf("Nachricht von User %s in Channel %s geloggt", user_message.Author.ID, user_message.ChannelID)
}
