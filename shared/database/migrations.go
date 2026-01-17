package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// CreateMigrationsTable erstellt die Tabelle zum Tracking der ausgeführten Migrationen
func CreateMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}
	return nil
}

// isMigrationExecuted prüft, ob eine Migration bereits ausgeführt wurde
func isMigrationExecuted(db *sql.DB, filename string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = $1", filename).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}
	return count > 0, nil
}

// recordMigration markiert eine Migration als ausgeführt
func recordMigration(db *sql.DB, filename string) error {
	_, err := db.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", filename)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}
	return nil
}

// RunMigrations führt alle ausstehenden Migrationen aus
func RunMigrations(db *sql.DB) error {
	log.Println("Starting database migrations...")

	// Erstelle Migrations-Tracking-Tabelle falls nicht vorhanden
	if err := CreateMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to initialize migrations table: %w", err)
	}

	// Lese alle Migrationsdateien aus dem embedded filesystem
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sammle alle .sql Dateien
	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// Sortiere Dateien alphabetisch (wichtig für die Reihenfolge)
	sort.Strings(migrationFiles)

	if len(migrationFiles) == 0 {
		log.Println("No migration files found")
		return nil
	}

	executedCount := 0
	skippedCount := 0

	// Führe jede Migration aus
	for _, filename := range migrationFiles {
		// Prüfe ob Migration bereits ausgeführt wurde
		executed, err := isMigrationExecuted(db, filename)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", filename, err)
		}

		if executed {
			log.Printf("Skipping migration %s (already executed)", filename)
			skippedCount++
			continue
		}

		// Lese Migrationsdatei
		content, err := migrationsFS.ReadFile(filepath.Join("migrations", filename))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		log.Printf("Executing migration: %s", filename)

		// Führe Migration aus
		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Markiere Migration als ausgeführt
		if err := recordMigration(db, filename); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		log.Printf("Successfully executed migration: %s", filename)
		executedCount++
	}

	log.Printf("Migrations completed: %d executed, %d skipped", executedCount, skippedCount)
	return nil
}
