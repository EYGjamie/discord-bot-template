package bot

import (
	"discord-bot-template/internal/bot/events"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) registerHandlers() {
	bot.session.AddHandler(bot.onReady)
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		events.OnMessageCreate(s, m, bot.db)
	})
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
		events.OnGuildMemberAdd(s, m, bot.db)
	})
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberRemove) {
		events.OnGuildMemberRemove(s, m, bot.db)
	})
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
		events.OnGuildMemberUpdate(s, m, bot.db)
	})
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageUpdate) {
		events.OnMessageUpdate(s, m, bot.settings)
	})
	bot.session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageDelete) {
		events.OnMessageDelete(s, m, bot.settings)
	})
}
