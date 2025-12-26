package tables

import (
	"context"
	"database/sql"
	"time"
)

// LeaveReason definiert den Grund für das Verlassen
type LeaveReason string

const (
	LeaveReasonLeave LeaveReason = "leave" // User ist selbst gegangen
	LeaveReasonKick  LeaveReason = "kick"  // User wurde gekickt
	LeaveReasonBan   LeaveReason = "ban"   // User wurde gebannt
)

// UserLeave repräsentiert einen Leave-Event eines Users
type UserLeave struct {
	ID       int         `json:"id"`        // Auto-Increment ID
	UserID   string      `json:"user_id"`   // Discord User ID
	GuildID  string      `json:"guild_id"`  // Discord Guild ID
	Reason   LeaveReason `json:"reason"`    // Grund: leave, kick, ban
	LeftAt   time.Time   `json:"left_at"`   // Zeitpunkt des Leaves
	BannedBy *string     `json:"banned_by"` // User ID der den Ban/Kick ausgelöst hat (nullable)
}

// CreateUserLeavesTable erstellt die User Leaves Tabelle
func CreateUserLeavesTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_leaves (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(32) NOT NULL,
			guild_id VARCHAR(32) NOT NULL,
			reason VARCHAR(10) NOT NULL CHECK (reason IN ('leave', 'kick', 'ban')),
			left_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			banned_by VARCHAR(32),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_user_leaves_user_id ON user_leaves(user_id);
		CREATE INDEX IF NOT EXISTS idx_user_leaves_guild_id ON user_leaves(guild_id);
		CREATE INDEX IF NOT EXISTS idx_user_leaves_reason ON user_leaves(reason);
		CREATE INDEX IF NOT EXISTS idx_user_leaves_left_at ON user_leaves(left_at);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// LogUserLeave speichert einen Leave-Event in der Datenbank
func LogUserLeave(db *sql.DB, userID, guildID string, reason LeaveReason, bannedBy *string) (*UserLeave, error) {
	query := `
		INSERT INTO user_leaves (user_id, guild_id, reason, left_at, banned_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, guild_id, reason, left_at, banned_by
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	leave := &UserLeave{}
	err := db.QueryRowContext(
		ctx,
		query,
		userID,
		guildID,
		reason,
		time.Now(),
		bannedBy,
	).Scan(
		&leave.ID,
		&leave.UserID,
		&leave.GuildID,
		&leave.Reason,
		&leave.LeftAt,
		&leave.BannedBy,
	)

	if err != nil {
		return nil, err
	}

	return leave, nil
}

// GetUserLeaves gibt alle Leaves eines Users zurück
func GetUserLeaves(db *sql.DB, userID string) ([]*UserLeave, error) {
	query := `
		SELECT id, user_id, guild_id, reason, left_at, banned_by
		FROM user_leaves
		WHERE user_id = $1
		ORDER BY left_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaves []*UserLeave
	for rows.Next() {
		leave := &UserLeave{}
		err := rows.Scan(
			&leave.ID,
			&leave.UserID,
			&leave.GuildID,
			&leave.Reason,
			&leave.LeftAt,
			&leave.BannedBy,
		)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leave)
	}

	return leaves, rows.Err()
}

// GetLeavesByReason gibt alle Leaves nach Grund zurück
func GetLeavesByReason(db *sql.DB, guildID string, reason LeaveReason) ([]*UserLeave, error) {
	query := `
		SELECT id, user_id, guild_id, reason, left_at, banned_by
		FROM user_leaves
		WHERE guild_id = $1 AND reason = $2
		ORDER BY left_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID, reason)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaves []*UserLeave
	for rows.Next() {
		leave := &UserLeave{}
		err := rows.Scan(
			&leave.ID,
			&leave.UserID,
			&leave.GuildID,
			&leave.Reason,
			&leave.LeftAt,
			&leave.BannedBy,
		)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leave)
	}

	return leaves, rows.Err()
}

// CountLeavesByReason zählt Leaves nach Grund
func CountLeavesByReason(db *sql.DB, guildID string, reason LeaveReason) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM user_leaves
		WHERE guild_id = $1 AND reason = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err := db.QueryRowContext(ctx, query, guildID, reason).Scan(&count)
	return count, err
}
