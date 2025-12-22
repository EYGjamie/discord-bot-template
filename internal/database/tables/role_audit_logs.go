package tables

import (
	"context"
	"database/sql"
	"time"
)

// RoleAuditLogAction repräsentiert die Art der Rollen-Änderung
type RoleAuditLogAction string

const (
	RoleActionAdded   RoleAuditLogAction = "ROLE_ADDED"
	RoleActionRemoved RoleAuditLogAction = "ROLE_REMOVED"
)

// RoleAuditLog repräsentiert einen Audit Log Eintrag für Rollen-Änderungen
type RoleAuditLog struct {
	ID         int64              `json:"id"`
	Action     RoleAuditLogAction `json:"action"`      // ROLE_ADDED oder ROLE_REMOVED
	GuildID    string             `json:"guild_id"`    // Guild ID
	UserID     string             `json:"user_id"`     // Betroffener User
	RoleID     string             `json:"role_id"`     // Betroffene Rolle
	ExecutorID string             `json:"executor_id"` // Wer hat die Änderung vorgenommen
	CreatedAt  time.Time          `json:"created_at"`  // Zeitstempel
}

// CreateRoleAuditLogsTable erstellt die role_audit_logs-Tabelle
func CreateRoleAuditLogsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS role_audit_logs (
			id BIGSERIAL PRIMARY KEY,
			action VARCHAR(50) NOT NULL,
			guild_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			role_id VARCHAR(255) NOT NULL,
			executor_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_role_audit_guild_user ON role_audit_logs(guild_id, user_id);
		CREATE INDEX IF NOT EXISTS idx_role_audit_created_at ON role_audit_logs(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_role_audit_executor ON role_audit_logs(executor_id);
		CREATE INDEX IF NOT EXISTS idx_role_audit_role ON role_audit_logs(role_id);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertRoleAuditLog fügt einen neuen Rollen-Audit-Log-Eintrag hinzu
func InsertRoleAuditLog(db *sql.DB, action RoleAuditLogAction, guildID, userID, roleID, executorID string) error {
	query := `
		INSERT INTO role_audit_logs (action, guild_id, user_id, role_id, executor_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, action, guildID, userID, roleID, executorID, time.Now())
	return err
}

// GetRoleAuditLogsByUser holt alle Rollen-Änderungen für einen User
func GetRoleAuditLogsByUser(db *sql.DB, guildID, userID string, limit int) ([]RoleAuditLog, error) {
	query := `
		SELECT id, action, guild_id, user_id, role_id, executor_id, created_at
		FROM role_audit_logs
		WHERE guild_id = $1 AND user_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []RoleAuditLog
	for rows.Next() {
		var log RoleAuditLog
		err := rows.Scan(&log.ID, &log.Action, &log.GuildID, &log.UserID, &log.RoleID, &log.ExecutorID, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetRoleAuditLogsByRole holt alle Änderungen für eine spezifische Rolle
func GetRoleAuditLogsByRole(db *sql.DB, guildID, roleID string, limit int) ([]RoleAuditLog, error) {
	query := `
		SELECT id, action, guild_id, user_id, role_id, executor_id, created_at
		FROM role_audit_logs
		WHERE guild_id = $1 AND role_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, guildID, roleID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []RoleAuditLog
	for rows.Next() {
		var log RoleAuditLog
		err := rows.Scan(&log.ID, &log.Action, &log.GuildID, &log.UserID, &log.RoleID, &log.ExecutorID, &log.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
