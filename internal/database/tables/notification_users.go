package tables

import (
	"context"
	"database/sql"
	"time"
)

// NotificationType definiert den Typ der Notification
type NotificationType string

const (
	NotificationTypeInfo  NotificationType = "info"
	NotificationTypeError NotificationType = "error"
	NotificationTypeBoth  NotificationType = "both"
)

// NotificationUser repräsentiert einen User der Notifications erhält
type NotificationUser struct {
	UserID           string           `json:"user_id"`           // Discord User ID
	GuildID          string           `json:"guild_id"`          // Discord Guild ID
	NotificationType NotificationType `json:"notification_type"` // info, error, both
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// CreateNotificationUsersTable erstellt die Notification Users Tabelle
func CreateNotificationUsersTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS notification_users (
			user_id VARCHAR(32) NOT NULL,
			guild_id VARCHAR(32) NOT NULL,
			notification_type VARCHAR(10) NOT NULL CHECK (notification_type IN ('info', 'error', 'both')),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, guild_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_notification_users_guild_id ON notification_users(guild_id);
		CREATE INDEX IF NOT EXISTS idx_notification_users_type ON notification_users(notification_type);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// AddNotificationUser fügt einen User zur Notification-Liste hinzu
func AddNotificationUser(db *sql.DB, userID, guildID string, notificationType NotificationType) (*NotificationUser, error) {
	query := `
		INSERT INTO notification_users (user_id, guild_id, notification_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, guild_id) 
		DO UPDATE SET 
			notification_type = $3,
			updated_at = CURRENT_TIMESTAMP
		RETURNING user_id, guild_id, notification_type, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := &NotificationUser{}
	err := db.QueryRowContext(ctx, query, userID, guildID, notificationType).Scan(
		&user.UserID,
		&user.GuildID,
		&user.NotificationType,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// RemoveNotificationUser entfernt einen User von der Notification-Liste
func RemoveNotificationUser(db *sql.DB, userID, guildID string) error {
	query := `
		DELETE FROM notification_users
		WHERE user_id = $1 AND guild_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, userID, guildID)
	return err
}

// GetNotificationUsers gibt alle Notification-User für eine Guild zurück
func GetNotificationUsers(db *sql.DB, guildID string, notificationType NotificationType) ([]*NotificationUser, error) {
	var query string
	var args []interface{}

	if notificationType == "" {
		// Alle User
		query = `
			SELECT user_id, guild_id, notification_type, created_at, updated_at
			FROM notification_users
			WHERE guild_id = $1
			ORDER BY created_at DESC
		`
		args = []interface{}{guildID}
	} else {
		// Nur User mit spezifischem Typ oder "both"
		query = `
			SELECT user_id, guild_id, notification_type, created_at, updated_at
			FROM notification_users
			WHERE guild_id = $1 AND (notification_type = $2 OR notification_type = 'both')
			ORDER BY created_at DESC
		`
		args = []interface{}{guildID, notificationType}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*NotificationUser
	for rows.Next() {
		user := &NotificationUser{}
		err := rows.Scan(
			&user.UserID,
			&user.GuildID,
			&user.NotificationType,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// IsNotificationUser prüft ob ein User Notifications erhält
func IsNotificationUser(db *sql.DB, userID, guildID string) (bool, NotificationType, error) {
	query := `
		SELECT notification_type
		FROM notification_users
		WHERE user_id = $1 AND guild_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var notificationType NotificationType
	err := db.QueryRowContext(ctx, query, userID, guildID).Scan(&notificationType)

	if err == sql.ErrNoRows {
		return false, "", nil
	}

	if err != nil {
		return false, "", err
	}

	return true, notificationType, nil
}
