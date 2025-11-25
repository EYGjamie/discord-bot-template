package events

import (
	"database/sql"

	"discord-bot-template/internal/shared/services/events/message"

	"github.com/bwmarrin/discordgo"
)

func OnMessageCreate(bot_session *discordgo.Session, user_message *discordgo.MessageCreate, db *sql.DB) {
	// Logge die Nachricht in der Datenbank
	message.LogMessage(db, user_message)
}
