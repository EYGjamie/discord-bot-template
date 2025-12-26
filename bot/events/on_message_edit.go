package events

import (
	"discord-bot-template/bot/settings"
	"discord-bot-template/bot/services/moderation"

	"github.com/bwmarrin/discordgo"
)

func OnMessageUpdate(bot_session *discordgo.Session, message *discordgo.MessageUpdate, settingsManager *settings.Manager) {
	// Hole die vollständige Nachricht vor dem Edit (wenn gecacht)
	if message.BeforeUpdate == nil {
		// Keine Before-Daten verfügbar, überspringen
		return
	}

	// Logge die bearbeitete Nachricht
	moderation.LogMessageEdit(bot_session, settingsManager, message.BeforeUpdate, message.Message)
}
