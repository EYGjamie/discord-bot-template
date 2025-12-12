package events

import (
	"database/sql"

	"discord-bot-template/internal/shared/services/events/user"
	"discord-bot-template/internal/shared/services/events/voice"

	"github.com/bwmarrin/discordgo"
)

func OnVoiceStateUpdate(bot_session *discordgo.Session, voiceState *discordgo.VoiceStateUpdate, db *sql.DB) {
	// Update/Insert User in Datenbank
	if voiceState.Member != nil && voiceState.Member.User != nil {
		user.UpsertUser(db, voiceState.Member.User, voiceState.Member)
	}

	// Logge den Voice State Change in der Datenbank
	voice.LogVoiceStateUpdate(db, voiceState)
}
