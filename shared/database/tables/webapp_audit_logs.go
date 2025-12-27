package tables

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// ActionType repräsentiert den Typ einer User-Aktion
type ActionType string

const (
	ActionTypeLogin          ActionType = "LOGIN"
	ActionTypeLogout         ActionType = "LOGOUT"
	ActionTypeView           ActionType = "VIEW"
	ActionTypeCreate         ActionType = "CREATE"
	ActionTypeUpdate         ActionType = "UPDATE"
	ActionTypeDelete         ActionType = "DELETE"
	ActionTypeWarnCreate     ActionType = "WARN_CREATE"
	ActionTypeNoteCreate     ActionType = "NOTE_CREATE"
	ActionTypeWarnDelete     ActionType = "WARN_DELETE"
	ActionTypeNoteDelete     ActionType = "NOTE_DELETE"
	ActionTypeMemberSearch   ActionType = "MEMBER_SEARCH"
	ActionTypeDashboardView  ActionType = "DASHBOARD_VIEW"
	ActionTypeProfileView    ActionType = "PROFILE_VIEW"
	ActionTypeSettingsChange ActionType = "SETTINGS_CHANGE"
	ActionTypeAPICall        ActionType = "API_CALL"
	ActionTypeError          ActionType = "ERROR"
)

// ResourceType repräsentiert den Typ einer Ressource
type ResourceType string

const (
	ResourceTypeUser      ResourceType = "USER"
	ResourceTypeMember    ResourceType = "MEMBER"
	ResourceTypeWarn      ResourceType = "WARN"
	ResourceTypeNote      ResourceType = "NOTE"
	ResourceTypeDashboard ResourceType = "DASHBOARD"
	ResourceTypeSettings  ResourceType = "SETTINGS"
	ResourceTypeAPI       ResourceType = "API"
)

// WebAppAuditLog repräsentiert einen Audit-Log-Eintrag für Webapp-Aktionen
type WebAppAuditLog struct {
	ID            int64                  `json:"id"`
	UserID        string                 `json:"user_id"`        // Discord User ID des Akteurs
	ActionType    ActionType             `json:"action_type"`    // Art der Aktion
	ResourceType  ResourceType           `json:"resource_type"`  // Art der Ressource
	ResourceID    string                 `json:"resource_id"`    // ID der betroffenen Ressource
	IPAddress     string                 `json:"ip_address"`     // IP-Adresse des Users
	UserAgent     string                 `json:"user_agent"`     // Browser User-Agent
	RequestMethod string                 `json:"request_method"` // HTTP Method (GET, POST, etc.)
	RequestPath   string                 `json:"request_path"`   // API Endpoint Pfad
	StatusCode    int                    `json:"status_code"`    // HTTP Status Code
	Details       map[string]interface{} `json:"details"`        // Zusätzliche Details als JSON
	DetailsJSON   string                 `json:"-"`              // Interne Speicherung
	CreatedAt     time.Time              `json:"created_at"`
}

