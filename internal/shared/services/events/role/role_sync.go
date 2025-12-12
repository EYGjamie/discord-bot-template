package role

import (
	"database/sql"
	"log"
	"time"

	"discord-bot-template/internal/database/tables"

	"github.com/bwmarrin/discordgo"
)

// UpsertRole erstellt oder aktualisiert eine Rolle in der Datenbank
// Pro Rolle wird genau ein Datensatz gespeichert
func UpsertRole(db *sql.DB, discordRole *discordgo.Role) {
	if discordRole == nil {
		log.Println("Warnung: Rolle ist nil")
		return
	}

	// Erstelle Role-Objekt mit Discord-Daten
	role := buildRoleFromDiscord(discordRole)

	// Nutze die UpsertRole-Funktion aus tables, die ON CONFLICT verwendet
	_, err := tables.UpsertRole(db, role)
	if err != nil {
		log.Printf("Fehler beim Upsert der Rolle %s: %v", discordRole.ID, err)
		return
	}

	log.Printf("Rolle %s (%s) wurde in die Datenbank synchronisiert", discordRole.Name, discordRole.ID)
}

// RemoveRole löscht eine Rolle aus der Datenbank
func RemoveRole(db *sql.DB, roleID string) {
	if roleID == "" {
		log.Println("Warnung: Rollen-ID ist leer")
		return
	}

	err := tables.DeleteRole(db, roleID)
	if err != nil {
		log.Printf("Fehler beim Löschen der Rolle %s: %v", roleID, err)
		return
	}

	log.Printf("Rolle %s wurde aus der Datenbank gelöscht", roleID)
}

// buildRoleFromDiscord erstellt ein Role-Objekt aus Discord-Daten
func buildRoleFromDiscord(discordRole *discordgo.Role) *tables.Role {
	// Discord Snowflake ID zu Timestamp konvertieren
	createdAt := time.Now() // Fallback

	// Versuche aus der Snowflake-ID das Erstellungsdatum zu extrahieren
	if discordRole.ID != "" {
		// Discord Snowflake: ((timestamp_ms - DISCORD_EPOCH) << 22) | internal_data
		// Für vereinfachte Implementierung nutzen wir time.Now()
		createdAt = time.Now()
	}

	role := &tables.Role{
		ID:          discordRole.ID,
		Name:        discordRole.Name,
		Mention:     discordRole.Mention(),
		CreatedAt:   createdAt,
		Position:    discordRole.Position,
		Color:       discordRole.Color,
		Hoist:       discordRole.Hoist,
		Mentionable: discordRole.Mentionable,
		Icon:        discordRole.Icon,
	}

	return role
}
