package bot

import (
	"discord-bot-template/internal/bot/commands"

	"github.com/bwmarrin/discordgo"
)

// onInteractionCreate wird aufgerufen wenn ein Slash Command ausgeführt wird
func (bot *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Nur Application Commands verarbeiten
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	// Prüfe welcher Command ausgeführt wurde
	switch i.ApplicationCommandData().Name {
	case "moderation":
		commands.HandleModerationCommand(s, i, bot.settings)
	}
}
