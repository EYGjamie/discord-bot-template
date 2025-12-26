package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"discord-bot-template/bot/settings"
	"discord-bot-template/shared/database"
	"discord-bot-template/bot/services"
	"discord-bot-template/bot/utils/logging"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session        *discordgo.Session
	token          string
	db             *sql.DB
	settings       *settings.Manager
	inviteCache    *InviteCache
	purgeScheduler *services.PurgeScheduler
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

	// Initialize database connection
	dbService := database.New()

	// Initialize database tables
	db := dbService.DB()
	logger := logging.NewLogger(db, nil, "", "bot.startup")
	if err := database.InitializeTables(db); err != nil {
		logger.LogError("Database Initialization Failed", fmt.Sprintf("Failed to initialize tables: %v", err), "")
		log.Printf("Warning: Failed to initialize database tables: %v", err)
	} else {
		logger.LogInfo("Database Initialized", "All database tables created successfully", false)
	}

	// Initialize settings manager
	settingsManager := settings.NewManager(db)

	// Initialize invite cache
	inviteCache := NewInviteCache()

	return &Bot{
		session:        session,
		token:          token,
		db:             db,
		settings:       settingsManager,
		inviteCache:    inviteCache,
		purgeScheduler: nil, // Wird nach dem Login initialisiert
	}, nil
}

func (bot *Bot) Start(ctx context.Context) error {
	logger := logging.NewLogger(bot.db, bot.session, "", "bot.startup")
	bot.registerHandlers()

	if err := bot.session.Open(); err != nil {
		logger.LogError("Bot Connection Failed", fmt.Sprintf("Failed to connect to Discord: %v", err), "")
		return err
	}

	logger.LogInfo("Bot Started", "Bot successfully connected to Discord", false)
	log.Println("Bot successfully connected to Discord")

	// Initialize and start purge scheduler
	bot.purgeScheduler = services.NewPurgeScheduler(bot.session, bot.db)
	bot.purgeScheduler.Start()
	log.Println("Purge Scheduler started")

	return nil
}

func (bot *Bot) Stop() error {
	logger := logging.NewLogger(bot.db, bot.session, "", "bot.shutdown")

	// Stop purge scheduler
	if bot.purgeScheduler != nil {
		bot.purgeScheduler.Stop()
		log.Println("Purge Scheduler stopped")
	}

	bot.removeCommands()
	err := bot.session.Close()
	if err != nil {
		logger.LogError("Bot Shutdown Error", fmt.Sprintf("Error during shutdown: %v", err), "")
	} else {
		logger.LogInfo("Bot Stopped", "Bot successfully disconnected from Discord", false)
	}
	return err
}
