package bot

import (
	"log"

	"discord-bot-template/internal/bot/commands"
)

// registerCommands registriert alle Slash Commands für alle Guilds
func (bot *Bot) registerCommands() {
	// Warte bis Bot Ready ist
	if bot.session.State.User == nil {
		log.Println("Bot ist noch nicht bereit, Commands werden beim Ready-Event registriert")
		return
	}

	// Hole alle Guilds in denen der Bot ist
	guilds := bot.session.State.Guilds
	if len(guilds) == 0 {
		log.Println("Bot ist in keinen Guilds, Commands werden später registriert")
		return
	}

	log.Printf("Registriere Commands für %d Guild(s)...", len(guilds))

	for _, guild := range guilds {
		// Registriere Commands für jede Guild
		if err := commands.SetupModerationCommand(bot.session, guild.ID); err != nil {
			log.Printf("Fehler beim Registrieren des /moderation Commands für Guild %s: %v", guild.ID, err)
		} else {
			log.Printf("✓ Command /moderation registriert für Guild: %s (%s)", guild.Name, guild.ID)
		}
	}

	log.Println("Command-Registrierung abgeschlossen")
}

// removeCommands entfernt alle registrierten Commands (z.B. beim Shutdown)
func (bot *Bot) removeCommands() {
	if bot.session.State.User == nil {
		return
	}

	guilds := bot.session.State.Guilds
	for _, guild := range guilds {
		commands, err := bot.session.ApplicationCommands(bot.session.State.User.ID, guild.ID)
		if err != nil {
			log.Printf("Fehler beim Abrufen der Commands für Guild %s: %v", guild.ID, err)
			continue
		}

		for _, cmd := range commands {
			err := bot.session.ApplicationCommandDelete(bot.session.State.User.ID, guild.ID, cmd.ID)
			if err != nil {
				log.Printf("Fehler beim Löschen des Commands %s: %v", cmd.Name, err)
			} else {
				log.Printf("Command %s gelöscht von Guild %s", cmd.Name, guild.Name)
			}
		}
	}
}
