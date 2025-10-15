package tabel

import (
	"context"
	"database/sql"
	"time"
)

// UserRole repräsentiert die Many-to-Many Beziehung zwischen Users und Roles
type UserRole struct {
	UserID    string    `json:"user_id"`    // Discord User ID (Foreign Key)
	RoleID    string    `json:"role_id"`    // Discord Role ID (Foreign Key)
	CreatedAt time.Time `json:"created_at"` // Wann wurde die Rolle dem User zugewiesen
	UpdatedAt time.Time `json:"updated_at"` // Letzte Aktualisierung
}

// CreateUserRoleTable erstellt die User-Role Junction-Tabelle
func CreateUserRoleTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS user_roles (
			user_id VARCHAR(32) NOT NULL,
			role_id VARCHAR(32) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (user_id, role_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
		CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// AssignRoleToUser weist einem User eine Rolle zu
func AssignRoleToUser(db *sql.DB, userID, roleID string) (*UserRole, error) {
	query := `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO UPDATE 
		SET updated_at = CURRENT_TIMESTAMP
		RETURNING user_id, role_id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	userRole := &UserRole{}
	err := db.QueryRowContext(ctx, query, userID, roleID).Scan(
		&userRole.UserID,
		&userRole.RoleID,
		&userRole.CreatedAt,
		&userRole.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return userRole, nil
}

// RemoveRoleFromUser entfernt eine Rolle von einem User
func RemoveRoleFromUser(db *sql.DB, userID, roleID string) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, userID, roleID)
	return err
}

// GetUserRoles ruft alle Rollen eines Users ab
func GetUserRoles(db *sql.DB, userID string) ([]*Role, error) {
	query := `
		SELECT r.id, r.name, r.mention, r.created_at, r.position, r.color, r.hoist, r.mentionable, r.icon, r.updated_at
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.position DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*Role
	for rows.Next() {
		role := &Role{}
		err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Mention,
			&role.CreatedAt,
			&role.Position,
			&role.Color,
			&role.Hoist,
			&role.Mentionable,
			&role.Icon,
			&role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// GetRoleUsers ruft alle Users ab, die eine bestimmte Rolle haben
func GetRoleUsers(db *sql.DB, roleID string) ([]*User, error) {
	query := `
		SELECT u.id, u.name, u.global_name, u.display_name, u.bot, u.avatar, u.avatar_url, u.mention, u.created_at, u.nick, u.joined_at, u.top_role, u.timed_out_until, u.premium_since, u.updated_at
		FROM users u
		INNER JOIN user_roles ur ON u.id = ur.user_id
		WHERE ur.role_id = $1
		ORDER BY u.name
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.GlobalName,
			&user.DisplayName,
			&user.Bot,
			&user.Avatar,
			&user.AvatarURL,
			&user.Mention,
			&user.CreatedAt,
			&user.Nick,
			&user.JoinedAt,
			&user.TopRole,
			&user.TimedOutUntil,
			&user.PremiumSince,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// SyncUserRoles synchronisiert alle Rollen eines Users (löscht alte, fügt neue hinzu)
func SyncUserRoles(db *sql.DB, userID string, roleIDs []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Transaktion starten
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Alle bestehenden Rollen des Users löschen
	deleteQuery := `DELETE FROM user_roles WHERE user_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, userID)
	if err != nil {
		return err
	}

	// Neue Rollen einfügen
	if len(roleIDs) > 0 {
		insertQuery := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`
		for _, roleID := range roleIDs {
			_, err = tx.ExecContext(ctx, insertQuery, userID, roleID)
			if err != nil {
				return err
			}
		}
	}

	// Transaktion committen
	return tx.Commit()
}

// UserHasRole prüft, ob ein User eine bestimmte Rolle hat
func UserHasRole(db *sql.DB, userID, roleID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists bool
	err := db.QueryRowContext(ctx, query, userID, roleID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetUserRoleCount gibt die Anzahl der Rollen eines Users zurück
func GetUserRoleCount(db *sql.DB, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM user_roles WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetRoleMemberCount gibt die Anzahl der Mitglieder mit einer bestimmten Rolle zurück
func GetRoleMemberCount(db *sql.DB, roleID string) (int, error) {
	query := `SELECT COUNT(*) FROM user_roles WHERE role_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int
	err := db.QueryRowContext(ctx, query, roleID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
