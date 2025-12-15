package logging

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"discord-bot-template/internal/database/tables"
	"discord-bot-template/internal/shared/utils/notifications"

	"github.com/bwmarrin/discordgo"
)

const (
	logDir = "logs"
)

// Logger ist die zentrale Logging-Struktur
type Logger struct {
	db      *sql.DB
	session *discordgo.Session
	guildID string
	source  string
}

// NewLogger erstellt einen neuen Logger
func NewLogger(db *sql.DB, session *discordgo.Session, guildID, source string) *Logger {
	return &Logger{
		db:      db,
		session: session,
		guildID: guildID,
		source:  source,
	}
}

// logToDatabase versucht einen Log-Eintrag in die Datenbank zu schreiben
func (l *Logger) logToDatabase(level tables.LogLevel, title, message, stackTrace string) error {
	if l.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	logEntry := &tables.Log{
		GuildID:    l.guildID,
		Level:      level,
		Title:      title,
		Message:    message,
		StackTrace: stackTrace,
		Source:     l.source,
	}

	return tables.InsertLog(l.db, logEntry)
}

// logToFile schreibt einen Log-Eintrag in eine Datei (Fallback)
func (l *Logger) logToFile(level tables.LogLevel, title, message, stackTrace string) error {
	// Erstelle Log-Verzeichnis wenn es nicht existiert
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Erstelle Dateiname basierend auf Datum und Level
	date := time.Now().Format("2006-01-02")
	filename := filepath.Join(logDir, fmt.Sprintf("%s_%s.log", strings.ToLower(string(level)), date))

	// Öffne oder erstelle die Log-Datei
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	// Formatiere Log-Eintrag
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("[%s] [%s] [%s] %s: %s\n", timestamp, level, l.source, title, message)

	if stackTrace != "" {
		logLine += fmt.Sprintf("Stack Trace:\n%s\n", stackTrace)
	}
	logLine += "---\n"

	// Schreibe in Datei
	if _, err := file.WriteString(logLine); err != nil {
		return fmt.Errorf("failed to write to log file: %v", err)
	}

	return nil
}

// LogError loggt einen Error und sendet IMMER eine Notification
func (l *Logger) LogError(title, message, stackTrace string) {
	// Console-Ausgabe (wie bisher)
	if stackTrace != "" {
		log.Printf("[ERROR] [%s] %s: %s\nStack: %s", l.source, title, message, stackTrace)
	} else {
		log.Printf("[ERROR] [%s] %s: %s", l.source, title, message)
	}

	// Versuche in Datenbank zu schreiben
	err := l.logToDatabase(tables.LogLevelError, title, message, stackTrace)
	if err != nil {
		// Fallback auf File-Logging
		log.Printf("[WARNING] Failed to log to database: %v. Using file fallback.", err)
		if fileErr := l.logToFile(tables.LogLevelError, title, message, stackTrace); fileErr != nil {
			log.Printf("[ERROR] Failed to log to file: %v", fileErr)
		}
	}

	// IMMER Error-Notification senden
	if l.session != nil && l.db != nil && l.guildID != "" {
		var fields []*discordgo.MessageEmbedField
		if stackTrace != "" {
			// Kürze Stack Trace wenn zu lang
			shortStack := stackTrace
			if len(stackTrace) > 1000 {
				shortStack = stackTrace[:1000] + "..."
			}
			fields = []*discordgo.MessageEmbedField{
				{
					Name:  "Stack Trace",
					Value: fmt.Sprintf("```\n%s\n```", shortStack),
				},
			}
		}

		if err := notifications.SendErrorNotification(l.session, l.db, l.guildID, title, message, fields); err != nil {
			log.Printf("[ERROR] Failed to send error notification: %v", err)
		}
	}
}

// LogInfo loggt eine Info-Message mit optionaler Notification
func (l *Logger) LogInfo(title, message string, sendNotification bool) {
	// Console-Ausgabe (wie bisher)
	log.Printf("[INFO] [%s] %s: %s", l.source, title, message)

	// Versuche in Datenbank zu schreiben
	err := l.logToDatabase(tables.LogLevelInfo, title, message, "")
	if err != nil {
		// Fallback auf File-Logging
		log.Printf("[WARNING] Failed to log to database: %v. Using file fallback.", err)
		if fileErr := l.logToFile(tables.LogLevelInfo, title, message, ""); fileErr != nil {
			log.Printf("[ERROR] Failed to log to file: %v", fileErr)
		}
	}

	// Optional Info-Notification senden
	if sendNotification && l.session != nil && l.db != nil && l.guildID != "" {
		if err := notifications.SendInfoNotification(l.session, l.db, l.guildID, title, message, nil); err != nil {
			log.Printf("[ERROR] Failed to send info notification: %v", err)
		}
	}
}

// LogWarn loggt eine Warning
func (l *Logger) LogWarn(title, message string) {
	// Console-Ausgabe
	log.Printf("[WARN] [%s] %s: %s", l.source, title, message)

	// Versuche in Datenbank zu schreiben
	err := l.logToDatabase(tables.LogLevelWarn, title, message, "")
	if err != nil {
		// Fallback auf File-Logging
		log.Printf("[WARNING] Failed to log to database: %v. Using file fallback.", err)
		if fileErr := l.logToFile(tables.LogLevelWarn, title, message, ""); fileErr != nil {
			log.Printf("[ERROR] Failed to log to file: %v", fileErr)
		}
	}
}

// LogDebug loggt eine Debug-Message
func (l *Logger) LogDebug(title, message string) {
	// Console-Ausgabe
	log.Printf("[DEBUG] [%s] %s: %s", l.source, title, message)

	// Versuche in Datenbank zu schreiben
	err := l.logToDatabase(tables.LogLevelDebug, title, message, "")
	if err != nil {
		// Fallback auf File-Logging
		log.Printf("[WARNING] Failed to log to database: %v. Using file fallback.", err)
		if fileErr := l.logToFile(tables.LogLevelDebug, title, message, ""); fileErr != nil {
			log.Printf("[ERROR] Failed to log to file: %v", fileErr)
		}
	}
}
