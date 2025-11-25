package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"discord-bot-template/internal/bot"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	botInstance, err := bot.New()
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	if err := botInstance.Start(ctx); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}

	log.Println("Bot is running. Press CTRL+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down bot...")
	if err := botInstance.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
