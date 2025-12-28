package services

import (
	"database/sql"
	"discord-bot-template/backend/handlers"
	"log"
	"time"
)

// DiscordStatsScheduler plant automatische Statistik-Erfassungen
type DiscordStatsScheduler struct {
	statsHandler *handlers.DiscordStatsHandler
	stopChan     chan bool
}

// NewDiscordStatsScheduler erstellt einen neuen Scheduler
func NewDiscordStatsScheduler(db *sql.DB) *DiscordStatsScheduler {
	return &DiscordStatsScheduler{
		statsHandler: handlers.NewDiscordStatsHandler(db),
		stopChan:     make(chan bool),
	}
}

// Start startet den Scheduler
func (s *DiscordStatsScheduler) Start() {
	log.Println("Discord Statistics Scheduler started")

	// Führe sofort eine initiale Statistik-Erfassung durch
	go func() {
		time.Sleep(5 * time.Second) // Warte kurz bis alle Services hochgefahren sind
		if _, err := s.statsHandler.FetchAndSaveStatistics("scheduled_initial"); err != nil {
			log.Printf("Error in initial statistics collection: %v", err)
		}
	}()

	// Starte den Haupt-Scheduler-Loop
	go s.schedulerLoop()
}

// Stop stoppt den Scheduler
func (s *DiscordStatsScheduler) Stop() {
	log.Println("Stopping Discord Statistics Scheduler...")
	s.stopChan <- true
}

// schedulerLoop ist die Hauptschleife des Schedulers
func (s *DiscordStatsScheduler) schedulerLoop() {
	// Stündliche Ausführung
	for {
		now := time.Now()
		nextRun := s.calculateNextHourRun(now)
		duration := nextRun.Sub(now)

		log.Printf("Next statistics collection scheduled for: %s (in %v)", nextRun.Format(time.RFC3339), duration)

		select {
		case <-time.After(duration):
			// Zeit für die nächste Statistik-Erfassung
			log.Printf("Running scheduled statistics collection at %s", time.Now().Format(time.RFC3339))
			if _, err := s.statsHandler.FetchAndSaveStatistics("scheduled"); err != nil {
				log.Printf("Error in scheduled statistics collection: %v", err)
			}

		case <-s.stopChan:
			log.Println("Discord Statistics Scheduler stopped")
			return
		}
	}
}

// calculateNextHourRun berechnet die nächste volle Stunde
func (s *DiscordStatsScheduler) calculateNextHourRun(now time.Time) time.Time {
	// Nächste volle Stunde
	nextRun := now.Truncate(time.Hour).Add(time.Hour)
	return nextRun
}

// RunNow führt sofort eine Statistik-Erfassung durch (für manuellen Aufruf)
func (s *DiscordStatsScheduler) RunNow() error {
	_, err := s.statsHandler.FetchAndSaveStatistics("manual_trigger")
	return err
}
