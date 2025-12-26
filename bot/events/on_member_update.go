package events

import (
	"database/sql"
	"log"
	"time"

	"discord-bot-template/shared/database/tables"
	"discord-bot-template/bot/services/events/user"

	"github.com/bwmarrin/discordgo"
)

func OnGuildMemberUpdate(bot_session *discordgo.Session, memberUpdate *discordgo.GuildMemberUpdate, db *sql.DB) {
	// Aktualisiere User in Datenbank
	user.UpsertUser(db, memberUpdate.User, memberUpdate.Member)

	// Prüfe Rollen-Änderungen wenn vorheriger State vorhanden
	if memberUpdate.BeforeUpdate != nil {
		logRoleChanges(bot_session, memberUpdate, db)
	}
}

// logRoleChanges vergleicht alte und neue Rollen und loggt Änderungen
func logRoleChanges(s *discordgo.Session, m *discordgo.GuildMemberUpdate, db *sql.DB) {
	oldRoles := make(map[string]bool)
	for _, roleID := range m.BeforeUpdate.Roles {
		oldRoles[roleID] = true
	}

	newRoles := make(map[string]bool)
	for _, roleID := range m.Member.Roles {
		newRoles[roleID] = true
	}

	// Hinzugefügte Rollen
	for _, roleID := range m.Member.Roles {
		if !oldRoles[roleID] {
			executor := getExecutorFromAuditLog(s, m.GuildID, m.User.ID, discordgo.AuditLogActionMemberRoleUpdate)
			if err := tables.InsertRoleAuditLog(db, tables.RoleActionAdded, m.GuildID, m.User.ID, roleID, executor); err != nil {
				log.Printf("Error logging role addition: %v", err)
			} else {
				log.Printf("Role %s added to user %s (by %s)", roleID, m.User.Username, executor)
			}
		}
	}

	// Entfernte Rollen
	for _, roleID := range m.BeforeUpdate.Roles {
		if !newRoles[roleID] {
			executor := getExecutorFromAuditLog(s, m.GuildID, m.User.ID, discordgo.AuditLogActionMemberRoleUpdate)
			if err := tables.InsertRoleAuditLog(db, tables.RoleActionRemoved, m.GuildID, m.User.ID, roleID, executor); err != nil {
				log.Printf("Error logging role removal: %v", err)
			} else {
				log.Printf("Role %s removed from user %s (by %s)", roleID, m.User.Username, executor)
			}
		}
	}
}

// getExecutorFromAuditLog versucht den Executor aus den Audit Logs zu ermitteln
func getExecutorFromAuditLog(s *discordgo.Session, guildID, targetUserID string, actionType discordgo.AuditLogAction) string {
	auditLog, err := s.GuildAuditLog(guildID, "", "", int(actionType), 5)
	if err != nil {
		log.Printf("Warning: Could not fetch audit log: %v", err)
		return "Unknown"
	}

	// Finde den passenden Audit Log Eintrag (innerhalb der letzten 5 Sekunden)
	currentTime := time.Now()
	for _, entry := range auditLog.AuditLogEntries {
		if entry.TargetID == targetUserID {
			// Snowflake ID zu Zeit konvertieren (Discord Snowflake Epoch: 1420070400000)
			// entry.ID ist ein String, daher müssen wir ihn zuerst parsen
			idInt, err := discordgo.SnowflakeTimestamp(entry.ID)
			if err != nil {
				log.Printf("Warning: Could not parse snowflake ID: %v", err)
				continue
			}

			// Prüfe ob der Eintrag aktuell ist (max 5 Sekunden alt)
			if currentTime.Sub(idInt) < 5*time.Second {
				return entry.UserID
			}
		}
	}

	return "Unknown"
}