// CreateWebAppAuditLogsTable erstellt die webapp_audit_logs Tabelle
func CreateWebAppAuditLogsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS webapp_audit_logs (
			id BIGSERIAL PRIMARY KEY,
			user_id VARCHAR(32),
			action_type VARCHAR(50) NOT NULL,
			resource_type VARCHAR(50) NOT NULL,
			resource_id VARCHAR(255),
			ip_address VARCHAR(45),
			user_agent TEXT,
			request_method VARCHAR(10),
			request_path VARCHAR(500),
			status_code INTEGER,
			details JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_webapp_audit_logs_user_id ON webapp_audit_logs(user_id);
		CREATE INDEX IF NOT EXISTS idx_webapp_audit_logs_action_type ON webapp_audit_logs(action_type);
		CREATE INDEX IF NOT EXISTS idx_webapp_audit_logs_resource_type ON webapp_audit_logs(resource_type);
		CREATE INDEX IF NOT EXISTS idx_webapp_audit_logs_created_at ON webapp_audit_logs(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_webapp_audit_logs_resource ON webapp_audit_logs(resource_type, resource_id);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertWebAppAuditLog fügt einen neuen Audit-Log-Eintrag in die Datenbank ein
func InsertWebAppAuditLog(db *sql.DB, log *WebAppAuditLog) error {
	// Details zu JSON konvertieren
	var detailsJSON []byte
	var err error
	if log.Details != nil {
		detailsJSON, err = json.Marshal(log.Details)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO webapp_audit_logs (user_id, action_type, resource_type, resource_id, ip_address, user_agent, request_method, request_path, status_code, details)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = db.QueryRowContext(ctx, query,
		log.UserID,
		log.ActionType,
		log.ResourceType,
		log.ResourceID,
		log.IPAddress,
		log.UserAgent,
		log.RequestMethod,
		log.RequestPath,
		log.StatusCode,
		detailsJSON,
	).Scan(&log.ID, &log.CreatedAt)

	return err
}

// GetWebAppAuditLogs ruft Audit-Logs mit Filtern ab
func GetWebAppAuditLogs(db *sql.DB, userID string, actionType ActionType, limit int, offset int) ([]*WebAppAuditLog, error) {
	query := `
		SELECT id, user_id, action_type, resource_type, resource_id, ip_address, user_agent, request_method, request_path, status_code, details, created_at
		FROM webapp_audit_logs
		WHERE ($1 = '' OR user_id = $1)
		AND ($2 = '' OR action_type = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, userID, actionType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*WebAppAuditLog
	for rows.Next() {
		log := &WebAppAuditLog{}
		var detailsJSON []byte
		var userID sql.NullString

		err := rows.Scan(
			&log.ID,
			&userID,
			&log.ActionType,
			&log.ResourceType,
			&log.ResourceID,
			&log.IPAddress,
			&log.UserAgent,
			&log.RequestMethod,
			&log.RequestPath,
			&log.StatusCode,
			&detailsJSON,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			log.UserID = userID.String
		}

		// JSON Details parsen
		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
				log.Details = make(map[string]interface{})
			}
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetUserAuditLogs ruft alle Audit-Logs für einen bestimmten User ab
func GetUserAuditLogs(db *sql.DB, userID string, limit int, offset int) ([]*WebAppAuditLog, error) {
	return GetWebAppAuditLogs(db, userID, "", limit, offset)
}

// GetAuditLogsByAction ruft Audit-Logs nach Action-Type ab
func GetAuditLogsByAction(db *sql.DB, actionType ActionType, limit int, offset int) ([]*WebAppAuditLog, error) {
	return GetWebAppAuditLogs(db, "", actionType, limit, offset)
}

// GetAuditLogsByResource ruft Audit-Logs für eine bestimmte Ressource ab
func GetAuditLogsByResource(db *sql.DB, resourceType ResourceType, resourceID string, limit int, offset int) ([]*WebAppAuditLog, error) {
	query := `
		SELECT id, user_id, action_type, resource_type, resource_id, ip_address, user_agent, request_method, request_path, status_code, details, created_at
		FROM webapp_audit_logs
		WHERE resource_type = $1 AND resource_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, resourceType, resourceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*WebAppAuditLog
	for rows.Next() {
		log := &WebAppAuditLog{}
		var detailsJSON []byte
		var userID sql.NullString

		err := rows.Scan(
			&log.ID,
			&userID,
			&log.ActionType,
			&log.ResourceType,
			&log.ResourceID,
			&log.IPAddress,
			&log.UserAgent,
			&log.RequestMethod,
			&log.RequestPath,
			&log.StatusCode,
			&detailsJSON,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			log.UserID = userID.String
		}

		// JSON Details parsen
		if len(detailsJSON) > 0 {
			if err := json.Unmarshal(detailsJSON, &log.Details); err != nil {
				log.Details = make(map[string]interface{})
			}
		}

		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// CountAuditLogs zählt die Gesamtanzahl der Audit-Logs mit Filtern
func CountAuditLogs(db *sql.DB, userID string, actionType ActionType) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM webapp_audit_logs
		WHERE ($1 = '' OR user_id = $1)
		AND ($2 = '' OR action_type = $2)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := db.QueryRowContext(ctx, query, userID, actionType).Scan(&count)
	return count, err
}

// LogUserAction ist eine Convenience-Funktion zum schnellen Loggen von User-Aktionen
func LogUserAction(db *sql.DB, userID string, actionType ActionType, resourceType ResourceType, resourceID string, details map[string]interface{}) error {
	log := &WebAppAuditLog{
		UserID:       userID,
		ActionType:   actionType,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
	}
	return InsertWebAppAuditLog(db, log)
}
