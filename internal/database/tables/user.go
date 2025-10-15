package tables

import (
	"context"
	"database/sql"
	"time"
)

// User repräsentiert einen Discord-Benutzer in der Datenbank
type User struct {
	ID            string     `json:"id"`              // Discord User ID als Primary Key
	Name          string     `json:"name"`            // Discord Username
	GlobalName    string     `json:"global_name"`     // Discord Global Name
	DisplayName   string     `json:"display_name"`    // Discord Display Name
	Bot           bool       `json:"bot"`             // Ist der User ein Bot
	Avatar        string     `json:"avatar"`          // Discord Avatar Hash
	AvatarURL     string     `json:"avatar_url"`      // Vollständige Avatar URL
	Mention       string     `json:"mention"`         // Discord Mention String
	CreatedAt     time.Time  `json:"created_at"`      // Discord Account Erstellungsdatum
	Nick          string     `json:"nick"`            // Server Nickname
	JoinedAt      *time.Time `json:"joined_at"`       // Server Beitrittsdatum
	TopRole       string     `json:"top_role"`        // Höchste Rolle ID
	TimedOutUntil *time.Time `json:"timed_out_until"` // Timeout bis Datum
	PremiumSince  *time.Time `json:"premium_since"`   // Server Boost seit Datum
	UpdatedAt     time.Time  `json:"updated_at"`      // Letzte Aktualisierung
}

// CreateUserTable erstellt die User-Tabelle in der Datenbank
func CreateUserTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(32) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			global_name VARCHAR(255),
			display_name VARCHAR(255),
			bot BOOLEAN NOT NULL DEFAULT FALSE,
			avatar VARCHAR(255),
			avatar_url TEXT,
			mention VARCHAR(50),
			created_at TIMESTAMP NOT NULL,
			nick VARCHAR(255),
			joined_at TIMESTAMP,
			top_role VARCHAR(32),
			timed_out_until TIMESTAMP,
			premium_since TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);
		CREATE INDEX IF NOT EXISTS idx_users_joined_at ON users(joined_at);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertUser fügt einen neuen Discord-Benutzer in die Datenbank ein
func InsertUser(db *sql.DB, user *User) (*User, error) {
	query := `
		INSERT INTO users (id, name, global_name, display_name, bot, avatar, avatar_url, mention, created_at, nick, joined_at, top_role, timed_out_until, premium_since)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, name, global_name, display_name, bot, avatar, avatar_url, mention, created_at, nick, joined_at, top_role, timed_out_until, premium_since, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := &User{}
	err := db.QueryRowContext(ctx, query,
		user.ID,
		user.Name,
		user.GlobalName,
		user.DisplayName,
		user.Bot,
		user.Avatar,
		user.AvatarURL,
		user.Mention,
		user.CreatedAt,
		user.Nick,
		user.JoinedAt,
		user.TopRole,
		user.TimedOutUntil,
		user.PremiumSince,
	).Scan(
		&result.ID,
		&result.Name,
		&result.GlobalName,
		&result.DisplayName,
		&result.Bot,
		&result.Avatar,
		&result.AvatarURL,
		&result.Mention,
		&result.CreatedAt,
		&result.Nick,
		&result.JoinedAt,
		&result.TopRole,
		&result.TimedOutUntil,
		&result.PremiumSince,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetUserByID ruft einen Discord-Benutzer anhand der Discord ID ab
func GetUserByID(db *sql.DB, id string) (*User, error) {
	query := `
		SELECT id, name, global_name, display_name, bot, avatar, avatar_url, mention, created_at, nick, joined_at, top_role, timed_out_until, premium_since, updated_at
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := &User{}
	err := db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.GlobalName,
		&user.DisplayName,
		&user.Bot,
		&user.Avatar,
		&user.AvatarURL,
		&user.Mention,
		&user.CreatedAt,
		&user.Nick,
		&user.JoinedAt,
		&user.TopRole,
		&user.TimedOutUntil,
		&user.PremiumSince,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByUsername ruft einen Discord-Benutzer anhand des Benutzernamens ab
func GetUserByUsername(db *sql.DB, name string) (*User, error) {
	query := `
		SELECT id, name, global_name, display_name, bot, avatar, avatar_url, mention, created_at, nick, joined_at, top_role, timed_out_until, premium_since, updated_at
		FROM users
		WHERE name = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user := &User{}
	err := db.QueryRowContext(ctx, query, name).Scan(
		&user.ID,
		&user.Name,
		&user.GlobalName,
		&user.DisplayName,
		&user.Bot,
		&user.Avatar,
		&user.AvatarURL,
		&user.Mention,
		&user.CreatedAt,
		&user.Nick,
		&user.JoinedAt,
		&user.TopRole,
		&user.TimedOutUntil,
		&user.PremiumSince,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser aktualisiert einen Discord-Benutzer
func UpdateUser(db *sql.DB, user *User) (*User, error) {
	query := `
		UPDATE users
		SET name = $2, global_name = $3, display_name = $4, bot = $5, avatar = $6, avatar_url = $7, 
		    mention = $8, created_at = $9, nick = $10, joined_at = $11, top_role = $12, 
		    timed_out_until = $13, premium_since = $14, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, global_name, display_name, bot, avatar, avatar_url, mention, created_at, nick, joined_at, top_role, timed_out_until, premium_since, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := &User{}
	err := db.QueryRowContext(ctx, query,
		user.ID,
		user.Name,
		user.GlobalName,
		user.DisplayName,
		user.Bot,
		user.Avatar,
		user.AvatarURL,
		user.Mention,
		user.CreatedAt,
		user.Nick,
		user.JoinedAt,
		user.TopRole,
		user.TimedOutUntil,
		user.PremiumSince,
	).Scan(
		&result.ID,
		&result.Name,
		&result.GlobalName,
		&result.DisplayName,
		&result.Bot,
		&result.Avatar,
		&result.AvatarURL,
		&result.Mention,
		&result.CreatedAt,
		&result.Nick,
		&result.JoinedAt,
		&result.TopRole,
		&result.TimedOutUntil,
		&result.PremiumSince,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteUser löscht einen Discord-Benutzer
func DeleteUser(db *sql.DB, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, id)
	return err
}

// GetAllUsers ruft alle Discord-Benutzer ab
func GetAllUsers(db *sql.DB) ([]*User, error) {
	query := `
		SELECT id, name, global_name, display_name, bot, avatar, avatar_url, mention, created_at, nick, joined_at, top_role, timed_out_until, premium_since, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.GlobalName,
			&user.DisplayName,
			&user.Bot,
			&user.Avatar,
			&user.AvatarURL,
			&user.Mention,
			&user.CreatedAt,
			&user.Nick,
			&user.JoinedAt,
			&user.TopRole,
			&user.TimedOutUntil,
			&user.PremiumSince,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
