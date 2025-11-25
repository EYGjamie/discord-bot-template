package events

import (
	"discord-bot-template/internal/bot/settings"
	"discord-bot-template/internal/shared/services/moderation"

	"github.com/bwmarrin/discordgo"
)

func OnMessageDelete(bot_session *discordgo.Session, message *discordgo.MessageDelete, settingsManager *settings.Manager) {
	// Hole die gecachte Nachricht wenn verfügbar
	if message.BeforeDelete == nil {
		// Keine Before-Daten verfügbar, überspringen
		return
	}

	// Logge die gelöschte Nachricht
	moderation.LogMessageDelete(bot_session, settingsManager, message.BeforeDelete)
}
