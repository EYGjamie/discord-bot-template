package tables

import (
	"context"
	"database/sql"
	"time"
)

// CreateVoiceSetting repräsentiert die Konfiguration eines Create-Voice-Channels
type CreateVoiceSetting struct {
	ID               int64     `json:"id"`
	GuildID          string    `json:"guild_id"`
	ChannelID        string    `json:"channel_id"`         // Der "Create Voice" Channel
	DefaultUserLimit int       `json:"default_user_limit"` // Standard Max-User-Anzahl
	ControlChannelID string    `json:"control_channel_id"` // Text-Channel für Control Panel
	ControlMessageID string    `json:"control_message_id"` // Message-ID des Control Panels
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TemporaryVoiceChannel repräsentiert einen temporär erstellten Voice Channel
type TemporaryVoiceChannel struct {
	ID            int64     `json:"id"`
	GuildID       string    `json:"guild_id"`
	ChannelID     string    `json:"channel_id"`      // Der temporäre Channel
	OwnerID       string    `json:"owner_id"`        // Der Ersteller/Owner
	CreateVoiceID int64     `json:"create_voice_id"` // Referenz zu CreateVoiceSetting
	BlockedUsers  string    `json:"blocked_users"`   // JSON-Array von User-IDs
	CreatedAt     time.Time `json:"created_at"`
}

// CreateCreateVoiceSettingsTable erstellt die Tabelle für Create-Voice-Settings
func CreateCreateVoiceSettingsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS create_voice_settings (
			id SERIAL PRIMARY KEY,
			guild_id VARCHAR(20) NOT NULL,
			channel_id VARCHAR(20) NOT NULL UNIQUE,
			default_user_limit INTEGER NOT NULL DEFAULT 0,
			control_channel_id VARCHAR(20),
			control_message_id VARCHAR(20),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_create_voice_guild ON create_voice_settings(guild_id);
		CREATE INDEX IF NOT EXISTS idx_create_voice_channel ON create_voice_settings(channel_id);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// CreateTemporaryVoiceChannelsTable erstellt die Tabelle für temporäre Channels
func CreateTemporaryVoiceChannelsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS temporary_voice_channels (
			id SERIAL PRIMARY KEY,
			guild_id VARCHAR(20) NOT NULL,
			channel_id VARCHAR(20) NOT NULL UNIQUE,
			owner_id VARCHAR(20) NOT NULL,
			create_voice_id INTEGER NOT NULL REFERENCES create_voice_settings(id) ON DELETE CASCADE,
			blocked_users TEXT NOT NULL DEFAULT '[]',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_temp_voice_guild ON temporary_voice_channels(guild_id);
		CREATE INDEX IF NOT EXISTS idx_temp_voice_channel ON temporary_voice_channels(channel_id);
		CREATE INDEX IF NOT EXISTS idx_temp_voice_owner ON temporary_voice_channels(owner_id);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// UpsertCreateVoiceSetting fügt eine Create-Voice-Setting ein oder aktualisiert sie
func UpsertCreateVoiceSetting(db *sql.DB, setting *CreateVoiceSetting) (*CreateVoiceSetting, error) {
	query := `
		INSERT INTO create_voice_settings (guild_id, channel_id, default_user_limit, control_channel_id, control_message_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (channel_id) DO UPDATE SET
			default_user_limit = EXCLUDED.default_user_limit,
			control_channel_id = EXCLUDED.control_channel_id,
			control_message_id = EXCLUDED.control_message_id,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		setting.GuildID,
		setting.ChannelID,
		setting.DefaultUserLimit,
		setting.ControlChannelID,
		setting.ControlMessageID,
	).Scan(&setting.ID, &setting.CreatedAt, &setting.UpdatedAt)

	return setting, err
}

// GetCreateVoiceSettingByChannelID holt eine Create-Voice-Setting anhand der Channel-ID
func GetCreateVoiceSettingByChannelID(db *sql.DB, channelID string) (*CreateVoiceSetting, error) {
	query := `
		SELECT id, guild_id, channel_id, default_user_limit, control_channel_id, control_message_id, created_at, updated_at
		FROM create_voice_settings
		WHERE channel_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	setting := &CreateVoiceSetting{}
	err := db.QueryRowContext(ctx, query, channelID).Scan(
		&setting.ID,
		&setting.GuildID,
		&setting.ChannelID,
		&setting.DefaultUserLimit,
		&setting.ControlChannelID,
		&setting.ControlMessageID,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return setting, err
}

// GetCreateVoiceSettingsByGuildID holt alle Create-Voice-Settings für eine Guild
func GetCreateVoiceSettingsByGuildID(db *sql.DB, guildID string) ([]*CreateVoiceSetting, error) {
	query := `
		SELECT id, guild_id, channel_id, default_user_limit, control_channel_id, control_message_id, created_at, updated_at
		FROM create_voice_settings
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

	var settings []*CreateVoiceSetting
	for rows.Next() {
		setting := &CreateVoiceSetting{}
		err := rows.Scan(
			&setting.ID,
			&setting.GuildID,
			&setting.ChannelID,
			&setting.DefaultUserLimit,
			&setting.ControlChannelID,
			&setting.ControlMessageID,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, rows.Err()
}

// DeleteCreateVoiceSetting löscht eine Create-Voice-Setting
func DeleteCreateVoiceSetting(db *sql.DB, channelID string) error {
	query := `DELETE FROM create_voice_settings WHERE channel_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, channelID)
	return err
}

// InsertTemporaryVoiceChannel erstellt einen Eintrag für einen temporären Channel
func InsertTemporaryVoiceChannel(db *sql.DB, channel *TemporaryVoiceChannel) (*TemporaryVoiceChannel, error) {
	query := `
		INSERT INTO temporary_voice_channels (guild_id, channel_id, owner_id, create_voice_id, blocked_users)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		channel.GuildID,
		channel.ChannelID,
		channel.OwnerID,
		channel.CreateVoiceID,
		channel.BlockedUsers,
	).Scan(&channel.ID, &channel.CreatedAt)

	return channel, err
}

// GetTemporaryVoiceChannelByChannelID holt einen temporären Channel anhand der Channel-ID
func GetTemporaryVoiceChannelByChannelID(db *sql.DB, channelID string) (*TemporaryVoiceChannel, error) {
	query := `
		SELECT id, guild_id, channel_id, owner_id, create_voice_id, blocked_users, created_at
		FROM temporary_voice_channels
		WHERE channel_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channel := &TemporaryVoiceChannel{}
	err := db.QueryRowContext(ctx, query, channelID).Scan(
		&channel.ID,
		&channel.GuildID,
		&channel.ChannelID,
		&channel.OwnerID,
		&channel.CreateVoiceID,
		&channel.BlockedUsers,
		&channel.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return channel, err
}

// UpdateTemporaryVoiceChannelBlockedUsers aktualisiert die blockierten User
func UpdateTemporaryVoiceChannelBlockedUsers(db *sql.DB, channelID string, blockedUsers string) error {
	query := `
		UPDATE temporary_voice_channels
		SET blocked_users = $1
		WHERE channel_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, blockedUsers, channelID)
	return err
}

// DeleteTemporaryVoiceChannel löscht einen temporären Channel aus der Datenbank
func DeleteTemporaryVoiceChannel(db *sql.DB, channelID string) error {
	query := `DELETE FROM temporary_voice_channels WHERE channel_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, channelID)
	return err
}

// GetTemporaryChannelsByOwnerID holt alle temporären Channels eines Users
func GetTemporaryChannelsByOwnerID(db *sql.DB, ownerID string) ([]*TemporaryVoiceChannel, error) {
	query := `
		SELECT id, guild_id, channel_id, owner_id, create_voice_id, blocked_users, created_at
		FROM temporary_voice_channels
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*TemporaryVoiceChannel
	for rows.Next() {
		channel := &TemporaryVoiceChannel{}
		err := rows.Scan(
			&channel.ID,
			&channel.GuildID,
			&channel.ChannelID,
			&channel.OwnerID,
			&channel.CreateVoiceID,
			&channel.BlockedUsers,
			&channel.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, rows.Err()
}
