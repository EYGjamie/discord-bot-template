package events

import (
	"database/sql"
	"log"

	"discord-bot-template/internal/database/tables"
	"discord-bot-template/internal/shared/services/events/user"

	"github.com/bwmarrin/discordgo"
)

func OnGuildMemberRemove(bot_session *discordgo.Session, memberRemove *discordgo.GuildMemberRemove, db *sql.DB) {
	// Aktualisiere User in Datenbank (joined_at auf NULL)
	user.RemoveUser(db, memberRemove.User)

	// Ermittle den Grund für das Verlassen (leave, kick oder ban)
	reason := tables.LeaveReasonLeave
	var bannedByID *string

	// Prüfe ob User gebannt wurde
	ban, err := bot_session.GuildBan(memberRemove.GuildID, memberRemove.User.ID)
	if err == nil && ban != nil {
		reason = tables.LeaveReasonBan
		// Versuche herauszufinden wer gebannt hat über Audit Log
		auditLogs, err := bot_session.GuildAuditLog(memberRemove.GuildID, "", "", int(discordgo.AuditLogActionMemberBanAdd), 5)
		if err == nil {
			// Suche den relevanten Audit Log Entry
			for _, entry := range auditLogs.AuditLogEntries {
				if entry.TargetID == memberRemove.User.ID {
					bannedByID = &entry.UserID
					break
				}
			}
		}
		log.Printf("User %s wurde gebannt", memberRemove.User.Username)
	} else {
		// Wenn kein Ban, prüfe auf Kick über Audit Log
		auditLogs, err := bot_session.GuildAuditLog(memberRemove.GuildID, "", "", int(discordgo.AuditLogActionMemberKick), 5)
		if err == nil {
			for _, entry := range auditLogs.AuditLogEntries {
				if entry.TargetID == memberRemove.User.ID {
					reason = tables.LeaveReasonKick
					bannedByID = &entry.UserID
					break
				}
			}
		}

		if reason == tables.LeaveReasonKick {
			log.Printf("User %s wurde gekickt", memberRemove.User.Username)
		} else {
			log.Printf("User %s hat den Server verlassen", memberRemove.User.Username)
		}
	}

	// Logge den Leave in die Datenbank
	_, err = tables.LogUserLeave(db, memberRemove.User.ID, memberRemove.GuildID, reason, bannedByID)
	if err != nil {
		log.Printf("Fehler beim Loggen des Leaves für User %s: %v", memberRemove.User.ID, err)
	}
}
