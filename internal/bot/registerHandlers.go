package bot

func (bot *Bot) registerHandlers() {
	bot.session.AddHandler(bot.onReady)
}