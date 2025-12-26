package tables

import (
	"context"
	"database/sql"
	"time"
)

// ModerationType repräsentiert den Typ der Moderation
type ModerationType string

const (
	ModerationTypeWarn ModerationType = "WARN"
	ModerationTypeNote ModerationType = "NOTE"
)

// UserModerationLog repräsentiert einen Warn oder Note Eintrag
type UserModerationLog struct {
	ID          int64          `json:"id"`
	GuildID     string         `json:"guild_id"`     // Guild ID
	UserID      string         `json:"user_id"`      // Betroffener User
	ModeratorID string         `json:"moderator_id"` // Moderator der die Aktion ausgeführt hat
	Type        ModerationType `json:"type"`         // WARN oder NOTE
	Reason      string         `json:"reason"`       // Grund/Beschreibung
	CreatedAt   time.Time      `json:"created_at"`   // Zeitstempel
}

// CreateUserModerationLogsTable erstellt die user_moderation_logs-Tabelle
func CreateUserModerationLogsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_moderation_logs (
			id BIGSERIAL PRIMARY KEY,
			guild_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			moderator_id VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			reason TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_mod_logs_guild_user ON user_moderation_logs(guild_id, user_id);
		CREATE INDEX IF NOT EXISTS idx_mod_logs_type ON user_moderation_logs(type);
		CREATE INDEX IF NOT EXISTS idx_mod_logs_created_at ON user_moderation_logs(created_at DESC);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertModerationLog fügt einen neuen Warn oder Note hinzu
func InsertModerationLog(db *sql.DB, guildID, userID, moderatorID string, modType ModerationType, reason string) error {
	query := `
		INSERT INTO user_moderation_logs (guild_id, user_id, moderator_id, type, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, guildID, userID, moderatorID, modType, reason, time.Now())
	return err
}

// GetModerationLogsByUser holt alle Warns und Notes für einen User
func GetModerationLogsByUser(db *sql.DB, guildID, userID string) ([]UserModerationLog, error) {
	query := `
		SELECT id, guild_id, user_id, moderator_id, type, reason, created_at
		FROM user_moderation_logs
		WHERE guild_id = $1 AND user_id = $2
		ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []UserModerationLog
	for rows.Next() {
		var log UserModerationLog
		err := rows.Scan(&log.ID, &log.GuildID, &log.UserID, &log.ModeratorID, &log.Type, &log.Reason, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// CountModerationLogsByType zählt Warns oder Notes für einen User
func CountModerationLogsByType(db *sql.DB, guildID, userID string, modType ModerationType) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_moderation_logs
		WHERE guild_id = $1 AND user_id = $2 AND type = $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := db.QueryRowContext(ctx, query, guildID, userID, modType).Scan(&count)
	return count, err
}

// DeleteModerationLog löscht einen Warn oder Note
func DeleteModerationLog(db *sql.DB, id int64) error {
	query := `DELETE FROM user_moderation_logs WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, id)
	return err
}
