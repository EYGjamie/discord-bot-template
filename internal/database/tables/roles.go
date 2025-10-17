package tables

import (
	"context"
	"database/sql"
	"time"
)

// Role repräsentiert eine Discord-Rolle in der Datenbank
type Role struct {
	ID          string    `json:"id"`          // Discord Role ID als Primary Key
	Name        string    `json:"name"`        // Rollenname
	Mention     string    `json:"mention"`     // Discord Mention String
	CreatedAt   time.Time `json:"created_at"`  // Discord Rolle Erstellungsdatum
	Position    int       `json:"position"`    // Position in der Rollenhierarchie
	Color       int       `json:"color"`       // Rollenfarbe (Hex als Integer)
	Hoist       bool      `json:"hoist"`       // Wird separat angezeigt
	Mentionable bool      `json:"mentionable"` // Kann erwähnt werden
	Icon        string    `json:"icon"`        // Rollen-Icon Hash
	UpdatedAt   time.Time `json:"updated_at"`  // Letzte Aktualisierung
}

// CreateRoleTable erstellt die Role-Tabelle in der Datenbank
func CreateRoleTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS roles (
			id VARCHAR(32) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			mention VARCHAR(50),
			created_at TIMESTAMP NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			color INTEGER NOT NULL DEFAULT 0,
			hoist BOOLEAN NOT NULL DEFAULT FALSE,
			mentionable BOOLEAN NOT NULL DEFAULT FALSE,
			icon VARCHAR(255),
			alias VARCHAR(255),
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
		CREATE INDEX IF NOT EXISTS idx_roles_position ON roles(position);
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query)
	return err
}

// InsertRole fügt eine neue Discord-Rolle in die Datenbank ein
func InsertRole(db *sql.DB, role *Role) (*Role, error) {
	query := `
		INSERT INTO roles (id, name, mention, created_at, position, color, hoist, mentionable, icon)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, mention, created_at, position, color, hoist, mentionable, icon, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := &Role{}
	err := db.QueryRowContext(ctx, query,
		role.ID,
		role.Name,
		role.Mention,
		role.CreatedAt,
		role.Position,
		role.Color,
		role.Hoist,
		role.Mentionable,
		role.Icon,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Mention,
		&result.CreatedAt,
		&result.Position,
		&result.Color,
		&result.Hoist,
		&result.Mentionable,
		&result.Icon,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetRoleByID ruft eine Discord-Rolle anhand der Role ID ab
func GetRoleByID(db *sql.DB, id string) (*Role, error) {
	query := `
		SELECT id, name, mention, created_at, position, color, hoist, mentionable, icon, updated_at
		FROM roles
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	role := &Role{}
	err := db.QueryRowContext(ctx, query, id).Scan(
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

	return role, nil
}

// GetRoleByName ruft eine Discord-Rolle anhand des Namens ab
func GetRoleByName(db *sql.DB, name string) (*Role, error) {
	query := `
		SELECT id, name, mention, created_at, position, color, hoist, mentionable, icon, updated_at
		FROM roles
		WHERE name = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	role := &Role{}
	err := db.QueryRowContext(ctx, query, name).Scan(
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

	return role, nil
}

// GetRolesByAlias ruft alle Discord-Rollen mit einem bestimmten Alias ab
func GetRolesByAlias(db *sql.DB, alias string) ([]*Role, error) {
	query := `
		SELECT id, name, mention, created_at, position, color, hoist, mentionable, icon, updated_at
		FROM roles
		WHERE alias = $1
		ORDER BY position DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query, alias)
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

// UpdateRole aktualisiert eine Discord-Rolle
func UpdateRole(db *sql.DB, role *Role) (*Role, error) {
	query := `
		UPDATE roles
		SET name = $2, mention = $3, created_at = $4, position = $5, color = $6, 
		    hoist = $7, mentionable = $8, icon = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, mention, created_at, position, color, hoist, mentionable, icon, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result := &Role{}
	err := db.QueryRowContext(ctx, query,
		role.ID,
		role.Name,
		role.Mention,
		role.CreatedAt,
		role.Position,
		role.Color,
		role.Hoist,
		role.Mentionable,
		role.Icon,
	).Scan(
		&result.ID,
		&result.Name,
		&result.Mention,
		&result.CreatedAt,
		&result.Position,
		&result.Color,
		&result.Hoist,
		&result.Mentionable,
		&result.Icon,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRole löscht eine Discord-Rolle
func DeleteRole(db *sql.DB, id string) error {
	query := `DELETE FROM roles WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, id)
	return err
}

// GetAllRoles ruft alle Discord-Rollen ab, sortiert nach Position
func GetAllRoles(db *sql.DB) ([]*Role, error) {
	query := `
		SELECT id, name, mention, created_at, position, color, hoist, mentionable, icon, updated_at
		FROM roles
		ORDER BY position DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
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

// GetMentionableRoles ruft alle erwähnbaren Rollen ab
func GetMentionableRoles(db *sql.DB) ([]*Role, error) {
	query := `
		SELECT id, name, mention, created_at, position, color, hoist, mentionable, icon, updated_at
		FROM roles
		WHERE mentionable = TRUE
		ORDER BY position DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
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
