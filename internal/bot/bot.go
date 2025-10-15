package bot

import (
	"context"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session
	token   string
}

func New() (*Bot, error) {
	// Load Bot Token from Environment Variable
	var token string
	if os.Getenv("APP_ENV") == "production" {
		token = os.Getenv("PROD_DISCORD_BOT_TOKEN")
		if token == "" {
			log.Fatal("PROD_DISCORD_BOT_TOKEN environment variable is required")
			os.Exit(1)
		}
	} else {
		token = os.Getenv("DEV_DISCORD_BOT_TOKEN")
		if token == "" {
			log.Fatal("DEV_DISCORD_BOT_TOKEN environment variable is required")
			os.Exit(1)
		}
	}

	// Creation Discord-Session
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	// Register ALL Bot-Intents
	session.Identify.Intents = discordgo.IntentsAll

	return &Bot{
		session: session,
		token:   token,
	}, nil
}

func (bot *Bot) Start(ctx context.Context) error {
	bot.registerHandlers()

	if err := bot.session.Open(); err != nil {
		return err
	}

	log.Println("Bot successfully connected to Discord")
	return nil
}

func (bot *Bot) Stop() error {
	return bot.session.Close()
}
