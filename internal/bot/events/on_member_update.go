package events

import (
	"database/sql"

	"discord-bot-template/internal/shared/services/events/user"

	"github.com/bwmarrin/discordgo"
)

func OnGuildMemberUpdate(bot_session *discordgo.Session, memberUpdate *discordgo.GuildMemberUpdate, db *sql.DB) {
	// Aktualisiere User in Datenbank
	user.UpsertUser(db, memberUpdate.User, memberUpdate.Member)
}
