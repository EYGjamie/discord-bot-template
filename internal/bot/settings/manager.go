package settings

import (
	"database/sql"
	"log"
	"strconv"
	"sync"
	"time"

	"discord-bot-template/internal/database/tables"
)

// Manager verwaltet Bot-Einstellungen mit Caching
type Manager struct {
	db    *sql.DB
	cache map[string]*tables.BotSetting
	mu    sync.RWMutex
}

// GetDB gibt die Datenbank-Verbindung zurück
func (m *Manager) GetDB() *sql.DB {
	return m.db
}

// NewManager erstellt einen neuen Settings Manager
func NewManager(db *sql.DB) *Manager {
	m := &Manager{
		db:    db,
		cache: make(map[string]*tables.BotSetting),
	}

	// Initial cache laden
	m.RefreshCache()

	// Automatisches Cache-Refresh alle 30 Sekunden
	go m.autoRefresh()

	return m
}

// autoRefresh aktualisiert den Cache periodisch
func (m *Manager) autoRefresh() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.RefreshCache()
	}
}

// RefreshCache lädt alle Einstellungen neu
func (m *Manager) RefreshCache() {
	settings, err := tables.GetAllBotSettings(m.db)
	if err != nil {
		log.Printf("Fehler beim Laden der Bot-Einstellungen: %v", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Cache aktualisieren
	for _, setting := range settings {
		m.cache[setting.Key] = setting
	}

	log.Printf("Cache aktualisiert: %d Einstellungen geladen", len(settings))
}

// GetString gibt einen String-Wert zurück
func (m *Manager) GetString(key, defaultValue string) string {
	m.mu.RLock()
	setting, exists := m.cache[key]
	m.mu.RUnlock()

	if !exists || !setting.Enabled {
		return defaultValue
	}

	return setting.Value
}

// GetBool gibt einen Boolean-Wert zurück
func (m *Manager) GetBool(key string, defaultValue bool) bool {
	m.mu.RLock()
	setting, exists := m.cache[key]
	m.mu.RUnlock()

	if !exists || !setting.Enabled {
		return defaultValue
	}

	value, err := strconv.ParseBool(setting.Value)
	if err != nil {
		log.Printf("Fehler beim Parsen von Bool-Einstellung %s: %v", key, err)
		return defaultValue
	}

	return value
}

// GetInt gibt einen Integer-Wert zurück
func (m *Manager) GetInt(key string, defaultValue int) int {
	m.mu.RLock()
	setting, exists := m.cache[key]
	m.mu.RUnlock()

	if !exists || !setting.Enabled {
		return defaultValue
	}

	value, err := strconv.Atoi(setting.Value)
	if err != nil {
		log.Printf("Fehler beim Parsen von Int-Einstellung %s: %v", key, err)
		return defaultValue
	}

	return value
}

// IsEnabled prüft ob eine Einstellung aktiviert ist
func (m *Manager) IsEnabled(key string) bool {
	m.mu.RLock()
	setting, exists := m.cache[key]
	m.mu.RUnlock()

	if !exists {
		return false
	}

	return setting.Enabled
}

// SetString speichert einen String-Wert
func (m *Manager) SetString(key, value string, enabled bool) error {
	setting := &tables.BotSetting{
		Key:     key,
		Value:   value,
		Type:    "string",
		Enabled: enabled,
	}

	_, err := tables.UpsertBotSetting(m.db, setting)
	if err != nil {
		return err
	}

	// Cache aktualisieren
	m.mu.Lock()
	m.cache[key] = setting
	m.mu.Unlock()

	log.Printf("Einstellung %s gespeichert: %s (enabled: %v)", key, value, enabled)
	return nil
}

// SetBool speichert einen Boolean-Wert
func (m *Manager) SetBool(key string, value bool, enabled bool) error {
	setting := &tables.BotSetting{
		Key:     key,
		Value:   strconv.FormatBool(value),
		Type:    "bool",
		Enabled: enabled,
	}

	_, err := tables.UpsertBotSetting(m.db, setting)
	if err != nil {
		return err
	}

	// Cache aktualisieren
	m.mu.Lock()
	m.cache[key] = setting
	m.mu.Unlock()

	log.Printf("Einstellung %s gespeichert: %v (enabled: %v)", key, value, enabled)
	return nil
}

// SetEnabled aktiviert oder deaktiviert eine Einstellung
func (m *Manager) SetEnabled(key string, enabled bool) error {
	err := tables.UpdateBotSettingEnabled(m.db, key, enabled)
	if err != nil {
		return err
	}

	// Cache aktualisieren
	m.mu.Lock()
	if setting, exists := m.cache[key]; exists {
		setting.Enabled = enabled
	}
	m.mu.Unlock()

	log.Printf("Einstellung %s %s", key, map[bool]string{true: "aktiviert", false: "deaktiviert"}[enabled])
	return nil
}

// Delete löscht eine Einstellung
func (m *Manager) Delete(key string) error {
	err := tables.DeleteBotSetting(m.db, key)
	if err != nil {
		return err
	}

	// Aus Cache entfernen
	m.mu.Lock()
	delete(m.cache, key)
	m.mu.Unlock()

	log.Printf("Einstellung %s gelöscht", key)
	return nil
}

// GetAll gibt alle Einstellungen zurück
func (m *Manager) GetAll() map[string]*tables.BotSetting {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Kopie erstellen
	result := make(map[string]*tables.BotSetting)
	for k, v := range m.cache {
		result[k] = v
	}

	return result
}
