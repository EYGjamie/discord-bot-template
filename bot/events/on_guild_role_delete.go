package events

import (
	"database/sql"

	"discord-bot-template/bot/services/events/role"

	"github.com/bwmarrin/discordgo"
)

func OnGuildRoleDelete(bot_session *discordgo.Session, roleEvent *discordgo.GuildRoleDelete, db *sql.DB) {
	// Lösche Rolle aus Datenbank
	role.RemoveRole(db, roleEvent.RoleID)
}
