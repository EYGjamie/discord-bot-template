package bot

import (
	"fmt"
	"log"

	"discord-bot-template/bot/utils/logging"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) onReady(s *discordgo.Session, event *discordgo.Ready) {
	logger := logging.NewLogger(bot.db, s, "", "bot.ready")
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	logger.LogInfo("Bot Ready", fmt.Sprintf("Logged in as %v#%v", s.State.User.Username, s.State.User.Discriminator), false)

	// Registriere Commands nachdem der Bot bereit ist
	bot.registerCommands()

	// Initialisiere Invite-Cache für alle Guilds
	bot.inviteCache.InitializeForAllGuilds(s)
	logger.LogInfo("Invite Cache Initialized", fmt.Sprintf("Initialized invite cache for %d guilds", len(s.State.Guilds)), true)
}
