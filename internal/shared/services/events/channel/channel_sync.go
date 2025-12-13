package channel

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"discord-bot-template/internal/database/tables"

	"github.com/bwmarrin/discordgo"
)

// UpsertChannel erstellt oder aktualisiert einen Channel in der Datenbank
func UpsertChannel(db *sql.DB, discordChannel *discordgo.Channel) {
	if discordChannel == nil {
		log.Println("Warnung: Channel ist nil")
		return
	}

	// Erstelle Channel-Objekt mit Discord-Daten
	channel := buildChannelFromDiscord(discordChannel)

	// Nutze die UpsertChannel-Funktion aus tables
	_, err := tables.UpsertChannel(db, channel)
	if err != nil {
		log.Printf("Fehler beim Upsert des Channels %s: %v", discordChannel.ID, err)
		return
	}

	log.Printf("Channel %s (%s) wurde in die Datenbank synchronisiert", discordChannel.Name, discordChannel.ID)
}

// RemoveChannel löscht einen Channel aus der Datenbank
func RemoveChannel(db *sql.DB, channelID string) {
	if channelID == "" {
		log.Println("Warnung: Channel ID ist leer")
		return
	}

	err := tables.DeleteChannel(db, channelID)
	if err != nil {
		log.Printf("Fehler beim Löschen des Channels %s: %v", channelID, err)
		return
	}

	log.Printf("Channel %s wurde aus der Datenbank gelöscht", channelID)
}

// buildChannelFromDiscord konvertiert einen Discord-Channel in ein Channel-Objekt
func buildChannelFromDiscord(discordChannel *discordgo.Channel) *tables.Channel {
	// Konvertiere Permission Overwrites zu JSON
	permissionOverwrites := ""
	if len(discordChannel.PermissionOverwrites) > 0 {
		data, err := json.Marshal(discordChannel.PermissionOverwrites)
		if err == nil {
			permissionOverwrites = string(data)
		}
	}

	// Extrahiere Snowflake-Timestamp für created_at
	createdAt := extractSnowflakeTimestamp(discordChannel.ID)

	return &tables.Channel{
		ID:                   discordChannel.ID,
		GuildID:              discordChannel.GuildID,
		Name:                 discordChannel.Name,
		Type:                 int(discordChannel.Type),
		Position:             discordChannel.Position,
		Topic:                discordChannel.Topic,
		NSFW:                 discordChannel.NSFW,
		Bitrate:              discordChannel.Bitrate,
		UserLimit:            discordChannel.UserLimit,
		ParentID:             discordChannel.ParentID,
		PermissionOverwrites: permissionOverwrites,
		CreatedAt:            createdAt,
	}
}

// extractSnowflakeTimestamp extrahiert den Timestamp aus einer Discord Snowflake ID
func extractSnowflakeTimestamp(snowflake string) time.Time {
	// Discord Snowflake Format: 64-bit integer
	// Bits 0-41: Milliseconds since Discord Epoch (2015-01-01)
	// Wir nutzen einen einfachen Parser
	var id int64
	_, err := fmt.Sscanf(snowflake, "%d", &id)
	if err != nil {
		return time.Now()
	}

	// Discord Epoch: 2015-01-01 00:00:00 UTC
	discordEpoch := int64(1420070400000)
	timestamp := (id >> 22) + discordEpoch

	return time.Unix(timestamp/1000, (timestamp%1000)*1000000)
}
