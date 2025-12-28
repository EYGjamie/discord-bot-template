package tables

import (
	"context"
	"database/sql"
	"time"
)

// DiscordStatistic repräsentiert eine Momentaufnahme der Discord-Server Statistiken
type DiscordStatistic struct {
	ID                  int       `json:"id"`
	GuildID             string    `json:"guild_id"`
	MemberCount         int       `json:"member_count"`
	RoleMemberCount     int       `json:"role_member_count"` // Anzahl Mitglieder mit bestimmter Rolle
	RoleID              string    `json:"role_id"`           // ID der überwachten Rolle
	TotalChannels       int       `json:"total_channels"`
	TextChannels        int       `json:"text_channels"`
	VoiceChannels       int       `json:"voice_channels"`
	CategoryChannels    int       `json:"category_channels"`
	VoiceUserCount      int       `json:"voice_user_count"`      // Anzahl User in Voice
	ActiveVoiceChannels int       `json:"active_voice_channels"` // Anzahl aktiver Voice Channels
	TotalVoiceTime      int64     `json:"total_voice_time"`      // Gesamte Voice Time in Sekunden bis zu diesem Zeitpunkt
	Timestamp           time.Time `json:"timestamp"`
	Source              string    `json:"source"` // "manual" oder "scheduled"
	CreatedAt           time.Time `json:"created_at"`
}

// CreateDiscordStatisticsTable erstellt die Discord-Statistiken-Tabelle
func CreateDiscordStatisticsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS discord_statistics (
			id SERIAL PRIMARY KEY,
			guild_id VARCHAR(32) NOT NULL,
			member_count INTEGER NOT NULL DEFAULT 0,
			role_member_count INTEGER NOT NULL DEFAULT 0,
			role_id VARCHAR(32),
			total_channels INTEGER NOT NULL DEFAULT 0,
			text_channels INTEGER NOT NULL DEFAULT 0,
			voice_channels INTEGER NOT NULL DEFAULT 0,
			category_channels INTEGER NOT NULL DEFAULT 0,
			voice_user_count INTEGER NOT NULL DEFAULT 0,
			active_voice_channels INTEGER NOT NULL DEFAULT 0,
			total_voice_time BIGINT NOT NULL DEFAULT 0,
			timestamp TIMESTAMP NOT NULL,
			source VARCHAR(20) NOT NULL DEFAULT 'manual',
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_discord_stats_guild_id ON discord_statistics(guild_id);
		CREATE INDEX IF NOT EXISTS idx_discord_stats_timestamp ON discord_statistics(timestamp);
		CREATE INDEX IF NOT EXISTS idx_discord_stats_source ON discord_statistics(source);
		CREATE INDEX IF NOT EXISTS idx_discord_stats_guild_timestamp ON discord_statistics(guild_id, timestamp DESC);
		
		-- Migration: Füge total_voice_time hinzu falls noch nicht vorhanden
		ALTER TABLE discord_statistics ADD COLUMN IF NOT EXISTS total_voice_time BIGINT NOT NULL DEFAULT 0;
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

// SaveDiscordStatistic speichert eine neue Statistik-Aufzeichnung
func SaveDiscordStatistic(db *sql.DB, stat *DiscordStatistic) error {
	query := `
		INSERT INTO discord_statistics (
			guild_id, member_count, role_member_count, role_id,
			total_channels, text_channels, voice_channels, category_channels,
			voice_user_count, active_voice_channels, total_voice_time, timestamp, source
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.QueryRowContext(ctx, query,
		stat.GuildID, stat.MemberCount, stat.RoleMemberCount, stat.RoleID,
		stat.TotalChannels, stat.TextChannels, stat.VoiceChannels, stat.CategoryChannels,
		stat.VoiceUserCount, stat.ActiveVoiceChannels, stat.TotalVoiceTime, stat.Timestamp, stat.Source,
	).Scan(&stat.ID, &stat.CreatedAt)

	return err
}

// GetDiscordStatistics ruft Statistiken ab mit optionalen Filtern
func GetDiscordStatistics(db *sql.DB, guildID string, since *time.Time, limit int) ([]DiscordStatistic, error) {
	query := `
		SELECT 
			id, guild_id, member_count, role_member_count, role_id,
			total_channels, text_channels, voice_channels, category_channels,
			voice_user_count, active_voice_channels, total_voice_time, timestamp, source, created_at
		FROM discord_statistics
		WHERE guild_id = $1
	`

	args := []interface{}{guildID}
	argCount := 1

	if since != nil {
		argCount++
		query += ` AND timestamp >= $` + string(rune(argCount+'0'))
		args = append(args, *since)
	}

	query += ` ORDER BY timestamp DESC`

	if limit > 0 {
		argCount++
		query += ` LIMIT $` + string(rune(argCount+'0'))
		args = append(args, limit)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statistics []DiscordStatistic

	for rows.Next() {
		var stat DiscordStatistic
		err := rows.Scan(
			&stat.ID, &stat.GuildID, &stat.MemberCount, &stat.RoleMemberCount, &stat.RoleID,
			&stat.TotalChannels, &stat.TextChannels, &stat.VoiceChannels, &stat.CategoryChannels,
			&stat.VoiceUserCount, &stat.ActiveVoiceChannels, &stat.TotalVoiceTime, &stat.Timestamp, &stat.Source, &stat.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		statistics = append(statistics, stat)
	}

	return statistics, rows.Err()
}

// GetLatestDiscordStatistic ruft die neueste Statistik ab
func GetLatestDiscordStatistic(db *sql.DB, guildID string) (*DiscordStatistic, error) {
	query := `
		SELECT 
			id, guild_id, member_count, role_member_count, role_id,
			total_channels, text_channels, voice_channels, category_channels,
		       voice_user_count, active_voice_channels, total_voice_time, timestamp, source, created_at
		FROM discord_statistics
		WHERE guild_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var stat DiscordStatistic
	err := db.QueryRowContext(ctx, query, guildID).Scan(
		&stat.ID, &stat.GuildID, &stat.MemberCount, &stat.RoleMemberCount, &stat.RoleID,
		&stat.TotalChannels, &stat.TextChannels, &stat.VoiceChannels, &stat.CategoryChannels,
		&stat.VoiceUserCount, &stat.ActiveVoiceChannels, &stat.TotalVoiceTime, &stat.Timestamp, &stat.Source, &stat.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &stat, nil
}

// GetStatisticsInTimeRange ruft Statistiken in einem Zeitbereich ab
func GetStatisticsInTimeRange(db *sql.DB, guildID string, startTime, endTime time.Time) ([]DiscordStatistic, error) {
	query := `
		SELECT 
			id, guild_id, member_count, role_member_count, role_id,
			total_channels, text_channels, voice_channels, category_channels,
			voice_user_count, active_voice_channels, total_voice_time, timestamp, source, created_at
		FROM discord_statistics
		WHERE guild_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statistics []DiscordStatistic

	for rows.Next() {
		var stat DiscordStatistic
		err := rows.Scan(
			&stat.ID, &stat.GuildID, &stat.MemberCount, &stat.RoleMemberCount, &stat.RoleID,
			&stat.TotalChannels, &stat.TextChannels, &stat.VoiceChannels, &stat.CategoryChannels,
			&stat.VoiceUserCount, &stat.ActiveVoiceChannels, &stat.TotalVoiceTime, &stat.Timestamp, &stat.Source, &stat.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		statistics = append(statistics, stat)
	}

	return statistics, rows.Err()
}
