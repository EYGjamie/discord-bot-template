package bot

import (
	"ovgu/internal/bot/events"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) registerHandlers() {
	bot.session.AddHandler(bot.onReady)
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		events.OnMessageCreate(s, m, bot.db)
	})
}
