package tables

import (
	"context"
	"database/sql"
	"time"
)

// UserJoin repräsentiert einen Join-Event eines Users
type UserJoin struct {
	ID         int       `json:"id"`          // Auto-Increment ID
	UserID     string    `json:"user_id"`     // Discord User ID
	GuildID    string    `json:"guild_id"`    // Discord Guild ID
	InviterID  *string   `json:"inviter_id"`  // Discord User ID des Inviters (nullable)
	InviteCode *string   `json:"invite_code"` // Der verwendete Invite-Code (nullable)
	JoinedAt   time.Time `json:"joined_at"`   // Zeitpunkt des Joins
}

// CreateUserJoinsTable erstellt die User Joins Tabelle
func CreateUserJoinsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_joins (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(32) NOT NULL,
			guild_id VARCHAR(32) NOT NULL,
			inviter_id VARCHAR(32),
			invite_code VARCHAR(50),
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_user_joins_user_id ON user_joins(user_id);
		CREATE INDEX IF NOT EXISTS idx_user_joins_guild_id ON user_joins(guild_id);
		CREATE INDEX IF NOT EXISTS idx_user_joins_inviter_id ON user_joins(inviter_id);
		CREATE INDEX IF NOT EXISTS idx_user_joins_joined_at ON user_joins(joined_at);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// LogUserJoin speichert einen Join-Event in der Datenbank
func LogUserJoin(db *sql.DB, userID, guildID string, inviterID, inviteCode *string) (*UserJoin, error) {
	query := `
		INSERT INTO user_joins (user_id, guild_id, inviter_id, invite_code, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, guild_id, inviter_id, invite_code, joined_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	join := &UserJoin{}
	err := db.QueryRowContext(
		ctx,
		query,
		userID,
		guildID,
		inviterID,
		inviteCode,
		time.Now(),
	).Scan(
		&join.ID,
		&join.UserID,
		&join.GuildID,
		&join.InviterID,
		&join.InviteCode,
		&join.JoinedAt,
	)

	if err != nil {
		return nil, err
	}

	return join, nil
}

// GetUserJoins gibt alle Joins eines Users zurück
func GetUserJoins(db *sql.DB, userID string) ([]*UserJoin, error) {
	query := `
		SELECT id, user_id, guild_id, inviter_id, invite_code, joined_at
		FROM user_joins
		WHERE user_id = $1
		ORDER BY joined_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var joins []*UserJoin
	for rows.Next() {
		join := &UserJoin{}
		err := rows.Scan(
			&join.ID,
			&join.UserID,
			&join.GuildID,
			&join.InviterID,
			&join.InviteCode,
			&join.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		joins = append(joins, join)
	}

	return joins, rows.Err()
}

// GetJoinsByInviter gibt alle Joins zurück die durch einen bestimmten Inviter entstanden sind
func GetJoinsByInviter(db *sql.DB, inviterID string) ([]*UserJoin, error) {
	query := `
		SELECT id, user_id, guild_id, inviter_id, invite_code, joined_at
		FROM user_joins
		WHERE inviter_id = $1
		ORDER BY joined_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, inviterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var joins []*UserJoin
	for rows.Next() {
		join := &UserJoin{}
		err := rows.Scan(
			&join.ID,
			&join.UserID,
			&join.GuildID,
			&join.InviterID,
			&join.InviteCode,
			&join.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		joins = append(joins, join)
	}

	return joins, rows.Err()
}

// CountInvitesByUser zählt wie viele User durch einen bestimmten Inviter eingeladen wurden
func CountInvitesByUser(db *sql.DB, inviterID, guildID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_joins
		WHERE inviter_id = $1 AND guild_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err := db.QueryRowContext(ctx, query, inviterID, guildID).Scan(&count)
	return count, err
}
