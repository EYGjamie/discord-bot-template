package events

import (
	"database/sql"

	"discord-bot-template/internal/shared/services/events/channel"

	"github.com/bwmarrin/discordgo"
)

func OnChannelUpdate(bot_session *discordgo.Session, channelUpdate *discordgo.ChannelUpdate, db *sql.DB) {
	// Aktualisiere Channel in Datenbank
	channel.UpsertChannel(db, channelUpdate.Channel)
}
