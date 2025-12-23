package settings

import (
	"database/sql"
	"discord-bot-template/internal/database/tables"
	"log"
)

// PurgeManager verwaltet Channel-Purge-Einstellungen
type PurgeManager struct {
	db *sql.DB
}

// NewPurgeManager erstellt einen neuen Purge Manager
func NewPurgeManager(db *sql.DB) *PurgeManager {
	return &PurgeManager{
		db: db,
	}
}

// SetPurgeSchedule richtet einen geplanten Purge ein
func (pm *PurgeManager) SetPurgeSchedule(guildID, channelID, scheduleTime string) error {
	setting := &tables.ChannelPurgeSetting{
		GuildID:      guildID,
		ChannelID:    channelID,
		ScheduleTime: scheduleTime,
		Enabled:      true,
	}

	_, err := tables.InsertChannelPurgeSetting(pm.db, setting)
	if err != nil {
		log.Printf("Error setting purge schedule: %v", err)
		return err
	}

	return nil
}

// RemovePurgeSchedule entfernt einen geplanten Purge
func (pm *PurgeManager) RemovePurgeSchedule(guildID, channelID string) error {
	err := tables.DeleteChannelPurgeSetting(pm.db, guildID, channelID)
	if err != nil {
		log.Printf("Error removing purge schedule: %v", err)
		return err
	}

	return nil
}

// TogglePurgeSchedule aktiviert/deaktiviert einen geplanten Purge
func (pm *PurgeManager) TogglePurgeSchedule(guildID, channelID string, enabled bool) error {
	err := tables.UpdatePurgeSettingEnabled(pm.db, guildID, channelID, enabled)
	if err != nil {
		log.Printf("Error toggling purge schedule: %v", err)
		return err
	}

	return nil
}

// GetPurgeSetting gibt die Purge-Einstellung für einen Channel zurück
func (pm *PurgeManager) GetPurgeSetting(guildID, channelID string) (*tables.ChannelPurgeSetting, error) {
	setting, err := tables.GetChannelPurgeSetting(pm.db, guildID, channelID)
	if err != nil {
		log.Printf("Error getting purge setting: %v", err)
		return nil, err
	}

	return setting, nil
}

// GetAllPurgeSettings gibt alle Purge-Einstellungen für eine Guild zurück
func (pm *PurgeManager) GetAllPurgeSettings(guildID string) ([]*tables.ChannelPurgeSetting, error) {
	settings, err := tables.GetGuildPurgeSettings(pm.db, guildID)
	if err != nil {
		log.Printf("Error getting all purge settings: %v", err)
		return nil, err
	}

	return settings, nil
}

// GetAllEnabledPurgeSettings gibt alle aktivierten Purge-Einstellungen zurück
func (pm *PurgeManager) GetAllEnabledPurgeSettings() ([]*tables.ChannelPurgeSetting, error) {
	settings, err := tables.GetAllEnabledPurgeSettings(pm.db)
	if err != nil {
		log.Printf("Error getting all enabled purge settings: %v", err)
		return nil, err
	}

	return settings, nil
}

// UpdateLastRun aktualisiert die letzte Ausführungszeit
func (pm *PurgeManager) UpdateLastRun(guildID, channelID string) error {
	err := tables.UpdatePurgeLastRun(pm.db, guildID, channelID)
	if err != nil {
		log.Printf("Error updating last run: %v", err)
		return err
	}

	return nil
}
