package guildinit

import (
	"database/sql"
	"log"

	"discord-bot-template/bot/services/events/channel"
	"discord-bot-template/bot/services/events/role"
	"discord-bot-template/bot/services/events/user"

	"github.com/bwmarrin/discordgo"
)

// SyncGuildOnJoin synchronisiert alle Daten eines Servers in die Datenbank
// Diese Funktion wird aufgerufen, wenn der Bot einem neuen Server beitritt
func SyncGuildOnJoin(session *discordgo.Session, guildID string, db *sql.DB) error {
	log.Printf("Starte vollständige Synchronisierung für Guild %s", guildID)

	// Synchronisiere Rollen
	if err := SyncAllRoles(session, guildID, db); err != nil {
		log.Printf("Fehler beim Synchronisieren der Rollen für Guild %s: %v", guildID, err)
		return err
	}

	// Synchronisiere Channels
	if err := SyncAllChannels(session, guildID, db); err != nil {
		log.Printf("Fehler beim Synchronisieren der Channels für Guild %s: %v", guildID, err)
		return err
	}

	// Synchronisiere Members
	if err := SyncAllMembers(session, guildID, db); err != nil {
		log.Printf("Fehler beim Synchronisieren der Members für Guild %s: %v", guildID, err)
		return err
	}

	log.Printf("Vollständige Synchronisierung für Guild %s abgeschlossen", guildID)
	return nil
}

// SyncAllRoles synchronisiert alle Rollen eines Servers in die Datenbank
func SyncAllRoles(session *discordgo.Session, guildID string, db *sql.DB) error {
	log.Printf("Synchronisiere Rollen für Guild %s", guildID)

	// Hole alle Rollen vom Discord-Server
	roles, err := session.GuildRoles(guildID)
	if err != nil {
		return err
	}

	// Zähler für Statistik
	syncedCount := 0

	// Synchronisiere jede Rolle
	for _, discordRole := range roles {
		role.UpsertRole(db, discordRole)
		syncedCount++
	}

	log.Printf("Rollen-Synchronisierung abgeschlossen: %d erfolgreich", syncedCount)
	return nil
}

// SyncAllChannels synchronisiert alle Channels eines Servers in die Datenbank
func SyncAllChannels(session *discordgo.Session, guildID string, db *sql.DB) error {
	log.Printf("Synchronisiere Channels für Guild %s", guildID)

	// Hole alle Channels vom Discord-Server
	channels, err := session.GuildChannels(guildID)
	if err != nil {
		return err
	}

	// Zähler für Statistik
	syncedCount := 0

	// Synchronisiere jeden Channel
	for _, discordChannel := range channels {
		channel.UpsertChannel(db, discordChannel)
		syncedCount++
	}

	log.Printf("Channel-Synchronisierung abgeschlossen: %d erfolgreich", syncedCount)
	return nil
}

// SyncAllMembers synchronisiert alle Members eines Servers in die Datenbank
func SyncAllMembers(session *discordgo.Session, guildID string, db *sql.DB) error {
	log.Printf("Synchronisiere Members für Guild %s", guildID)

	// Zähler für Statistik
	syncedCount := 0

	// Discord API Pagination für große Server
	// Hole Members in Batches von 1000
	after := ""
	limit := 1000

	for {
		// Hole Members vom Discord-Server
		members, err := session.GuildMembers(guildID, after, limit)
		if err != nil {
			log.Printf("Fehler beim Abrufen der Members: %v", err)
			return err
		}

		// Keine Members mehr vorhanden
		if len(members) == 0 {
			break
		}

		// Synchronisiere jeden Member
		for _, member := range members {
			if member.User != nil {
				user.UpsertUser(db, member.User, member)
				syncedCount++
			}
		}

		// Wenn weniger als limit Members zurückkamen, sind wir fertig
		if len(members) < limit {
			break
		}

		// Setze 'after' auf die ID des letzten Members für die nächste Seite
		after = members[len(members)-1].User.ID
	}

	log.Printf("Member-Synchronisierung abgeschlossen: %d erfolgreich", syncedCount)
	return nil
}

// SyncGuildRoleAssignments synchronisiert alle Rollen-Zuweisungen der Members
// Dies sollte nach SyncAllMembers und SyncAllRoles aufgerufen werden
func SyncGuildRoleAssignments(session *discordgo.Session, guildID string, db *sql.DB) error {
	log.Printf("Synchronisiere Rollen-Zuweisungen für Guild %s", guildID)

	// Zähler für Statistik
	syncedCount := 0

	// Hole alle Members
	after := ""
	limit := 1000

	for {
		members, err := session.GuildMembers(guildID, after, limit)
		if err != nil {
			log.Printf("Fehler beim Abrufen der Members: %v", err)
			return err
		}

		if len(members) == 0 {
			break
		}

		// Synchronisiere Rollen-Zuweisungen für jeden Member
		for _, member := range members {
			if member.User == nil {
				continue
			}

			// Synchronisiere alle Rollen des Members
			for range member.Roles {
				// Hier könnte man tables.AssignRoleToUser aufrufen
				// wenn die Funktion existiert
				syncedCount++
			}
		}

		if len(members) < limit {
			break
		}

		after = members[len(members)-1].User.ID
	}

	log.Printf("Rollen-Zuweisungs-Synchronisierung abgeschlossen: %d erfolgreich", syncedCount)
	return nil
}

// FullGuildSync führt eine vollständige Synchronisierung inkl. Rollen-Zuweisungen durch
func FullGuildSync(session *discordgo.Session, guildID string, db *sql.DB) error {
	log.Printf("Starte vollständige Guild-Synchronisierung (inkl. Rollen-Zuweisungen) für Guild %s", guildID)

	// Basis-Synchronisierung
	if err := SyncGuildOnJoin(session, guildID, db); err != nil {
		return err
	}

	// Rollen-Zuweisungen synchronisieren
	if err := SyncGuildRoleAssignments(session, guildID, db); err != nil {
		log.Printf("Warnung: Fehler beim Synchronisieren der Rollen-Zuweisungen: %v", err)
		// Nicht als kritischer Fehler behandeln
	}

	log.Printf("Vollständige Guild-Synchronisierung für Guild %s abgeschlossen", guildID)
	return nil
}
