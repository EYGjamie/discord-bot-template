package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"discord-bot-template/bot/api"
	"discord-bot-template/bot/services"
	"discord-bot-template/bot/settings"
	"discord-bot-template/bot/utils/logging"
	"discord-bot-template/shared/database"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session        *discordgo.Session
	token          string
	db             *sql.DB
	settings       *settings.Manager
	inviteCache    *InviteCache
	purgeScheduler *services.PurgeScheduler
	apiServer      *api.Server
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

	// Run database migrations for updates
	if err := database.RunMigrations(db); err != nil {
		logger.LogError("Database Migrations Failed", fmt.Sprintf("Failed to run migrations: %v", err), "")
		log.Printf("Warning: Failed to run database migrations: %v", err)
	} else {
		logger.LogInfo("Database Migrations Complete", "All migrations executed successfully", false)
	}

	// Initialize default data (e.g., event categories)
	guildID := os.Getenv("GUILD_ID")
	if guildID != "" {
		if err := database.InitializeDefaultData(db, guildID); err != nil {
			logger.LogError("Default Data Initialization Failed", fmt.Sprintf("Failed to initialize default data: %v", err), "")
			log.Printf("Warning: Failed to initialize default data: %v", err)
		} else {
			logger.LogInfo("Default Data Initialized", "Default event categories created successfully", false)
		}
	}

	// Initialize settings manager
	settingsManager := settings.NewManager(db)

	// Initialize invite cache
	inviteCache := NewInviteCache()

	// Initialize API server
	apiServer := api.NewServer(session, db)

	return &Bot{
		session:        session,
		token:          token,
		db:             db,
		settings:       settingsManager,
		inviteCache:    inviteCache,
		purgeScheduler: nil, // Wird nach dem Login initialisiert
		apiServer:      apiServer,
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

	// Start API server in goroutine
	if bot.apiServer != nil {
		go func() {
			log.Printf("Starting Bot API server on %s", bot.apiServer.GetAddr())
			if err := bot.apiServer.Start(); err != nil && err != http.ErrServerClosed {
				logger.LogError("Bot API Server Error", fmt.Sprintf("Failed to start Bot API server: %v", err), "")
				log.Printf("Bot API server error: %v", err)
			}
		}()
		log.Println("Bot API Server started")
	}

	return nil
}

func (bot *Bot) Stop() error {
	logger := logging.NewLogger(bot.db, bot.session, "", "bot.shutdown")

	// Stop purge scheduler
	if bot.purgeScheduler != nil {
		bot.purgeScheduler.Stop()
		log.Println("Purge Scheduler stopped")
	}

	// Stop API server
	if bot.apiServer != nil {
		log.Println("Shutting down Bot API server...")
		if err := bot.apiServer.Shutdown(5 * time.Second); err != nil {
			logger.LogError("Bot API Server Shutdown Error", fmt.Sprintf("Error shutting down Bot API server: %v", err), "")
			log.Printf("Error shutting down Bot API server: %v", err)
		} else {
			log.Println("Bot API Server stopped")
		}
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
