package events

import (
	"database/sql"

	"discord-bot-template/internal/shared/services/events/user"

	"github.com/bwmarrin/discordgo"
)

func OnGuildMemberAdd(bot_session *discordgo.Session, memberAdd *discordgo.GuildMemberAdd, db *sql.DB) {
	// Synchronisiere User in Datenbank
	user.UpsertUser(db, memberAdd.User, memberAdd.Member)
}
