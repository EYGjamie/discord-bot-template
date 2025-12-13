package events

import (
	"database/sql"

	"discord-bot-template/internal/shared/services/events/channel"

	"github.com/bwmarrin/discordgo"
)

func OnChannelDelete(bot_session *discordgo.Session, channelDelete *discordgo.ChannelDelete, db *sql.DB) {
	// Entferne Channel aus Datenbank
	channel.RemoveChannel(db, channelDelete.ID)
}
