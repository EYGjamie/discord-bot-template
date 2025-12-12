package events

import (
	"database/sql"

	"discord-bot-template/internal/shared/services/events/user"

	"github.com/bwmarrin/discordgo"
)

func OnUserUpdate(bot_session *discordgo.Session, userUpdate *discordgo.UserUpdate, db *sql.DB) {
	// Aktualisiere User in Datenbank
	user.UpsertUser(db, userUpdate.User, nil)
}
