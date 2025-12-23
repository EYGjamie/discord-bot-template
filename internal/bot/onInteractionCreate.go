package bot

import (
	"fmt"

	"discord-bot-template/internal/bot/commands"
	"discord-bot-template/internal/bot/handlers"
	"discord-bot-template/internal/shared/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// onInteractionCreate wird aufgerufen wenn ein Slash Command ausgeführt wird
func (bot *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Nur Application Commands verarbeiten
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
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
		case "warn":
			commands.HandleWarnCommand(s, i, bot.db)
		case "note":
			commands.HandleNoteCommand(s, i, bot.db)
		case "userinfo":
			commands.HandleUserInfoCommand(s, i, bot.db)
		case "setmodrole":
			commands.HandleSetModRoleCommand(s, i, bot.db)
		case "coinflip":
			commands.HandleCoinflipCommand(s, i, bot.db)
		case "setupcreatevoice":
			commands.HandleSetupCreateVoiceCommand(s, i, bot.db)
		}
	case discordgo.InteractionMessageComponent:
		// Handle Button/Select Menu Interactions
		customID := i.MessageComponentData().CustomID
		if len(customID) >= 3 && customID[:3] == "cv_" {
			// Create Voice Buttons und Select Menus
			if customID == "cv_kick_select" || customID == "cv_block_select" || customID == "cv_unblock_select" {
				handlers.HandleCreateVoiceSelects(s, i, bot.db)
			} else {
				handlers.HandleCreateVoiceButtons(s, i, bot.db)
			}
		}
	case discordgo.InteractionModalSubmit:
		// Handle Modal Submissions
		customID := i.ModalSubmitData().CustomID
		if customID == "cv_rename_modal" || customID == "cv_limit_modal" {
			handlers.HandleCreateVoiceModals(s, i, bot.db)
		}
	}
}
