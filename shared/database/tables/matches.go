package tables

import (
	"context"
	"database/sql"
	"time"
)

// Match repräsentiert einen Match-Kalendereintrag
type Match struct {
	ID          int64     `json:"id"`
	GuildID     string    `json:"guild_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   string    `json:"start_date"` // YYYY-MM-DD
	EndDate     string    `json:"end_date"`   // YYYY-MM-DD (kann gleich wie start_date sein)
	StartTime   string    `json:"start_time"` // HH:MM
	EndTime     string    `json:"end_time"`   // HH:MM
	IsAllDay    bool      `json:"is_all_day"` // Ganztägiges Match
	Color       string    `json:"color"`      // Hex color code (z.B. #4285F4)
	Location    string    `json:"location"`   // Optional: Ort des Matches
	Guests      string    `json:"guests"`     // Comma-separated list of guest names
	CreatedBy   string    `json:"created_by"` // User ID des Erstellers
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateMatchesTable erstellt die matches-Tabelle
func CreateMatchesTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS matches (
			id BIGSERIAL PRIMARY KEY,
			guild_id VARCHAR(255) NOT NULL,
			title VARCHAR(500) NOT NULL,
			description TEXT,
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			start_time TIME,
			end_time TIME,
			is_all_day BOOLEAN DEFAULT FALSE,
			color VARCHAR(7) DEFAULT '#4285F4',
			location VARCHAR(500),
			guests TEXT,
			created_by VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_matches_guild_id ON matches(guild_id);
		CREATE INDEX IF NOT EXISTS idx_matches_start_date ON matches(start_date);
		CREATE INDEX IF NOT EXISTS idx_matches_end_date ON matches(end_date);
		CREATE INDEX IF NOT EXISTS idx_matches_created_by ON matches(created_by);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertMatch fügt ein neues Match in die Datenbank ein
func InsertMatch(db *sql.DB, match *Match) error {
	query := `
		INSERT INTO matches (guild_id, title, description, start_date, end_date, start_time, end_time, is_all_day, color, location, guests, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		match.GuildID,
		match.Title,
		match.Description,
		match.StartDate,
		match.EndDate,
		match.StartTime,
		match.EndTime,
		match.IsAllDay,
		match.Color,
		match.Location,
		match.Guests,
		match.CreatedBy,
	).Scan(&match.ID, &match.CreatedAt, &match.UpdatedAt)

	return err
}

// UpdateMatch aktualisiert ein bestehendes Match
func UpdateMatch(db *sql.DB, match *Match) error {
	query := `
		UPDATE matches 
		SET title = $1, description = $2, start_date = $3, end_date = $4, start_time = $5, 
		    end_time = $6, is_all_day = $7, color = $8, location = $9, guests = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		match.Title,
		match.Description,
		match.StartDate,
		match.EndDate,
		match.StartTime,
		match.EndTime,
		match.IsAllDay,
		match.Color,
		match.Location,
		match.Guests,
		match.ID,
	).Scan(&match.UpdatedAt)

	return err
}

// DeleteMatch löscht ein Match
func DeleteMatch(db *sql.DB, matchID int64) error {
	query := `DELETE FROM matches WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, matchID)
	return err
}
