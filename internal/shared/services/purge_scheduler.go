package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"discord-bot-template/internal/bot/settings"
	"discord-bot-template/internal/database/tables"
	"discord-bot-template/internal/shared/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// PurgeScheduler verwaltet geplante Channel-Purges
type PurgeScheduler struct {
	session      *discordgo.Session
	db           *sql.DB
	purgeManager *settings.PurgeManager
	stopChan     chan bool
	isRunning    bool
}

// NewPurgeScheduler erstellt einen neuen Purge Scheduler
func NewPurgeScheduler(session *discordgo.Session, db *sql.DB) *PurgeScheduler {
	return &PurgeScheduler{
		session:      session,
		db:           db,
		purgeManager: settings.NewPurgeManager(db),
		stopChan:     make(chan bool),
		isRunning:    false,
	}
}

// Start startet den Scheduler
func (ps *PurgeScheduler) Start() {
	if ps.isRunning {
		log.Println("Purge Scheduler läuft bereits")
		return
	}

	ps.isRunning = true
	log.Println("Purge Scheduler gestartet")

	go ps.run()
}

// Stop stoppt den Scheduler
func (ps *PurgeScheduler) Stop() {
	if !ps.isRunning {
		return
	}

	ps.stopChan <- true
	ps.isRunning = false
	log.Println("Purge Scheduler gestoppt")
}

// run führt den Scheduler-Loop aus
func (ps *PurgeScheduler) run() {
	// Prüfe alle 60 Sekunden
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Erste Prüfung sofort beim Start
	ps.checkAndExecutePurges()

	for {
		select {
		case <-ticker.C:
			ps.checkAndExecutePurges()
		case <-ps.stopChan:
			return
		}
	}
}

// checkAndExecutePurges prüft und führt fällige Purges aus
func (ps *PurgeScheduler) checkAndExecutePurges() {
	settings, err := ps.purgeManager.GetAllEnabledPurgeSettings()
	if err != nil {
		log.Printf("Fehler beim Abrufen der Purge-Einstellungen: %v", err)
		return
	}

	currentTime := time.Now()
	currentHour := currentTime.Hour()
	currentMinute := currentTime.Minute()

	for _, setting := range settings {
		// Parse schedule time
		var scheduleHour, scheduleMinute int
		_, err := fmt.Sscanf(setting.ScheduleTime, "%d:%d", &scheduleHour, &scheduleMinute)
		if err != nil {
			log.Printf("Fehler beim Parsen der Zeit %s: %v", setting.ScheduleTime, err)
			continue
		}

		// Prüfe ob die aktuelle Zeit mit der geplanten Zeit übereinstimmt
		if currentHour == scheduleHour && currentMinute == scheduleMinute {
			// Prüfe ob bereits heute ausgeführt wurde
			if !setting.LastRun.IsZero() {
				lastRunDay := setting.LastRun.Day()
				currentDay := currentTime.Day()
				if lastRunDay == currentDay {
					// Bereits heute ausgeführt, überspringe
					continue
				}
			}

			// Führe Purge aus
			go ps.executePurge(setting)
		}
	}
}

// executePurge führt einen Purge aus
func (ps *PurgeScheduler) executePurge(setting *tables.ChannelPurgeSetting) {
	logger := logging.NewLogger(ps.db, ps.session, setting.GuildID, "scheduler.purge")

	log.Printf("Führe geplanten Purge aus für Channel %s in Guild %s", setting.ChannelID, setting.GuildID)

	// Führe den Purge aus
	deletedCount, err := ps.purgeChannel(setting.ChannelID)
	if err != nil {
		logger.LogError("Scheduled Purge Failed",
			fmt.Sprintf("Failed to purge channel %s: %v", setting.ChannelID, err),
			"")
		log.Printf("Fehler beim Purge von Channel %s: %v", setting.ChannelID, err)
		return
	}

	// Aktualisiere last_run
	err = ps.purgeManager.UpdateLastRun(setting.GuildID, setting.ChannelID)
	if err != nil {
		log.Printf("Fehler beim Aktualisieren von last_run: %v", err)
	}

	// Logge den Erfolg
	logger.LogInfo("Scheduled Purge Executed",
		fmt.Sprintf("Successfully purged %d messages from channel %s", deletedCount, setting.ChannelID),
		false)

	log.Printf("Geplanter Purge erfolgreich: %d Nachrichten gelöscht aus Channel %s", deletedCount, setting.ChannelID)
}

// purgeChannel löscht alle Nachrichten in einem Channel
func (ps *PurgeScheduler) purgeChannel(channelID string) (int, error) {
	deletedCount := 0
	var beforeID string

	for {
		// Hole Nachrichten in Batches von 100
		messages, err := ps.session.ChannelMessages(channelID, 100, beforeID, "", "")
		if err != nil {
			return deletedCount, err
		}

		if len(messages) == 0 {
			break
		}

		// Nachrichten nach Alter filtern (nur Nachrichten < 14 Tage können bulk-deleted werden)
		var bulkDeleteIDs []string
		var oldMessages []*discordgo.Message

		twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)

		for _, msg := range messages {
			if msg.Timestamp.After(twoWeeksAgo) {
				bulkDeleteIDs = append(bulkDeleteIDs, msg.ID)
			} else {
				oldMessages = append(oldMessages, msg)
			}
		}

		// Bulk-Delete für neue Nachrichten
		if len(bulkDeleteIDs) > 0 {
			if len(bulkDeleteIDs) == 1 {
				// Einzelne Nachricht löschen
				err = ps.session.ChannelMessageDelete(channelID, bulkDeleteIDs[0])
				if err != nil {
					return deletedCount, err
				}
				deletedCount++
			} else {
				// Bulk-Delete für mehrere Nachrichten
				err = ps.session.ChannelMessagesBulkDelete(channelID, bulkDeleteIDs)
				if err != nil {
					return deletedCount, err
				}
				deletedCount += len(bulkDeleteIDs)
			}
		}

		// Alte Nachrichten einzeln löschen
		for _, msg := range oldMessages {
			err = ps.session.ChannelMessageDelete(channelID, msg.ID)
			if err != nil {
				return deletedCount, err
			}
			deletedCount++
			time.Sleep(100 * time.Millisecond) // Rate limiting
		}

		// Wenn keine weiteren Nachrichten vorhanden sind, stop
		if len(messages) < 100 {
			break
		}

		beforeID = messages[len(messages)-1].ID
		time.Sleep(500 * time.Millisecond) // Rate limiting zwischen Batches
	}

	return deletedCount, nil
}
