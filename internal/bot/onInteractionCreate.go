package bot

import (
	"fmt"

	"discord-bot-template/internal/bot/commands"
	"discord-bot-template/internal/shared/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// onInteractionCreate wird aufgerufen wenn ein Slash Command ausgeführt wird
func (bot *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Nur Application Commands verarbeiten
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	logger := logging.NewLogger(bot.db, s, i.GuildID, "bot.interaction")
	cmdName := i.ApplicationCommandData().Name
	userTag := "Unknown"
	if i.Member != nil && i.Member.User != nil {
		userTag = i.Member.User.Username
	}
	logger.LogInfo("Command Executed", fmt.Sprintf("/%s by %s", cmdName, userTag), false)

	// Prüfe welcher Command ausgeführt wurde
	switch cmdName {
	case "moderation":
		commands.HandleModerationCommand(s, i, bot.settings)
	}
}
