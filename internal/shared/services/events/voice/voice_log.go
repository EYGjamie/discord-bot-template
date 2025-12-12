package voice

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"discord-bot-template/internal/database/tables"

	"github.com/bwmarrin/discordgo"
)

// VoiceSessionState hält den aktuellen State einer Voice Session für einen User
type VoiceSessionState struct {
	LogID               int64
	IsMuted             bool
	IsDeafened          bool
	IsStreaming         bool
	LastMutedChange     time.Time
	LastDeafenedChange  time.Time
	LastStreamingChange time.Time
	MutedDuration       int64 // in Sekunden
	DeafenDuration      int64 // in Sekunden
	StreamDuration      int64 // in Sekunden
}

var (
	// Map zum Speichern der aktiven Voice Sessions: UserID -> VoiceSessionState
	activeSessions = make(map[string]*VoiceSessionState)
	sessionsMutex  sync.RWMutex
)

// LogVoiceStateUpdate verarbeitet Voice State Updates und loggt die Voice-Aktivität
func LogVoiceStateUpdate(db *sql.DB, voiceState *discordgo.VoiceStateUpdate) {
	if voiceState.UserID == "" {
		log.Println("Warnung: VoiceStateUpdate ohne UserID erhalten")
		return
	}

	// Prüfe ob der User einen Channel betritt, verlässt oder den State ändert
	beforeChannelID := ""
	if voiceState.BeforeUpdate != nil {
		beforeChannelID = voiceState.BeforeUpdate.ChannelID
	}
	afterChannelID := voiceState.ChannelID

	// Fall 1: User betritt einen Voice Channel
	if beforeChannelID == "" && afterChannelID != "" {
		handleVoiceJoin(db, voiceState)
		return
	}

	// Fall 2: User verlässt einen Voice Channel
	if beforeChannelID != "" && afterChannelID == "" {
		handleVoiceLeave(db, voiceState)
		return
	}

	// Fall 3: User wechselt zwischen Voice Channels
	if beforeChannelID != "" && afterChannelID != "" && beforeChannelID != afterChannelID {
		handleVoiceLeave(db, voiceState)
		handleVoiceJoin(db, voiceState)
		return
	}

	// Fall 4: User ändert nur den State (Mute/Deafen/Stream) im selben Channel
	if beforeChannelID != "" && afterChannelID != "" && beforeChannelID == afterChannelID {
		handleVoiceStateChange(db, voiceState)
		return
	}
}

// handleVoiceJoin verarbeitet das Betreten eines Voice Channels
func handleVoiceJoin(db *sql.DB, voiceState *discordgo.VoiceStateUpdate) {
	// Erstelle neuen Log-Eintrag in der Datenbank
	logID, err := tables.InsertUserVoiceLog(db, voiceState.UserID, voiceState.ChannelID)
	if err != nil {
		log.Printf("Fehler beim Erstellen des Voice-Logs für User %s: %v", voiceState.UserID, err)
		return
	}

	// Initialisiere Session State
	now := time.Now()
	session := &VoiceSessionState{
		LogID:               logID,
		IsMuted:             voiceState.Mute || voiceState.SelfMute,
		IsDeafened:          voiceState.Deaf || voiceState.SelfDeaf,
		IsStreaming:         voiceState.SelfStream,
		LastMutedChange:     now,
		LastDeafenedChange:  now,
		LastStreamingChange: now,
		MutedDuration:       0,
		DeafenDuration:      0,
		StreamDuration:      0,
	}

	// Speichere Session
	sessionsMutex.Lock()
	activeSessions[voiceState.UserID] = session
	sessionsMutex.Unlock()

	log.Printf("User %s hat Voice Channel %s betreten (LogID: %d)", voiceState.UserID, voiceState.ChannelID, logID)
}

