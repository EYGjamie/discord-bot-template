package user

import (
	"database/sql"
	"log"
	"time"

	"discord-bot-template/internal/database/tables"

	"github.com/bwmarrin/discordgo"
)

// UpsertUser erstellt oder aktualisiert einen User in der Datenbank
// Pro User wird genau ein Datensatz gespeichert
func UpsertUser(db *sql.DB, discordUser *discordgo.User, member *discordgo.Member) {
	if discordUser == nil {
		log.Println("Warnung: User ist nil")
		return
	}

	// Erstelle User-Objekt mit Discord-Daten
	user := buildUserFromDiscord(discordUser, member)

	// Nutze die UpsertUser-Funktion aus tables, die ON CONFLICT verwendet
	_, err := tables.UpsertUser(db, user)
	if err != nil {
		log.Printf("Fehler beim Upsert des Users %s: %v", discordUser.ID, err)
		return
	}

	log.Printf("User %s (%s) wurde in die Datenbank synchronisiert", discordUser.Username, discordUser.ID)
}

// RemoveUser setzt das joined_at auf NULL wenn ein User den Server verlässt
// Der User-Datensatz bleibt erhalten für historische Daten
func RemoveUser(db *sql.DB, discordUser *discordgo.User) {
	if discordUser == nil {
		log.Println("Warnung: User ist nil")
		return
	}

	// Hole existierenden User
	existingUser, err := tables.GetUserByID(db, discordUser.ID)
	if err == sql.ErrNoRows || existingUser == nil {
		log.Printf("User %s existiert nicht in der Datenbank", discordUser.ID)
		return
	} else if err != nil {
		log.Printf("Fehler beim Abrufen des Users %s: %v", discordUser.ID, err)
		return
	}

	// Setze joined_at auf NULL
	existingUser.JoinedAt = nil
	existingUser.Nick = ""
	existingUser.TopRole = ""
	existingUser.TimedOutUntil = nil
	existingUser.PremiumSince = nil

	_, err = tables.UpdateUser(db, existingUser)
	if err != nil {
		log.Printf("Fehler beim Aktualisieren des Users %s nach Remove: %v", discordUser.ID, err)
		return
	}
	log.Printf("User %s (%s) hat den Server verlassen - joined_at wurde auf NULL gesetzt", discordUser.Username, discordUser.ID)
}

// buildUserFromDiscord erstellt ein User-Objekt aus Discord-Daten
func buildUserFromDiscord(discordUser *discordgo.User, member *discordgo.Member) *tables.User {
	// Discord Snowflake ID zu Timestamp konvertieren
	// Discord Snowflake Format: ((timestamp_ms - DISCORD_EPOCH) << 22) | internal_data
	const discordEpoch = 1420070400000 // Discord Epoch: 2015-01-01T00:00:00.000Z
	userID := discordUser.ID
	var createdAt time.Time

	// Versuche Discord ID in Timestamp zu konvertieren
	if id, err := time.Parse("", userID); err == nil {
		createdAt = id
	} else {
		// Fallback auf aktuelle Zeit wenn Parsing fehlschlägt
		createdAt = time.Now()
	}

	user := &tables.User{
		ID:          discordUser.ID,
		Name:        discordUser.Username,
		GlobalName:  discordUser.GlobalName,
		DisplayName: discordUser.GlobalName, // Fallback auf GlobalName
		Bot:         discordUser.Bot,
		Avatar:      discordUser.Avatar,
		AvatarURL:   discordUser.AvatarURL(""),
		Mention:     discordUser.Mention(),
		CreatedAt:   createdAt,
	}

	// Wenn Member-Daten vorhanden sind (Guild-spezifische Daten)
	if member != nil {
		if member.Nick != "" {
			user.Nick = member.Nick
			user.DisplayName = member.Nick // Nickname hat Vorrang
		}

		// JoinedAt ist bereits ein time.Time, kein Pointer
		if !member.JoinedAt.IsZero() {
			joinedTime := member.JoinedAt
			user.JoinedAt = &joinedTime
		}

		// Höchste Rolle ermitteln
		if len(member.Roles) > 0 {
			user.TopRole = member.Roles[0] // Erste Rolle ist die höchste
		}

		// Timeout-Status
		if member.CommunicationDisabledUntil != nil && member.CommunicationDisabledUntil.After(time.Now()) {
			user.TimedOutUntil = member.CommunicationDisabledUntil
		}

		// Premium (Boost) Status
		if member.PremiumSince != nil && !member.PremiumSince.IsZero() {
			user.PremiumSince = member.PremiumSince
		}
	}

	return user
}
