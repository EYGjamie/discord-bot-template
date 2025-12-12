package events

import (
	"database/sql"

	"discord-bot-template/internal/shared/services/events/role"

	"github.com/bwmarrin/discordgo"
)

func OnGuildRoleCreate(bot_session *discordgo.Session, roleEvent *discordgo.GuildRoleCreate, db *sql.DB) {
	// Synchronisiere neue Rolle in Datenbank
	role.UpsertRole(db, roleEvent.Role)
}
