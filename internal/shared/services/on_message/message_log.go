package on_message

import (
	"database/sql"
	"log"

	"ovgu/internal/database/tables"

	"github.com/bwmarrin/discordgo"
)

// LogMessage dokumentiert alle Nachrichten in der Datenbank
func LogMessage(db *sql.DB, message *discordgo.MessageCreate) {
	// Prüfe ob der Author existiert (vermeidet NULL-Werte)
	if message.Author == nil {
		log.Println("Warnung: Nachricht ohne Author erhalten")
		return
	}

	// Logge die Nachricht in der Datenbank
	err := tables.InsertUserMessageLog(db, message.Author.ID, message.ChannelID)
	if err != nil {
		log.Printf("Fehler beim Loggen der Nachricht: %v", err)
		return
	}

	log.Printf("Nachricht von User %s in Channel %s geloggt", message.Author.ID, message.ChannelID)
}
