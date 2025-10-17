package events

import (
	"database/sql"

	"ovgu/internal/shared/services/on_message"

	"github.com/bwmarrin/discordgo"
)

func OnMessageCreate(bot_session *discordgo.Session, message *discordgo.MessageCreate, db *sql.DB) {
	// Logge die Nachricht in der Datenbank
	on_message.LogMessage(db, message)
}
