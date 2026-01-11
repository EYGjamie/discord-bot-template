package tables

import (
	"context"
	"database/sql"
	"time"
)

// ChannelPurgeSetting repräsentiert eine Channel-Purge-Einstellung in der Datenbank
type ChannelPurgeSetting struct {
	ID           int64        `json:"id"`
	GuildID      string       `json:"guild_id"`
	ChannelID    string       `json:"channel_id"`
	ScheduleTime string       `json:"schedule_time"` // Format: "HH:MM" (24-Stunden-Format)
	Enabled      bool         `json:"enabled"`
	LastRun      sql.NullTime `json:"last_run"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// CreateChannelPurgeSettingsTable erstellt die Tabelle für Channel-Purge-Einstellungen
func CreateChannelPurgeSettingsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS channel_purge_settings (
			id SERIAL PRIMARY KEY,
			guild_id VARCHAR(255) NOT NULL,
			channel_id VARCHAR(255) NOT NULL,
			schedule_time VARCHAR(5) NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT TRUE,
			last_run TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(guild_id, channel_id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_channel_purge_guild ON channel_purge_settings(guild_id);
		CREATE INDEX IF NOT EXISTS idx_channel_purge_enabled ON channel_purge_settings(enabled);
		CREATE INDEX IF NOT EXISTS idx_channel_purge_schedule ON channel_purge_settings(schedule_time);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertChannelPurgeSetting fügt eine neue Channel-Purge-Einstellung ein
func InsertChannelPurgeSetting(db *sql.DB, setting *ChannelPurgeSetting) (*ChannelPurgeSetting, error) {
	query := `
		INSERT INTO channel_purge_settings (guild_id, channel_id, schedule_time, enabled)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (guild_id, channel_id) DO UPDATE SET
			schedule_time = EXCLUDED.schedule_time,
			enabled = EXCLUDED.enabled,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, guild_id, channel_id, schedule_time, enabled, last_run, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result ChannelPurgeSetting
	err := db.QueryRowContext(ctx, query,
		setting.GuildID,
		setting.ChannelID,
		setting.ScheduleTime,
		setting.Enabled,
	).Scan(
		&result.ID,
		&result.GuildID,
		&result.ChannelID,
		&result.ScheduleTime,
		&result.Enabled,
		&result.LastRun,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetChannelPurgeSetting holt eine Channel-Purge-Einstellung
func GetChannelPurgeSetting(db *sql.DB, guildID, channelID string) (*ChannelPurgeSetting, error) {
	query := `
		SELECT id, guild_id, channel_id, schedule_time, enabled, last_run, created_at, updated_at
		FROM channel_purge_settings
		WHERE guild_id = $1 AND channel_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var setting ChannelPurgeSetting
	err := db.QueryRowContext(ctx, query, guildID, channelID).Scan(
		&setting.ID,
		&setting.GuildID,
		&setting.ChannelID,
		&setting.ScheduleTime,
		&setting.Enabled,
		&setting.LastRun,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &setting, nil
}

// GetAllEnabledPurgeSettings holt alle aktivierten Purge-Einstellungen
func GetAllEnabledPurgeSettings(db *sql.DB) ([]*ChannelPurgeSetting, error) {
	query := `
		SELECT id, guild_id, channel_id, schedule_time, enabled, last_run, created_at, updated_at
		FROM channel_purge_settings
		WHERE enabled = TRUE
		ORDER BY schedule_time ASC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*ChannelPurgeSetting
	for rows.Next() {
		var setting ChannelPurgeSetting
		err := rows.Scan(
			&setting.ID,
			&setting.GuildID,
			&setting.ChannelID,
			&setting.ScheduleTime,
			&setting.Enabled,
			&setting.LastRun,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, &setting)
	}

	return settings, rows.Err()
}

// GetGuildPurgeSettings holt alle Purge-Einstellungen für eine Guild
func GetGuildPurgeSettings(db *sql.DB, guildID string) ([]*ChannelPurgeSetting, error) {
	query := `
		SELECT id, guild_id, channel_id, schedule_time, enabled, last_run, created_at, updated_at
		FROM channel_purge_settings
		WHERE guild_id = $1
		ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*ChannelPurgeSetting
	for rows.Next() {
		var setting ChannelPurgeSetting
		err := rows.Scan(
			&setting.ID,
			&setting.GuildID,
			&setting.ChannelID,
			&setting.ScheduleTime,
			&setting.Enabled,
			&setting.LastRun,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, &setting)
	}

	return settings, rows.Err()
}

// UpdatePurgeSettingEnabled aktiviert/deaktiviert eine Purge-Einstellung
func UpdatePurgeSettingEnabled(db *sql.DB, guildID, channelID string, enabled bool) error {
	query := `
		UPDATE channel_purge_settings
		SET enabled = $1, updated_at = CURRENT_TIMESTAMP
		WHERE guild_id = $2 AND channel_id = $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, enabled, guildID, channelID)
	return err
}

// UpdatePurgeLastRun aktualisiert die letzte Ausführungszeit
func UpdatePurgeLastRun(db *sql.DB, guildID, channelID string) error {
	query := `
		UPDATE channel_purge_settings
		SET last_run = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE guild_id = $1 AND channel_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, guildID, channelID)
	return err
}

// DeleteChannelPurgeSetting löscht eine Channel-Purge-Einstellung
func DeleteChannelPurgeSetting(db *sql.DB, guildID, channelID string) error {
	query := `
		DELETE FROM channel_purge_settings
		WHERE guild_id = $1 AND channel_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, guildID, channelID)
	return err
}
