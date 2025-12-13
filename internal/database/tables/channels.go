package tables

import (
	"context"
	"database/sql"
	"time"
)

// Channel repräsentiert einen Discord-Channel in der Datenbank
type Channel struct {
	ID                   string    `json:"id"`                    // Discord Channel ID als Primary Key
	GuildID              string    `json:"guild_id"`              // Discord Guild/Server ID
	Name                 string    `json:"name"`                  // Channel-Name
	Type                 int       `json:"type"`                  // Channel-Typ (Text, Voice, etc.)
	Position             int       `json:"position"`              // Position in der Channel-Liste
	Topic                string    `json:"topic"`                 // Channel-Thema/Beschreibung
	NSFW                 bool      `json:"nsfw"`                  // Ist NSFW-Channel
	Bitrate              int       `json:"bitrate"`               // Bitrate für Voice-Channels
	UserLimit            int       `json:"user_limit"`            // User-Limit für Voice-Channels
	ParentID             string    `json:"parent_id"`             // Parent Category ID
	PermissionOverwrites string    `json:"permission_overwrites"` // JSON String mit Permission Overwrites
	CreatedAt            time.Time `json:"created_at"`            // Discord Channel Erstellungsdatum
	UpdatedAt            time.Time `json:"updated_at"`            // Letzte Aktualisierung in DB
}

