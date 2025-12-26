package tables

import (
	"context"
	"database/sql"
	"time"
)

// LogLevel repräsentiert das Level eines Logs
type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelError LogLevel = "ERROR"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelDebug LogLevel = "DEBUG"
)

// Log repräsentiert einen Log-Eintrag in der Datenbank
type Log struct {
	ID         int64     `json:"id"`
	GuildID    string    `json:"guild_id"`    // Optional: Für Guild-spezifische Logs
	Level      LogLevel  `json:"level"`       // INFO, ERROR, WARN, DEBUG
	Title      string    `json:"title"`       // Kurzbeschreibung
	Message    string    `json:"message"`     // Detaillierte Log-Nachricht
	StackTrace string    `json:"stack_trace"` // Optional: Stack Trace für Errors
	Source     string    `json:"source"`      // Quelle des Logs (z.B. "bot.commands", "bot.events")
	CreatedAt  time.Time `json:"created_at"`
}

// CreateLogsTable erstellt die Logs-Tabelle
func CreateLogsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS logs (
			id BIGSERIAL PRIMARY KEY,
			guild_id VARCHAR(255),
			level VARCHAR(20) NOT NULL,
			title VARCHAR(500) NOT NULL,
			message TEXT NOT NULL,
			stack_trace TEXT,
			source VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
		CREATE INDEX IF NOT EXISTS idx_logs_guild_id ON logs(guild_id);
		CREATE INDEX IF NOT EXISTS idx_logs_created_at ON logs(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_logs_source ON logs(source);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertLog fügt einen Log-Eintrag in die Datenbank ein
func InsertLog(db *sql.DB, log *Log) error {
	query := `
		INSERT INTO logs (guild_id, level, title, message, stack_trace, source)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		log.GuildID,
		log.Level,
		log.Title,
		log.Message,
		log.StackTrace,
		log.Source,
	).Scan(&log.ID, &log.CreatedAt)

	return err
}

// GetLogs ruft Logs aus der Datenbank ab mit optionalen Filtern
func GetLogs(db *sql.DB, guildID string, level LogLevel, limit int) ([]Log, error) {
	query := `
		SELECT id, guild_id, level, title, message, stack_trace, source, created_at
		FROM logs
		WHERE ($1 = '' OR guild_id = $1)
		AND ($2 = '' OR level = $2)
		ORDER BY created_at DESC
		LIMIT $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID, level, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var log Log
		err := rows.Scan(
			&log.ID,
			&log.GuildID,
			&log.Level,
			&log.Title,
			&log.Message,
			&log.StackTrace,
			&log.Source,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// DeleteOldLogs löscht Logs, die älter als die angegebene Dauer sind
func DeleteOldLogs(db *sql.DB, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM logs
		WHERE created_at < $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cutoffTime := time.Now().Add(-olderThan)
	result, err := db.ExecContext(ctx, query, cutoffTime)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
