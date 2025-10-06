package bot

func (b *Bot) registerHandlers() {
	b.session.AddHandler(b.onReady)
}