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
	IsAllDay    bool      `json:"is_all_day"` // Ganztägiges Event
	Color       string    `json:"color"`      // Hex color code (z.B. #4285F4)
	Location    string    `json:"location"`   // Optional: Ort des Events
	Tags        string    `json:"tags"`       // JSON array of label names
	CreatedBy   string    `json:"created_by"` // User ID des Erstellers
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EventGuest repräsentiert einen eingeladenen Gast
type EventGuest struct {
	ID              int64     `json:"id"`
	EventID         int64     `json:"event_id"`
	UserID          string    `json:"user_id"`
	UserName        string    `json:"user_name"`
	UserDisplayName string    `json:"user_display_name"`
	UserAvatar      string    `json:"user_avatar"`
	RSVPStatus      string    `json:"rsvp_status"` // pending, accepted, declined
	RSVPAt          time.Time `json:"rsvp_at"`
	InvitedAt       time.Time `json:"invited_at"`
}

// EventLabel repräsentiert ein Label für Events
type EventLabel struct {
	ID        int64     `json:"id"`
	GuildID   string    `json:"guild_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

// EventChecklistItem repräsentiert einen Checklist-Eintrag
type EventChecklistItem struct {
	ID          int64     `json:"id"`
	EventID     int64     `json:"event_id"`
	Text        string    `json:"text"`
	IsCompleted bool      `json:"is_completed"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateEventsTable erstellt die events-Tabelle und zugehörige Tabellen
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
			is_all_day BOOLEAN DEFAULT FALSE,
			color VARCHAR(7) DEFAULT '#4285F4',
			location VARCHAR(500),
			tags TEXT DEFAULT '[]',
			created_by VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_events_guild_id ON events(guild_id);
		CREATE INDEX IF NOT EXISTS idx_events_start_date ON events(start_date);
		CREATE INDEX IF NOT EXISTS idx_events_end_date ON events(end_date);
		CREATE INDEX IF NOT EXISTS idx_events_created_by ON events(created_by);

		-- Event Guests table for tracking invitations and RSVP status
		CREATE TABLE IF NOT EXISTS event_guests (
			id BIGSERIAL PRIMARY KEY,
			event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			user_id VARCHAR(255) NOT NULL,
			user_name VARCHAR(255) NOT NULL,
			user_display_name VARCHAR(255),
			user_avatar TEXT,
			rsvp_status VARCHAR(20) DEFAULT 'pending',
			rsvp_at TIMESTAMP,
			invited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT unique_event_guest UNIQUE (event_id, user_id)
		);

		CREATE INDEX IF NOT EXISTS idx_event_guests_event_id ON event_guests(event_id);
		CREATE INDEX IF NOT EXISTS idx_event_guests_user_id ON event_guests(user_id);

		-- Event Labels table for categorizing events
		CREATE TABLE IF NOT EXISTS event_labels (
			id BIGSERIAL PRIMARY KEY,
			guild_id VARCHAR(255) NOT NULL,
			name VARCHAR(50) NOT NULL,
			color VARCHAR(20) DEFAULT 'blue',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT unique_event_label UNIQUE (guild_id, name)
		);

		CREATE INDEX IF NOT EXISTS idx_event_labels_guild_id ON event_labels(guild_id);

		-- Event Checklist Items table
		CREATE TABLE IF NOT EXISTS event_checklist_items (
			id BIGSERIAL PRIMARY KEY,
			event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
			text VARCHAR(500) NOT NULL,
			is_completed BOOLEAN DEFAULT FALSE,
			position INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_event_checklist_items_event_id ON event_checklist_items(event_id);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertEvent fügt ein neues Event in die Datenbank ein
func InsertEvent(db *sql.DB, event *Event) error {
	query := `
		INSERT INTO events (guild_id, title, description, start_date, end_date, start_time, end_time, is_all_day, color, location, tags, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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
		event.IsAllDay,
		event.Color,
		event.Location,
		event.Tags,
		event.CreatedBy,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)

	return err
}

// UpdateEvent aktualisiert ein bestehendes Event
func UpdateEvent(db *sql.DB, event *Event) error {
	query := `
		UPDATE events 
		SET title = $1, description = $2, start_date = $3, end_date = $4, start_time = $5, 
		    end_time = $6, is_all_day = $7, color = $8, location = $9, tags = $10, updated_at = CURRENT_TIMESTAMP
		WHERE id = $11
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
		event.IsAllDay,
		event.Color,
		event.Location,
		event.Tags,
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
