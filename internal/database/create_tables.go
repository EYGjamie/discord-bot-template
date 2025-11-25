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

	log.Println("All database tables initialized successfully")
	return nil
}
