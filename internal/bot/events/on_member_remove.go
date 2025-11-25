package events

import (
	"database/sql"

	"discord-bot-template/internal/shared/services/events/user"

	"github.com/bwmarrin/discordgo"
)

func OnGuildMemberRemove(bot_session *discordgo.Session, memberRemove *discordgo.GuildMemberRemove, db *sql.DB) {
	// Aktualisiere User in Datenbank (joined_at auf NULL)
	user.RemoveUser(db, memberRemove.User)
}
