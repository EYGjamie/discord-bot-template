package tables

import (
	"context"
	"database/sql"
	"time"
)

// MatchCategory repräsentiert eine Match-Kategorie/Tag mit Farbe
type MatchCategory struct {
	ID        int64     `json:"id"`
	GuildID   string    `json:"guild_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateMatchCategoriesTable erstellt die match_categories-Tabelle
func CreateMatchCategoriesTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS match_categories (
			id BIGSERIAL PRIMARY KEY,
			guild_id VARCHAR(255) NOT NULL,
			name VARCHAR(100) NOT NULL,
			color VARCHAR(7) NOT NULL,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_match_categories_guild_id ON match_categories(guild_id);
		CREATE INDEX IF NOT EXISTS idx_match_categories_sort_order ON match_categories(guild_id, sort_order);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertMatchCategory fügt eine neue Match-Kategorie ein
func InsertMatchCategory(db *sql.DB, category *MatchCategory) error {
	query := `
		INSERT INTO match_categories (guild_id, name, color, sort_order)
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

// UpdateMatchCategory aktualisiert eine Match-Kategorie
func UpdateMatchCategory(db *sql.DB, category *MatchCategory) error {
	query := `
		UPDATE match_categories 
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

// DeleteMatchCategory löscht eine Match-Kategorie
func DeleteMatchCategory(db *sql.DB, categoryID int64) error {
	query := `DELETE FROM match_categories WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, categoryID)
	return err
}

// GetMatchCategoriesByGuild holt alle Match-Kategorien für eine Guild
func GetMatchCategoriesByGuild(db *sql.DB, guildID string) ([]MatchCategory, error) {
	query := `
		SELECT id, guild_id, name, color, sort_order, created_at, updated_at
		FROM match_categories
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

	categories := []MatchCategory{}
	for rows.Next() {
		var cat MatchCategory
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

// InitializeDefaultMatchCategories erstellt Standard-Kategorien für eine Guild, falls noch keine existieren
func InitializeDefaultMatchCategories(db *sql.DB, guildID string) error {
	// Prüfe, ob bereits Kategorien existieren
	existing, err := GetMatchCategoriesByGuild(db, guildID)
	if err != nil {
		return err
	}

	// Wenn bereits Kategorien existieren, nichts tun
	if len(existing) > 0 {
		return nil
	}

	// Standard-Kategorien definieren (andere Farben als Events)
	defaultCategories := []MatchCategory{
		{GuildID: guildID, Name: "Competitive", Color: "#E53935", SortOrder: 1},
		{GuildID: guildID, Name: "Casual", Color: "#43A047", SortOrder: 2},
		{GuildID: guildID, Name: "Training", Color: "#FB8C00", SortOrder: 3},
		{GuildID: guildID, Name: "Tournament", Color: "#8E24AA", SortOrder: 4},
		{GuildID: guildID, Name: "Scrim", Color: "#3949AB", SortOrder: 5},
		{GuildID: guildID, Name: "League", Color: "#00ACC1", SortOrder: 6},
		{GuildID: guildID, Name: "Friendly", Color: "#FDD835", SortOrder: 7},
		{GuildID: guildID, Name: "Championship", Color: "#D81B60", SortOrder: 8},
	}

	// Kategorien einfügen
	for i := range defaultCategories {
		if err := InsertMatchCategory(db, &defaultCategories[i]); err != nil {
			return err
		}
	}

	return nil
}
