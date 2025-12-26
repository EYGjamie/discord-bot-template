package events

import (
	"database/sql"

	"discord-bot-template/bot/services/events/message"
	"discord-bot-template/bot/services/events/user"

	"github.com/bwmarrin/discordgo"
)

func OnMessageCreate(bot_session *discordgo.Session, user_message *discordgo.MessageCreate, db *sql.DB) {
	// Update/Insert User in Datenbank
	if user_message.Author != nil {
		user.UpsertUser(db, user_message.Author, user_message.Member)
	}

	// Logge die Nachricht in der Datenbank
	message.LogMessage(db, user_message)
}
