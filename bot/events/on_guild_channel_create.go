package events

import (
	"database/sql"

	"discord-bot-template/bot/services/events/channel"

	"github.com/bwmarrin/discordgo"
)

func OnChannelCreate(bot_session *discordgo.Session, channelCreate *discordgo.ChannelCreate, db *sql.DB) {
	// Synchronisiere neuen Channel in Datenbank
	channel.UpsertChannel(db, channelCreate.Channel)
}
