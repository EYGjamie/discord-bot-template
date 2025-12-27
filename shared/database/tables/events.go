package tables

import (
	"context"
	"database/sql"
	"time"
)

// Event repräsentiert einen Kalendereintrag
type Event struct {
	ID          int64     `json:"id"`
	GuildID     string    `json:"guild_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   string    `json:"start_date"` // YYYY-MM-DD
	EndDate     string    `json:"end_date"`   // YYYY-MM-DD (kann gleich wie start_date sein)
	StartTime   string    `json:"start_time"` // HH:MM
	EndTime     string    `json:"end_time"`   // HH:MM
	Color       string    `json:"color"`      // Hex color code (z.B. #4285F4)
	Location    string    `json:"location"`   // Optional: Ort des Events
	Guests      string    `json:"guests"`     // Comma-separated list of guest names
	CreatedBy   string    `json:"created_by"` // User ID des Erstellers
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateEventsTable erstellt die events-Tabelle
func CreateEventsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS events (
			id BIGSERIAL PRIMARY KEY,
			guild_id VARCHAR(255) NOT NULL,
			title VARCHAR(500) NOT NULL,
			description TEXT,
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			start_time TIME,
			end_time TIME,
			color VARCHAR(7) DEFAULT '#4285F4',
			location VARCHAR(500),
			guests TEXT,
			created_by VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_events_guild_id ON events(guild_id);
		CREATE INDEX IF NOT EXISTS idx_events_start_date ON events(start_date);
		CREATE INDEX IF NOT EXISTS idx_events_end_date ON events(end_date);
		CREATE INDEX IF NOT EXISTS idx_events_created_by ON events(created_by);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertEvent fügt ein neues Event in die Datenbank ein
func InsertEvent(db *sql.DB, event *Event) error {
	query := `
		INSERT INTO events (guild_id, title, description, start_date, end_date, start_time, end_time, color, location, guests, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		event.GuildID,
		event.Title,
		event.Description,
		event.StartDate,
		event.EndDate,
		event.StartTime,
		event.EndTime,
		event.Color,
		event.Location,
		event.Guests,
		event.CreatedBy,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	return err
}

// UpdateEvent aktualisiert ein bestehendes Event
func UpdateEvent(db *sql.DB, event *Event) error {
	query := `
		UPDATE events 
		SET title = $1, description = $2, start_date = $3, end_date = $4, start_time = $5, 
		    end_time = $6, color = $7, location = $8, guests = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $10
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		event.Title,
		event.Description,
		event.StartDate,
		event.EndDate,
		event.StartTime,
		event.EndTime,
		event.Color,
		event.Location,
		event.Guests,
		event.ID,
	).Scan(&event.UpdatedAt)

	return err
}

// DeleteEvent löscht ein Event aus der Datenbank
func DeleteEvent(db *sql.DB, eventID int64) error {
	query := `DELETE FROM events WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, eventID)
	return err
}
