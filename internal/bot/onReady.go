package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func (bot *Bot) onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

	// Registriere Commands nachdem der Bot bereit ist
	bot.registerCommands()

	// Initialisiere Invite-Cache für alle Guilds
	bot.inviteCache.InitializeForAllGuilds(s)
}
