package tables

import (
	"context"
	"database/sql"
	"time"
)

// EventCategory repräsentiert eine Event-Kategorie/Tag mit Farbe
type EventCategory struct {
	ID        int64     `json:"id"`
	GuildID   string    `json:"guild_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateEventCategoriesTable erstellt die event_categories-Tabelle
func CreateEventCategoriesTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS event_categories (
			id BIGSERIAL PRIMARY KEY,
			guild_id VARCHAR(255) NOT NULL,
			name VARCHAR(100) NOT NULL,
			color VARCHAR(7) NOT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_event_categories_guild_id ON event_categories(guild_id);
		CREATE INDEX IF NOT EXISTS idx_event_categories_sort_order ON event_categories(guild_id, sort_order);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertEventCategory fügt eine neue Event-Kategorie ein
func InsertEventCategory(db *sql.DB, category *EventCategory) error {
	query := `
		INSERT INTO event_categories (guild_id, name, color, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		category.GuildID,
		category.Name,
		category.Color,
		category.SortOrder,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)

	return err
}

// UpdateEventCategory aktualisiert eine Event-Kategorie
func UpdateEventCategory(db *sql.DB, category *EventCategory) error {
	query := `
		UPDATE event_categories 
		SET name = $1, color = $2, sort_order = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		category.Name,
		category.Color,
		category.SortOrder,
		category.ID,
	).Scan(&category.UpdatedAt)

	return err
}

// DeleteEventCategory löscht eine Event-Kategorie
func DeleteEventCategory(db *sql.DB, categoryID int64) error {
	query := `DELETE FROM event_categories WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, categoryID)
	return err
}

// GetEventCategoriesByGuild holt alle Event-Kategorien für eine Guild
func GetEventCategoriesByGuild(db *sql.DB, guildID string) ([]EventCategory, error) {
	query := `
		SELECT id, guild_id, name, color, sort_order, created_at, updated_at
		FROM event_categories
		WHERE guild_id = $1
		ORDER BY sort_order ASC, name ASC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []EventCategory{}
	for rows.Next() {
		var cat EventCategory
		err := rows.Scan(
			&cat.ID, &cat.GuildID, &cat.Name, &cat.Color,
			&cat.SortOrder, &cat.CreatedAt, &cat.UpdatedAt,
		)
		if err != nil {
			continue
		}
		categories = append(categories, cat)
	}

	return categories, nil
}
