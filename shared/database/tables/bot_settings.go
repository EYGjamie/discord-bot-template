package tables

import (
	"context"
	"database/sql"
	"time"
)

// BotSetting repräsentiert eine Bot-Einstellung in der Datenbank
type BotSetting struct {
	Key       string    `json:"key"`        // Eindeutiger Schlüssel für die Einstellung
	Value     string    `json:"value"`      // Wert als String (JSON für komplexe Werte)
	Type      string    `json:"type"`       // Datentyp: string, bool, int, json
	Enabled   bool      `json:"enabled"`    // Ob die Einstellung aktiv ist
	UpdatedAt time.Time `json:"updated_at"` // Letzte Aktualisierung
	CreatedAt time.Time `json:"created_at"` // Erstellungsdatum
}

// CreateBotSettingsTable erstellt die Bot-Einstellungen-Tabelle
func CreateBotSettingsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS bot_settings (
			key VARCHAR(255) PRIMARY KEY,
			value TEXT NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'string',
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_bot_settings_enabled ON bot_settings(enabled);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// UpsertBotSetting fügt eine Einstellung ein oder aktualisiert sie
func UpsertBotSetting(db *sql.DB, setting *BotSetting) (*BotSetting, error) {
	query := `
		INSERT INTO bot_settings (key, value, type, enabled)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			type = EXCLUDED.type,
			enabled = EXCLUDED.enabled,
			updated_at = CURRENT_TIMESTAMP
		RETURNING key, value, type, enabled, updated_at, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := &BotSetting{}
	err := db.QueryRowContext(ctx, query,
		setting.Key,
		setting.Value,
		setting.Type,
		setting.Enabled,
	).Scan(
		&result.Key,
		&result.Value,
		&result.Type,
		&result.Enabled,
		&result.UpdatedAt,
		&result.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetBotSetting ruft eine Einstellung anhand des Schlüssels ab
func GetBotSetting(db *sql.DB, key string) (*BotSetting, error) {
	query := `
		SELECT key, value, type, enabled, updated_at, created_at
		FROM bot_settings
		WHERE key = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	setting := &BotSetting{}
	err := db.QueryRowContext(ctx, query, key).Scan(
		&setting.Key,
		&setting.Value,
		&setting.Type,
		&setting.Enabled,
		&setting.UpdatedAt,
		&setting.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return setting, nil
}

// GetAllBotSettings ruft alle Einstellungen ab
func GetAllBotSettings(db *sql.DB) ([]*BotSetting, error) {
	query := `
		SELECT key, value, type, enabled, updated_at, created_at
		FROM bot_settings
		ORDER BY key
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*BotSetting
	for rows.Next() {
		setting := &BotSetting{}
		err := rows.Scan(
			&setting.Key,
			&setting.Value,
			&setting.Type,
			&setting.Enabled,
			&setting.UpdatedAt,
			&setting.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateBotSettingEnabled aktiviert oder deaktiviert eine Einstellung
func UpdateBotSettingEnabled(db *sql.DB, key string, enabled bool) error {
	query := `
		UPDATE bot_settings
		SET enabled = $2, updated_at = CURRENT_TIMESTAMP
		WHERE key = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, key, enabled)
	return err
}

// DeleteBotSetting löscht eine Einstellung
func DeleteBotSetting(db *sql.DB, key string) error {
	query := `DELETE FROM bot_settings WHERE key = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, key)
	return err
}
