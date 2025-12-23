package bot

import (
	"fmt"
	"log"

	"discord-bot-template/internal/bot/commands"
	"discord-bot-template/internal/shared/utils/logging"
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

	logger := logging.NewLogger(bot.db, bot.session, "", "bot.commands.register")
	log.Printf("Registriere Commands für %d Guild(s)...", len(guilds))
	logger.LogInfo("Command Registration Started", fmt.Sprintf("Registering commands for %d guilds", len(guilds)), false)

	successCount := 0
	for _, guild := range guilds {
		// Registriere Commands für jede Guild
		if err := commands.SetupModerationCommand(bot.session, guild.ID); err != nil {
			log.Printf("Fehler beim Registrieren des /moderation Commands für Guild %s: %v", guild.ID, err)
			logger.LogError("Command Registration Failed", fmt.Sprintf("Failed to register /moderation for guild %s: %v", guild.Name, err), "")
		} else {
			log.Printf("✓ Command /moderation registriert für Guild: %s (%s)", guild.Name, guild.ID)
		}

		if err := commands.SetupWarnCommand(bot.session, guild.ID); err != nil {
			log.Printf("Fehler beim Registrieren des /warn Commands für Guild %s: %v", guild.ID, err)
			logger.LogError("Command Registration Failed", fmt.Sprintf("Failed to register /warn for guild %s: %v", guild.Name, err), "")
		} else {
			log.Printf("✓ Command /warn registriert für Guild: %s (%s)", guild.Name, guild.ID)
		}

		if err := commands.SetupNoteCommand(bot.session, guild.ID); err != nil {
			log.Printf("Fehler beim Registrieren des /note Commands für Guild %s: %v", guild.ID, err)
			logger.LogError("Command Registration Failed", fmt.Sprintf("Failed to register /note for guild %s: %v", guild.Name, err), "")
		} else {
			log.Printf("✓ Command /note registriert für Guild: %s (%s)", guild.Name, guild.ID)
		}

		if err := commands.SetupUserInfoCommand(bot.session, guild.ID); err != nil {
			log.Printf("Fehler beim Registrieren des /userinfo Commands für Guild %s: %v", guild.ID, err)
			logger.LogError("Command Registration Failed", fmt.Sprintf("Failed to register /userinfo for guild %s: %v", guild.Name, err), "")
		} else {
			log.Printf("✓ Command /userinfo registriert für Guild: %s (%s)", guild.Name, guild.ID)
		}

		if err := commands.SetupSetModRoleCommand(bot.session, guild.ID); err != nil {
			log.Printf("Fehler beim Registrieren des /setmodrole Commands für Guild %s: %v", guild.ID, err)
			logger.LogError("Command Registration Failed", fmt.Sprintf("Failed to register /setmodrole for guild %s: %v", guild.Name, err), "")
		} else {
			log.Printf("✓ Command /setmodrole registriert für Guild: %s (%s)", guild.Name, guild.ID)
		}

		if err := commands.SetupCoinflipCommand(bot.session, guild.ID); err != nil {
			log.Printf("Fehler beim Registrieren des /coinflip Commands für Guild %s: %v", guild.ID, err)
			logger.LogError("Command Registration Failed", fmt.Sprintf("Failed to register /coinflip for guild %s: %v", guild.Name, err), "")
		} else {
			log.Printf("✓ Command /coinflip registriert für Guild: %s (%s)", guild.Name, guild.ID)
		}

		if err := commands.SetupCreateVoiceCommand(bot.session, guild.ID); err != nil {
			log.Printf("Fehler beim Registrieren des /setupcreatevoice Commands für Guild %s: %v", guild.ID, err)
			logger.LogError("Command Registration Failed", fmt.Sprintf("Failed to register /setupcreatevoice for guild %s: %v", guild.Name, err), "")
		} else {
			log.Printf("✓ Command /setupcreatevoice registriert für Guild: %s (%s)", guild.Name, guild.ID)
		}

		successCount++
	}

	log.Println("Command-Registrierung abgeschlossen")
	logger.LogInfo("Command Registration Completed", fmt.Sprintf("Successfully registered commands in %d/%d guilds", successCount, len(guilds)), false)
}

// removeCommands entfernt alle registrierten Commands (z.B. beim Shutdown)
func (bot *Bot) removeCommands() {
	if bot.session.State.User == nil {
		return
	}

	logger := logging.NewLogger(bot.db, bot.session, "", "bot.commands.cleanup")
	guilds := bot.session.State.Guilds
	for _, guild := range guilds {
		commands, err := bot.session.ApplicationCommands(bot.session.State.User.ID, guild.ID)
		if err != nil {
			log.Printf("Fehler beim Abrufen der Commands für Guild %s: %v", guild.ID, err)
			logger.LogError("Command Fetch Failed", fmt.Sprintf("Failed to fetch commands for guild %s: %v", guild.Name, err), "")
			continue
		}

		for _, cmd := range commands {
			err := bot.session.ApplicationCommandDelete(bot.session.State.User.ID, guild.ID, cmd.ID)
			if err != nil {
				log.Printf("Fehler beim Löschen des Commands %s: %v", cmd.Name, err)
				logger.LogError("Command Deletion Failed", fmt.Sprintf("Failed to delete command %s from guild %s: %v", cmd.Name, guild.Name, err), "")
			} else {
				log.Printf("Command %s gelöscht von Guild %s", cmd.Name, guild.Name)
			}
		}
	}
	logger.LogInfo("Commands Removed", "All commands successfully removed from guilds", false)
}
