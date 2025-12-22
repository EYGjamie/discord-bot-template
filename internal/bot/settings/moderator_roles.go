package settings

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"discord-bot-template/internal/database/tables"
)

// ModeratorRoles repräsentiert die Moderator-Rollen für eine Guild
type ModeratorRoles struct {
	GuildID string   `json:"guild_id"`
	RoleIDs []string `json:"role_ids"`
}

// SetModeratorRoles speichert die Moderator-Rollen für eine Guild
func SetModeratorRoles(db *sql.DB, guildID string, roleIDs []string) error {
	modRoles := ModeratorRoles{
		GuildID: guildID,
		RoleIDs: roleIDs,
	}

	jsonData, err := json.Marshal(modRoles)
	if err != nil {
		return fmt.Errorf("failed to marshal moderator roles: %w", err)
	}

	setting := &tables.BotSetting{
		Key:     fmt.Sprintf("mod_roles_%s", guildID),
		Value:   string(jsonData),
		Type:    "json",
		Enabled: true,
	}

	_, err = tables.UpsertBotSetting(db, setting)
	return err
}

// GetModeratorRoles holt die Moderator-Rollen für eine Guild
func GetModeratorRoles(db *sql.DB, guildID string) ([]string, error) {
	key := fmt.Sprintf("mod_roles_%s", guildID)
	setting, err := tables.GetBotSetting(db, key)
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil
		}
		return nil, err
	}

	var modRoles ModeratorRoles
	if err := json.Unmarshal([]byte(setting.Value), &modRoles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal moderator roles: %w", err)
	}

	return modRoles.RoleIDs, nil
}

// IsModerator prüft ob ein Member eine Moderator-Rolle hat
func IsModerator(db *sql.DB, guildID string, memberRoles []string) (bool, error) {
	modRoles, err := GetModeratorRoles(db, guildID)
	if err != nil {
		return false, err
	}

	// Admin hat immer Moderator-Rechte
	for _, roleID := range memberRoles {
		for _, modRoleID := range modRoles {
			if roleID == modRoleID {
				return true, nil
			}
		}
	}

	return false, nil
}
