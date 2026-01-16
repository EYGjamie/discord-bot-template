package database

import (
	"database/sql"
	"log"

	"discord-bot-template/shared/database/tables"
)

// InitializeTables erstellt alle benötigten Datenbanktabellen
// Die Reihenfolge ist wichtig wegen Foreign Key Constraints:
// 1. users und roles (keine Abhängigkeiten)
// 2. user_roles (abhängig von users und roles)
// 3. user_messages_logs (abhängig von users)
func InitializeTables(db *sql.DB) error {
	// Erstelle Users-Tabelle
	if err := tables.CreateUserTable(db); err != nil {
		log.Printf("Error creating users table: %v", err)
		return err
	}
	log.Println("Users table initialized successfully")

	// Erstelle Roles-Tabelle
	if err := tables.CreateRoleTable(db); err != nil {
		log.Printf("Error creating roles table: %v", err)
		return err
	}
	log.Println("Roles table initialized successfully")

	// Erstelle Channels-Tabelle
	if err := tables.CreateChannelTable(db); err != nil {
		log.Printf("Error creating channels table: %v", err)
		return err
	}
	log.Println("Channels table initialized successfully")

	// Erstelle User-Roles Junction-Tabelle
	if err := tables.CreateUserRoleTable(db); err != nil {
		log.Printf("Error creating user_roles table: %v", err)
		return err
	}
	log.Println("User-Roles table initialized successfully")

	// Erstelle User-Messages-Logs-Tabelle
	if err := tables.CreateUserMessagesLogTable(db); err != nil {
		log.Printf("Error creating user_messages_logs table: %v", err)
		return err
	}
	log.Println("User-Messages-Logs table initialized successfully")

	// Erstelle User-Voice-Logs-Tabelle
	if err := tables.CreateUserVoiceLogsTable(db); err != nil {
		log.Printf("Error creating user_voice_logs table: %v", err)
		return err
	}
	log.Println("User-Voice-Logs table initialized successfully")

	// Erstelle Bot-Settings-Tabelle
	if err := tables.CreateBotSettingsTable(db); err != nil {
		log.Printf("Error creating bot_settings table: %v", err)
		return err
	}
	log.Println("Bot-Settings table initialized successfully")

	// Erstelle User-Joins-Tabelle
	if err := tables.CreateUserJoinsTable(db); err != nil {
		log.Printf("Error creating user_joins table: %v", err)
		return err
	}
	log.Println("User-Joins table initialized successfully")

	// Erstelle User-Leaves-Tabelle
	if err := tables.CreateUserLeavesTable(db); err != nil {
		log.Printf("Error creating user_leaves table: %v", err)
		return err
	}
	log.Println("User-Leaves table initialized successfully")

	// Erstelle Notification-Users-Tabelle
	if err := tables.CreateNotificationUsersTable(db); err != nil {
		log.Printf("Error creating notification_users table: %v", err)
		return err
	}
	log.Println("Notification-Users table initialized successfully")

	// Erstelle Logs-Tabelle
	if err := tables.CreateLogsTable(db); err != nil {
		log.Printf("Error creating logs table: %v", err)
		return err
	}
	log.Println("Logs table initialized successfully")

	// Erstelle Role-Audit-Logs-Tabelle
	if err := tables.CreateRoleAuditLogsTable(db); err != nil {
		log.Printf("Error creating role_audit_logs table: %v", err)
		return err
	}
	log.Println("Role-Audit-Logs table initialized successfully")

	// Erstelle User-Moderation-Logs-Tabelle
	if err := tables.CreateUserModerationLogsTable(db); err != nil {
		log.Printf("Error creating user_moderation_logs table: %v", err)
		return err
	}
	log.Println("User-Moderation-Logs table initialized successfully")

	// Erstelle Create-Voice-Settings-Tabelle
	if err := tables.CreateCreateVoiceSettingsTable(db); err != nil {
		log.Printf("Error creating create_voice_settings table: %v", err)
		return err
	}
	log.Println("Create-Voice-Settings table initialized successfully")

	// Erstelle Temporary-Voice-Channels-Tabelle
	if err := tables.CreateTemporaryVoiceChannelsTable(db); err != nil {
		log.Printf("Error creating temporary_voice_channels table: %v", err)
		return err
	}
	log.Println("Temporary-Voice-Channels table initialized successfully")

	// Erstelle Channel-Purge-Settings-Tabelle
	if err := tables.CreateChannelPurgeSettingsTable(db); err != nil {
		log.Printf("Error creating channel_purge_settings table: %v", err)
		return err
	}
	log.Println("Channel-Purge-Settings table initialized successfully")

	// Erstelle WebApp-Audit-Logs-Tabelle
	if err := tables.CreateWebAppAuditLogsTable(db); err != nil {
		log.Printf("Error creating webapp_audit_logs table: %v", err)
		return err
	}
	log.Println("WebApp-Audit-Logs table initialized successfully")

	// Erstelle Events-Tabelle
	if err := tables.CreateEventsTable(db); err != nil {
		log.Printf("Error creating events table: %v", err)
		return err
	}
	log.Println("Events table initialized successfully")

	// Erstelle Event Categories-Tabelle
	if err := tables.CreateEventCategoriesTable(db); err != nil {
		log.Printf("Error creating event_categories table: %v", err)
		return err
	}
	log.Println("Event Categories table initialized successfully")

	// Erstelle Matches-Tabelle
	if err := tables.CreateMatchesTable(db); err != nil {
		log.Printf("Error creating matches table: %v", err)
		return err
	}
	log.Println("Matches table initialized successfully")

	// Erstelle Match Categories-Tabelle
	if err := tables.CreateMatchCategoriesTable(db); err != nil {
		log.Printf("Error creating match_categories table: %v", err)
		return err
	}
	log.Println("Match Categories table initialized successfully")

	// Erstelle Discord-Statistics-Tabelle
	if err := tables.CreateDiscordStatisticsTable(db); err != nil {
		log.Printf("Error creating discord_statistics table: %v", err)
		return err
	}
	log.Println("Discord-Statistics table initialized successfully")

	// Erstelle Boards-Tabelle (für Task-Management)
	if err := tables.CreateBoardsTable(db); err != nil {
		log.Printf("Error creating boards table: %v", err)
		return err
	}
	log.Println("Boards table initialized successfully")

	// Erstelle Tasks-Tabellen (für Task-Management)
	if err := tables.CreateTasksTable(db); err != nil {
		log.Printf("Error creating tasks tables: %v", err)
		return err
	}
	log.Println("Tasks tables initialized successfully")

	log.Println("All database tables initialized successfully")
	return nil
}

// InitializeDefaultData fügt Standard-Daten in die Datenbank ein, falls noch nicht vorhanden
func InitializeDefaultData(db *sql.DB, guildID string) error {
	// Initialisiere Standard-Event-Kategorien
	if err := tables.InitializeDefaultCategories(db, guildID); err != nil {
		log.Printf("Error initializing default event categories: %v", err)
		return err
	}
	log.Println("Default event categories initialized successfully")

	// Initialisiere Standard-Match-Kategorien
	if err := tables.InitializeDefaultMatchCategories(db, guildID); err != nil {
		log.Printf("Error initializing default match categories: %v", err)
		return err
	}
	log.Println("Default match categories initialized successfully")

	return nil
}
