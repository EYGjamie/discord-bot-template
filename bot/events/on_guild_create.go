package events

import (
	"database/sql"

	guildinit "discord-bot-template/bot/init"

	"github.com/bwmarrin/discordgo"
)

// OnGuildCreate wird ausgelöst, wenn der Bot einem Server beitritt
// oder wenn der Bot startet und Zugriff auf einen Server hat
func OnGuildCreate(bot_session *discordgo.Session, guild *discordgo.GuildCreate, db *sql.DB) {
	// Führe vollständige Synchronisierung durch
	err := guildinit.SyncGuildOnJoin(bot_session, guild.ID, db)
	if err != nil {
		// Fehler werden bereits in der Sync-Funktion geloggt
		return
	}
}
