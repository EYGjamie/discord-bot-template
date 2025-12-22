package tables

import (
	"context"
	"database/sql"
	"time"
)

// UserMessageLog repräsentiert einen Nachrichten-Log-Eintrag in der Datenbank
type UserMessageLog struct {
	ID        int64     `json:"id"`
	AuthorID  string    `json:"author_id"`  // Discord User ID (Foreign Key zu users.id)
	ChannelID string    `json:"channel_id"` // Discord Channel ID
	CreatedAt time.Time `json:"created_at"` // Zeitpunkt der Nachricht
}

// CreateUserMessagesLogTable erstellt die user_messages_logs-Tabelle in der Datenbank
func CreateUserMessagesLogTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_messages_logs (
			id BIGSERIAL PRIMARY KEY,
			author_id VARCHAR(32) NOT NULL,
			channel_id VARCHAR(32) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_author
				FOREIGN KEY (author_id)
				REFERENCES users(id)
				ON DELETE CASCADE
		);
	`

	_, err := db.Exec(query)
	return err
}

// InsertUserMessageLog fügt einen neuen Nachrichten-Log-Eintrag in die Datenbank ein
func InsertUserMessageLog(db *sql.DB, authorID, channelID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO user_messages_logs (author_id, channel_id, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := db.ExecContext(ctx, query, authorID, channelID, time.Now())
	return err
}

// CountUserMessages zählt alle Nachrichten eines Users
func CountUserMessages(db *sql.DB, userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT COUNT(*)
		FROM user_messages_logs
		WHERE author_id = $1
	`

	var count int
	err := db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

// GetUserMessageLogsByAuthor ruft alle Nachrichten-Logs für einen bestimmten Autor ab
func GetUserMessageLogsByAuthor(db *sql.DB, authorID string) ([]UserMessageLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, author_id, channel_id, created_at
		FROM user_messages_logs
		WHERE author_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []UserMessageLog
	for rows.Next() {
		var log UserMessageLog
		if err := rows.Scan(&log.ID, &log.AuthorID, &log.ChannelID, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetUserMessageLogsByChannel ruft alle Nachrichten-Logs für einen bestimmten Channel ab
func GetUserMessageLogsByChannel(db *sql.DB, channelID string) ([]UserMessageLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, author_id, channel_id, created_at
		FROM user_messages_logs
		WHERE channel_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []UserMessageLog
	for rows.Next() {
		var log UserMessageLog
		if err := rows.Scan(&log.ID, &log.AuthorID, &log.ChannelID, &log.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