// handleVoiceLeave verarbeitet das Verlassen eines Voice Channels
func handleVoiceLeave(db *sql.DB, voiceState *discordgo.VoiceStateUpdate) {
	sessionsMutex.Lock()
	session, exists := activeSessions[voiceState.UserID]
	if !exists {
		sessionsMutex.Unlock()
		log.Printf("Warnung: Keine aktive Session für User %s gefunden", voiceState.UserID)
		return
	}

	// Berechne finale Durations basierend auf BeforeUpdate
	if voiceState.BeforeUpdate != nil {
		now := time.Now()

		// Addiere finale Zeit für Mute
		if voiceState.BeforeUpdate.Mute || voiceState.BeforeUpdate.SelfMute {
			session.MutedDuration += int64(now.Sub(session.LastMutedChange).Seconds())
		}

		// Addiere finale Zeit für Deafen
		if voiceState.BeforeUpdate.Deaf || voiceState.BeforeUpdate.SelfDeaf {
			session.DeafenDuration += int64(now.Sub(session.LastDeafenedChange).Seconds())
		}

		// Addiere finale Zeit für Streaming
		if voiceState.BeforeUpdate.SelfStream {
			session.StreamDuration += int64(now.Sub(session.LastStreamingChange).Seconds())
		}
	}

	// Aktualisiere Datenbank mit finalen Werten
	err := tables.UpdateUserVoiceLogOnLeave(db, session.LogID, session.MutedDuration, session.DeafenDuration, session.StreamDuration)
	if err != nil {
		log.Printf("Fehler beim Aktualisieren des Voice-Logs (ID: %d): %v", session.LogID, err)
	} else {
		log.Printf("User %s hat Voice Channel verlassen (LogID: %d, Total: %ds, Muted: %ds, Deafened: %ds, Streaming: %ds)",
			voiceState.UserID, session.LogID, session.MutedDuration+session.DeafenDuration+session.StreamDuration,
			session.MutedDuration, session.DeafenDuration, session.StreamDuration)
	}

	// Entferne Session
	delete(activeSessions, voiceState.UserID)
	sessionsMutex.Unlock()
}

// handleVoiceStateChange verarbeitet State-Änderungen im selben Channel (Mute/Deafen/Stream)
func handleVoiceStateChange(db *sql.DB, voiceState *discordgo.VoiceStateUpdate) {
	sessionsMutex.Lock()
	session, exists := activeSessions[voiceState.UserID]
	if !exists {
		sessionsMutex.Unlock()
		log.Printf("Warnung: Keine aktive Session für User %s gefunden", voiceState.UserID)
		return
	}
	sessionsMutex.Unlock()

	if voiceState.BeforeUpdate == nil {
		return
	}

	now := time.Now()
	updated := false

	// Prüfe Mute-Änderung
	wasMuted := voiceState.BeforeUpdate.Mute || voiceState.BeforeUpdate.SelfMute
	isMuted := voiceState.Mute || voiceState.SelfMute

	if wasMuted != isMuted {
		if wasMuted {
			// User hat Mute deaktiviert
			duration := int64(now.Sub(session.LastMutedChange).Seconds())
			session.MutedDuration += duration
			log.Printf("User %s hat Mute deaktiviert (Dauer: %ds)", voiceState.UserID, duration)
		} else {
			// User hat Mute aktiviert
			log.Printf("User %s hat Mute aktiviert", voiceState.UserID)
		}
		session.IsMuted = isMuted
		session.LastMutedChange = now
		updated = true
	}

	// Prüfe Deafen-Änderung
	wasDeafened := voiceState.BeforeUpdate.Deaf || voiceState.BeforeUpdate.SelfDeaf
	isDeafened := voiceState.Deaf || voiceState.SelfDeaf

	if wasDeafened != isDeafened {
		if wasDeafened {
			// User hat Deafen deaktiviert
			duration := int64(now.Sub(session.LastDeafenedChange).Seconds())
			session.DeafenDuration += duration
			log.Printf("User %s hat Deafen deaktiviert (Dauer: %ds)", voiceState.UserID, duration)
		} else {
			// User hat Deafen aktiviert
			log.Printf("User %s hat Deafen aktiviert", voiceState.UserID)
		}
		session.IsDeafened = isDeafened
		session.LastDeafenedChange = now
		updated = true
	}

	// Prüfe Streaming-Änderung
	wasStreaming := voiceState.BeforeUpdate.SelfStream
	isStreaming := voiceState.SelfStream

	if wasStreaming != isStreaming {
		if wasStreaming {
			// User hat Streaming deaktiviert
			duration := int64(now.Sub(session.LastStreamingChange).Seconds())
			session.StreamDuration += duration
			log.Printf("User %s hat Screenshare deaktiviert (Dauer: %ds)", voiceState.UserID, duration)
		} else {
			// User hat Streaming aktiviert
			log.Printf("User %s hat Screenshare aktiviert", voiceState.UserID)
		}
		session.IsStreaming = isStreaming
		session.LastStreamingChange = now
		updated = true
	}

	// Aktualisiere Datenbank wenn sich etwas geändert hat
	if updated {
		sessionsMutex.Lock()
		err := tables.UpdateUserVoiceLogDurations(db, session.LogID, session.MutedDuration, session.DeafenDuration, session.StreamDuration)
		sessionsMutex.Unlock()

		if err != nil {
			log.Printf("Fehler beim Aktualisieren der Voice-Log Durations (ID: %d): %v", session.LogID, err)
		}
	}
}
