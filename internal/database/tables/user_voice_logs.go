package tables

import (
	"context"
	"database/sql"
	"time"
)

// UserVoiceLog repräsentiert einen Voice-Log-Eintrag in der Datenbank
type UserVoiceLog struct {
	ID             int64      `json:"id"`
	UserID         string     `json:"user_id"`         // Discord User ID (Foreign Key zu users.id)
	ChannelID      string     `json:"channel_id"`      // Discord Voice Channel ID
	JoinedAt       time.Time  `json:"joined_at"`       // Zeitpunkt des Betretens
	LeftAt         *time.Time `json:"left_at"`         // Zeitpunkt des Verlassens (NULL wenn noch im Channel)
	MutedDuration  int64      `json:"muted_duration"`  // Gesamtzeit in Sekunden mit Mute
	DeafenDuration int64      `json:"deafen_duration"` // Gesamtzeit in Sekunden mit Deafen (Full Mute)
	StreamDuration int64      `json:"stream_duration"` // Gesamtzeit in Sekunden mit aktivem Screenshare
	TotalDuration  int64      `json:"total_duration"`  // Gesamtzeit in Sekunden im Channel (berechnet bei LeftAt)
}

// CreateUserVoiceLogsTable erstellt die user_voice_logs-Tabelle in der Datenbank
func CreateUserVoiceLogsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_voice_logs (
			id BIGSERIAL PRIMARY KEY,
			user_id VARCHAR(32) NOT NULL,
			channel_id VARCHAR(32) NOT NULL,
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			left_at TIMESTAMP,
			muted_duration BIGINT NOT NULL DEFAULT 0,
			deafen_duration BIGINT NOT NULL DEFAULT 0,
			stream_duration BIGINT NOT NULL DEFAULT 0,
			total_duration BIGINT NOT NULL DEFAULT 0,
			CONSTRAINT fk_user
				FOREIGN KEY (user_id)
				REFERENCES users(id)
				ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_user_voice_logs_user_id ON user_voice_logs(user_id);
		CREATE INDEX IF NOT EXISTS idx_user_voice_logs_channel_id ON user_voice_logs(channel_id);
		CREATE INDEX IF NOT EXISTS idx_user_voice_logs_joined_at ON user_voice_logs(joined_at);
	`

	_, err := db.Exec(query)
	return err
}

// InsertUserVoiceLog fügt einen neuen Voice-Log-Eintrag beim Betreten eines Channels ein
func InsertUserVoiceLog(db *sql.DB, userID, channelID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO user_voice_logs (user_id, channel_id, joined_at, muted_duration, deafen_duration, stream_duration, total_duration)
		VALUES ($1, $2, $3, 0, 0, 0, 0)
		RETURNING id
	`

	var id int64
	err := db.QueryRowContext(ctx, query, userID, channelID, time.Now()).Scan(&id)
	return id, err
}

// UpdateUserVoiceLogOnLeave aktualisiert einen Voice-Log-Eintrag beim Verlassen des Channels
func UpdateUserVoiceLogOnLeave(db *sql.DB, logID int64, mutedDuration, deafenDuration, streamDuration int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE user_voice_logs
		SET left_at = $1,
		    muted_duration = $2,
		    deafen_duration = $3,
		    stream_duration = $4,
		    total_duration = EXTRACT(EPOCH FROM ($1 - joined_at))::BIGINT
		WHERE id = $5
	`

	_, err := db.ExecContext(ctx, query, time.Now(), mutedDuration, deafenDuration, streamDuration, logID)
	return err
}

// UpdateUserVoiceLogDurations aktualisiert die Durations eines aktiven Voice-Log-Eintrags
func UpdateUserVoiceLogDurations(db *sql.DB, logID int64, mutedDuration, deafenDuration, streamDuration int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE user_voice_logs
		SET muted_duration = $1,
		    deafen_duration = $2,
		    stream_duration = $3
		WHERE id = $4
	`

	_, err := db.ExecContext(ctx, query, mutedDuration, deafenDuration, streamDuration, logID)
	return err
}

// GetActiveVoiceLogByUser ruft den aktiven Voice-Log-Eintrag für einen User ab (left_at IS NULL)
func GetActiveVoiceLogByUser(db *sql.DB, userID string) (*UserVoiceLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, channel_id, joined_at, left_at, muted_duration, deafen_duration, stream_duration, total_duration
		FROM user_voice_logs
		WHERE user_id = $1 AND left_at IS NULL
		ORDER BY joined_at DESC
		LIMIT 1
	`

	var log UserVoiceLog
	err := db.QueryRowContext(ctx, query, userID).Scan(
		&log.ID, &log.UserID, &log.ChannelID, &log.JoinedAt,
		&log.LeftAt, &log.MutedDuration, &log.DeafenDuration,
		&log.StreamDuration, &log.TotalDuration,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &log, nil
}

// GetUserVoiceLogsByUser ruft alle Voice-Logs für einen bestimmten User ab
func GetUserVoiceLogsByUser(db *sql.DB, userID string) ([]UserVoiceLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, channel_id, joined_at, left_at, muted_duration, deafen_duration, stream_duration, total_duration
		FROM user_voice_logs
		WHERE user_id = $1
		ORDER BY joined_at DESC
	`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []UserVoiceLog
	for rows.Next() {
		var log UserVoiceLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.ChannelID, &log.JoinedAt,
			&log.LeftAt, &log.MutedDuration, &log.DeafenDuration,
			&log.StreamDuration, &log.TotalDuration); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetUserVoiceLogsByChannel ruft alle Voice-Logs für einen bestimmten Channel ab
func GetUserVoiceLogsByChannel(db *sql.DB, channelID string) ([]UserVoiceLog, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, channel_id, joined_at, left_at, muted_duration, deafen_duration, stream_duration, total_duration
		FROM user_voice_logs
		WHERE channel_id = $1
		ORDER BY joined_at DESC
	`

	rows, err := db.QueryContext(ctx, query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []UserVoiceLog
	for rows.Next() {
		var log UserVoiceLog
		if err := rows.Scan(&log.ID, &log.UserID, &log.ChannelID, &log.JoinedAt,
			&log.LeftAt, &log.MutedDuration, &log.DeafenDuration,
			&log.StreamDuration, &log.TotalDuration); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetUserVoiceStatistics berechnet Gesamtstatistiken für einen User
func GetUserVoiceStatistics(db *sql.DB, userID string) (totalTime, mutedTime, deafenTime, streamTime int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT 
			COALESCE(SUM(total_duration), 0) as total_time,
			COALESCE(SUM(muted_duration), 0) as muted_time,
			COALESCE(SUM(deafen_duration), 0) as deafen_time,
			COALESCE(SUM(stream_duration), 0) as stream_time
		FROM user_voice_logs
		WHERE user_id = $1 AND left_at IS NOT NULL
	`

	err = db.QueryRowContext(ctx, query, userID).Scan(&totalTime, &mutedTime, &deafenTime, &streamTime)
	return
}
