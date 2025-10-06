package bot

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session
	token   string
}

func New() (*Bot, error) {
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		session: session,
		token:   token,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	b.registerHandlers()

	if err := b.session.Open(); err != nil {
		return err
	}

	log.Println("Bot successfully connected to Discord")
	return nil
}

func (b *Bot) Stop() error {
	return b.session.Close()
}


