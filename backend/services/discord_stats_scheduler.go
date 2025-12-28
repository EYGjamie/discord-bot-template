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
	// Definiere die Zielzeiten: 6:00, 12:00, 18:00, 22:00
	targetHours := []int{6, 12, 18, 22}

	for {
		now := time.Now()
		nextRun := s.calculateNextRun(now, targetHours)
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

// calculateNextRun berechnet die nächste Ausführungszeit
func (s *DiscordStatsScheduler) calculateNextRun(now time.Time, targetHours []int) time.Time {
	// Erstelle Zeit für heute
	year, month, day := now.Date()
	location := now.Location()

	// Finde die nächste Zielstunde
	currentHour := now.Hour()
	var nextHour int
	nextDay := false

	for _, hour := range targetHours {
		if hour > currentHour {
			nextHour = hour
			break
		}
	}

	// Wenn keine Stunde mehr heute übrig ist, nimm die erste von morgen
	if nextHour == 0 {
		nextHour = targetHours[0]
		nextDay = true
	}

	// Wenn wir bereits in der aktuellen Zielstunde sind, prüfe die Minuten
	if nextHour == currentHour && !nextDay {
		// Prüfe ob wir bereits nach Minute 0 sind
		if now.Minute() >= 1 {
			// Gehe zur nächsten Zielstunde
			found := false
			for i, hour := range targetHours {
				if hour == currentHour && i+1 < len(targetHours) {
					nextHour = targetHours[i+1]
					found = true
					break
				}
			}
			if !found {
				nextHour = targetHours[0]
				nextDay = true
			}
		}
	}

	// Erstelle die nächste Ausführungszeit
	nextRun := time.Date(year, month, day, nextHour, 0, 0, 0, location)
	if nextDay {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	return nextRun
}

// RunNow führt sofort eine Statistik-Erfassung durch (für manuellen Aufruf)
func (s *DiscordStatsScheduler) RunNow() error {
	_, err := s.statsHandler.FetchAndSaveStatistics("manual_trigger")
	return err
}
