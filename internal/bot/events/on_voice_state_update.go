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

	// Handle Create Voice Channel Logic
	// Wenn User einen Channel betritt (ChannelID ist gesetzt)
	if voiceState.ChannelID != "" {
		// Prüfe ob der User von einem anderen Channel kommt oder neu joint
		oldChannelID := ""
		if voiceState.BeforeUpdate != nil {
			oldChannelID = voiceState.BeforeUpdate.ChannelID
		}

		// Nur wenn User den Channel gewechselt hat oder neu beigetreten ist
		if oldChannelID != voiceState.ChannelID {
			// Prüfe ob es ein Create-Voice-Channel ist
			voice.HandleCreateVoiceJoin(bot_session, db, voiceState)
		}
	}

	// Wenn User einen Channel verlässt
	if voiceState.BeforeUpdate != nil && voiceState.BeforeUpdate.ChannelID != "" {
		oldChannelID := voiceState.BeforeUpdate.ChannelID
		// Prüfe ob der alte Channel ein temporärer Channel ist und leer ist
		voice.HandleTemporaryVoiceLeave(bot_session, db, oldChannelID, voiceState.GuildID)
	}
}
