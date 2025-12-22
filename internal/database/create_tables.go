package database

import (
	"database/sql"
	"log"

	"discord-bot-template/internal/database/tables"
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

	log.Println("All database tables initialized successfully")
	return nil
}
