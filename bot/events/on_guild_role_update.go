package events

import (
	"database/sql"

	"discord-bot-template/bot/services/events/role"

	"github.com/bwmarrin/discordgo"
)

func OnGuildRoleUpdate(bot_session *discordgo.Session, roleEvent *discordgo.GuildRoleUpdate, db *sql.DB) {
	// Aktualisiere Rolle in Datenbank
	role.UpsertRole(db, roleEvent.Role)
}