// CreateChannelTable erstellt die Channel-Tabelle in der Datenbank
func CreateChannelTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS channels (
			id VARCHAR(32) PRIMARY KEY,
			guild_id VARCHAR(32) NOT NULL,
			name VARCHAR(255) NOT NULL,
			type INTEGER NOT NULL DEFAULT 0,
			position INTEGER NOT NULL DEFAULT 0,
			topic TEXT,
			nsfw BOOLEAN NOT NULL DEFAULT FALSE,
			bitrate INTEGER NOT NULL DEFAULT 0,
			user_limit INTEGER NOT NULL DEFAULT 0,
			parent_id VARCHAR(32),
			permission_overwrites TEXT,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_channels_guild_id ON channels(guild_id);
		CREATE INDEX IF NOT EXISTS idx_channels_name ON channels(name);
		CREATE INDEX IF NOT EXISTS idx_channels_type ON channels(type);
		CREATE INDEX IF NOT EXISTS idx_channels_parent_id ON channels(parent_id);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertChannel fügt einen neuen Discord-Channel in die Datenbank ein
func InsertChannel(db *sql.DB, channel *Channel) (*Channel, error) {
	query := `
		INSERT INTO channels (id, guild_id, name, type, position, topic, nsfw, bitrate, user_limit, parent_id, permission_overwrites, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, guild_id, name, type, position, topic, nsfw, bitrate, user_limit, parent_id, permission_overwrites, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := &Channel{}
	err := db.QueryRowContext(ctx, query,
		channel.ID,
		channel.GuildID,
		channel.Name,
		channel.Type,
		channel.Position,
		channel.Topic,
		channel.NSFW,
		channel.Bitrate,
		channel.UserLimit,
		channel.ParentID,
		channel.PermissionOverwrites,
		channel.CreatedAt,
	).Scan(
		&result.ID,
		&result.GuildID,
		&result.Name,
		&result.Type,
		&result.Position,
		&result.Topic,
		&result.NSFW,
		&result.Bitrate,
		&result.UserLimit,
		&result.ParentID,
		&result.PermissionOverwrites,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetChannelByID ruft einen Discord-Channel anhand der Channel ID ab
func GetChannelByID(db *sql.DB, id string) (*Channel, error) {
	query := `
		SELECT id, guild_id, name, type, position, topic, nsfw, bitrate, user_limit, parent_id, permission_overwrites, created_at, updated_at
		FROM channels
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	channel := &Channel{}
	err := db.QueryRowContext(ctx, query, id).Scan(
		&channel.ID,
		&channel.GuildID,
		&channel.Name,
		&channel.Type,
		&channel.Position,
		&channel.Topic,
		&channel.NSFW,
		&channel.Bitrate,
		&channel.UserLimit,
		&channel.ParentID,
		&channel.PermissionOverwrites,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return channel, nil
}

// GetChannelsByGuildID ruft alle Channels einer Guild ab
func GetChannelsByGuildID(db *sql.DB, guildID string) ([]*Channel, error) {
	query := `
		SELECT id, guild_id, name, type, position, topic, nsfw, bitrate, user_limit, parent_id, permission_overwrites, created_at, updated_at
		FROM channels
		WHERE guild_id = $1
		ORDER BY position ASC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		channel := &Channel{}
		err := rows.Scan(
			&channel.ID,
			&channel.GuildID,
			&channel.Name,
			&channel.Type,
			&channel.Position,
			&channel.Topic,
			&channel.NSFW,
			&channel.Bitrate,
			&channel.UserLimit,
			&channel.ParentID,
			&channel.PermissionOverwrites,
			&channel.CreatedAt,
			&channel.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, rows.Err()
}

// GetChannelsByType ruft alle Channels eines bestimmten Typs ab
func GetChannelsByType(db *sql.DB, guildID string, channelType int) ([]*Channel, error) {
	query := `
		SELECT id, guild_id, name, type, position, topic, nsfw, bitrate, user_limit, parent_id, permission_overwrites, created_at, updated_at
		FROM channels
		WHERE guild_id = $1 AND type = $2
		ORDER BY position ASC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID, channelType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		channel := &Channel{}
		err := rows.Scan(
			&channel.ID,
			&channel.GuildID,
			&channel.Name,
			&channel.Type,
			&channel.Position,
			&channel.Topic,
			&channel.NSFW,
			&channel.Bitrate,
			&channel.UserLimit,
			&channel.ParentID,
			&channel.PermissionOverwrites,
			&channel.CreatedAt,
			&channel.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, rows.Err()
}

// UpdateChannel aktualisiert einen Discord-Channel
func UpdateChannel(db *sql.DB, channel *Channel) (*Channel, error) {
	query := `
		UPDATE channels
		SET guild_id = $2,
		    name = $3,
		    type = $4,
		    position = $5,
		    topic = $6,
		    nsfw = $7,
		    bitrate = $8,
		    user_limit = $9,
		    parent_id = $10,
		    permission_overwrites = $11,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, guild_id, name, type, position, topic, nsfw, bitrate, user_limit, parent_id, permission_overwrites, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := &Channel{}
	err := db.QueryRowContext(ctx, query,
		channel.ID,
		channel.GuildID,
		channel.Name,
		channel.Type,
		channel.Position,
		channel.Topic,
		channel.NSFW,
		channel.Bitrate,
		channel.UserLimit,
		channel.ParentID,
		channel.PermissionOverwrites,
	).Scan(
		&result.ID,
		&result.GuildID,
		&result.Name,
		&result.Type,
		&result.Position,
		&result.Topic,
		&result.NSFW,
		&result.Bitrate,
		&result.UserLimit,
		&result.ParentID,
		&result.PermissionOverwrites,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpsertChannel fügt einen neuen Channel ein oder aktualisiert einen existierenden
// Pro Channel gibt es genau einen Datensatz in der Datenbank
func UpsertChannel(db *sql.DB, channel *Channel) (*Channel, error) {
	query := `
		INSERT INTO channels (id, guild_id, name, type, position, topic, nsfw, bitrate, user_limit, parent_id, permission_overwrites, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET
			guild_id = EXCLUDED.guild_id,
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			position = EXCLUDED.position,
			topic = EXCLUDED.topic,
			nsfw = EXCLUDED.nsfw,
			bitrate = EXCLUDED.bitrate,
			user_limit = EXCLUDED.user_limit,
			parent_id = EXCLUDED.parent_id,
			permission_overwrites = EXCLUDED.permission_overwrites,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, guild_id, name, type, position, topic, nsfw, bitrate, user_limit, parent_id, permission_overwrites, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := &Channel{}
	err := db.QueryRowContext(ctx, query,
		channel.ID,
		channel.GuildID,
		channel.Name,
		channel.Type,
		channel.Position,
		channel.Topic,
		channel.NSFW,
		channel.Bitrate,
		channel.UserLimit,
		channel.ParentID,
		channel.PermissionOverwrites,
		channel.CreatedAt,
	).Scan(
		&result.ID,
		&result.GuildID,
		&result.Name,
		&result.Type,
		&result.Position,
		&result.Topic,
		&result.NSFW,
		&result.Bitrate,
		&result.UserLimit,
		&result.ParentID,
		&result.PermissionOverwrites,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteChannel löscht einen Discord-Channel
func DeleteChannel(db *sql.DB, id string) error {
	query := `DELETE FROM channels WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, id)
	return err
}

// GetAllChannels ruft alle Discord-Channels ab, sortiert nach Guild und Position
func GetAllChannels(db *sql.DB) ([]*Channel, error) {
	query := `
		SELECT id, guild_id, name, type, position, topic, nsfw, bitrate, user_limit, parent_id, permission_overwrites, created_at, updated_at
		FROM channels
		ORDER BY guild_id ASC, position ASC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*Channel
	for rows.Next() {
		channel := &Channel{}
		err := rows.Scan(
			&channel.ID,
			&channel.GuildID,
			&channel.Name,
			&channel.Type,
			&channel.Position,
			&channel.Topic,
			&channel.NSFW,
			&channel.Bitrate,
			&channel.UserLimit,
			&channel.ParentID,
			&channel.PermissionOverwrites,
			&channel.CreatedAt,
			&channel.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}

	return channels, rows.Err()
}
