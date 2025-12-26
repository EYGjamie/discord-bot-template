package voice

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"discord-bot-template/shared/database/tables"
	"discord-bot-template/bot/utils/logging"

	"github.com/bwmarrin/discordgo"
)

// HandleCreateVoiceJoin behandelt das Beitreten eines Create-Voice-Channels
func HandleCreateVoiceJoin(s *discordgo.Session, db *sql.DB, voiceState *discordgo.VoiceStateUpdate) error {
	if voiceState.ChannelID == "" {
		return nil
	}

	logger := logging.NewLogger(db, s, voiceState.GuildID, "voice.create_voice")
	logger.LogInfo("Voice Join Event", fmt.Sprintf("User %s joined channel %s", voiceState.UserID, voiceState.ChannelID), false)

	// Prüfe ob der Channel ein Create-Voice-Channel ist
	setting, err := tables.GetCreateVoiceSettingByChannelID(db, voiceState.ChannelID)
	if err != nil {
		logger.LogError("Fehler beim Abrufen der Create-Voice-Settings", err.Error(), "")
		return err
	}

	if setting == nil {
		logger.LogInfo("Kein Create-Voice-Channel", fmt.Sprintf("Channel %s ist kein Create-Voice-Channel", voiceState.ChannelID), false)
		return nil // Kein Create-Voice-Channel
	}

	logger.LogInfo("Create-Voice-Channel erkannt", fmt.Sprintf("Erstelle temporären Channel für User %s", voiceState.UserID), false)

	// Hole den originalen Channel um die Kategorie und Position zu bekommen
	originalChannel, err := s.Channel(voiceState.ChannelID)
	if err != nil {
		logger.LogError("Fehler beim Abrufen des Original-Channels", err.Error(), "")
		return err
	}

	// Erstelle einen neuen temporären Voice Channel
	channelName := fmt.Sprintf("%s's Channel", voiceState.Member.User.Username)
	newChannel, err := s.GuildChannelCreateComplex(voiceState.GuildID, discordgo.GuildChannelCreateData{
		Name:      channelName,
		Type:      discordgo.ChannelTypeGuildVoice,
		ParentID:  originalChannel.ParentID,
		Position:  originalChannel.Position + 1,
		UserLimit: setting.DefaultUserLimit,
	})

	if err != nil {
		logger.LogError("Fehler beim Erstellen des temporären Channels", err.Error(), "")
		return err
	}

	// Speichere den temporären Channel in der Datenbank
	tempChannel := &tables.TemporaryVoiceChannel{
		GuildID:       voiceState.GuildID,
		ChannelID:     newChannel.ID,
		OwnerID:       voiceState.UserID,
		CreateVoiceID: setting.ID,
		BlockedUsers:  "[]",
	}

	_, err = tables.InsertTemporaryVoiceChannel(db, tempChannel)
	if err != nil {
		logger.LogError("Fehler beim Speichern des temporären Channels", err.Error(), "")
		// Channel trotzdem weiter nutzen
	}

	// Bewege den User in den neuen Channel
	err = s.GuildMemberMove(voiceState.GuildID, voiceState.UserID, &newChannel.ID)
	if err != nil {
		logger.LogError("Fehler beim Bewegen des Users", err.Error(), "")
		// Versuche den Channel zu löschen
		s.ChannelDelete(newChannel.ID)
		tables.DeleteTemporaryVoiceChannel(db, newChannel.ID)
		return err
	}

	logger.LogInfo("Temporärer Channel erstellt", fmt.Sprintf("User: %s, Channel: %s", voiceState.UserID, newChannel.ID), false)
	return nil
}

// HandleTemporaryVoiceLeave behandelt das Verlassen eines temporären Channels
func HandleTemporaryVoiceLeave(s *discordgo.Session, db *sql.DB, oldChannelID, guildID string) error {
	if oldChannelID == "" {
		return nil
	}

	// Prüfe ob der alte Channel ein temporärer Channel ist
	tempChannel, err := tables.GetTemporaryVoiceChannelByChannelID(db, oldChannelID)
	if err != nil {
		return err
	}

	if tempChannel == nil {
		return nil // Kein temporärer Channel
	}

	// Prüfe ob noch User im Channel sind
	_, err = s.Channel(oldChannelID)
	if err != nil {
		// Channel existiert nicht mehr, lösche aus DB
		tables.DeleteTemporaryVoiceChannel(db, oldChannelID)
		return nil
	}

	// Hole die Guild um die VoiceStates zu bekommen
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return err
	}

	// Zähle User im Channel
	usersInChannel := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == oldChannelID {
			usersInChannel++
		}
	}

	// Wenn keine User mehr im Channel sind, lösche ihn
	if usersInChannel == 0 {
		logger := logging.NewLogger(db, s, guildID, "voice.create_voice")

		_, err := s.ChannelDelete(oldChannelID)
		if err != nil {
			logger.LogError("Fehler beim Löschen des temporären Channels", err.Error(), "")
		} else {
			logger.LogInfo("Temporärer Channel gelöscht", fmt.Sprintf("Channel: %s", oldChannelID), false)
		}

		// Lösche aus Datenbank
		tables.DeleteTemporaryVoiceChannel(db, oldChannelID)
	}

	return nil
}

// CheckUserBlocked prüft ob ein User von einem temporären Channel blockiert ist
func CheckUserBlocked(s *discordgo.Session, db *sql.DB, channelID, userID string) (bool, error) {
	tempChannel, err := tables.GetTemporaryVoiceChannelByChannelID(db, channelID)
	if err != nil || tempChannel == nil {
		return false, err
	}

	var blockedUsers []string
	err = json.Unmarshal([]byte(tempChannel.BlockedUsers), &blockedUsers)
	if err != nil {
		return false, err
	}

	for _, blocked := range blockedUsers {
		if blocked == userID {
			return true, nil
		}
	}

	return false, nil
}

// KickUserFromChannel kickt einen User aus einem temporären Channel
func KickUserFromChannel(s *discordgo.Session, db *sql.DB, guildID, channelID, userID, kickerID string) error {
	tempChannel, err := tables.GetTemporaryVoiceChannelByChannelID(db, channelID)
	if err != nil {
		return err
	}

	if tempChannel == nil {
		return fmt.Errorf("kein temporärer Channel")
	}

	// Prüfe ob der Kicker der Owner ist
	if tempChannel.OwnerID != kickerID {
		return fmt.Errorf("nur der Owner kann User kicken")
	}

	// Bewege den User in keinen Channel (disconnect)
	err = s.GuildMemberMove(guildID, userID, nil)
	if err != nil {
		return err
	}

	logger := logging.NewLogger(db, s, guildID, "voice.create_voice")
	logger.LogInfo("User gekickt", fmt.Sprintf("Channel: %s, User: %s, Kicker: %s", channelID, userID, kickerID), false)

	return nil
}

// BlockUserFromChannel blockiert einen User von einem temporären Channel
func BlockUserFromChannel(s *discordgo.Session, db *sql.DB, guildID, channelID, userID, blockerID string) error {
	tempChannel, err := tables.GetTemporaryVoiceChannelByChannelID(db, channelID)
	if err != nil {
		return err
	}

	if tempChannel == nil {
		return fmt.Errorf("kein temporärer Channel")
	}

	// Prüfe ob der Blocker der Owner ist
	if tempChannel.OwnerID != blockerID {
		return fmt.Errorf("nur der Owner kann User blockieren")
	}

	// Parse blocked users
	var blockedUsers []string
	err = json.Unmarshal([]byte(tempChannel.BlockedUsers), &blockedUsers)
	if err != nil {
		return err
	}

	// Prüfe ob User bereits blockiert ist
	for _, blocked := range blockedUsers {
		if blocked == userID {
			return fmt.Errorf("user ist bereits blockiert")
		}
	}

	// Füge User zur Blocklist hinzu
	blockedUsers = append(blockedUsers, userID)
	blockedJSON, err := json.Marshal(blockedUsers)
	if err != nil {
		return err
	}

	err = tables.UpdateTemporaryVoiceChannelBlockedUsers(db, channelID, string(blockedJSON))
	if err != nil {
		return err
	}

	// Setze Channel Permission Override
	err = s.ChannelPermissionSet(channelID, userID, discordgo.PermissionOverwriteTypeMember, 0, discordgo.PermissionVoiceConnect)
	if err != nil {
		return err
	}

	// Wenn User im Channel ist, kicke ihn
	guild, err := s.State.Guild(guildID)
	if err == nil {
		for _, vs := range guild.VoiceStates {
			if vs.ChannelID == channelID && vs.UserID == userID {
				s.GuildMemberMove(guildID, userID, nil)
				break
			}
		}
	}

	logger := logging.NewLogger(db, s, guildID, "voice.create_voice")
	logger.LogInfo("User blockiert", fmt.Sprintf("Channel: %s, User: %s, Blocker: %s", channelID, userID, blockerID), false)

	return nil
}

// UnblockUserFromChannel hebt die Blockierung eines Users auf
func UnblockUserFromChannel(s *discordgo.Session, db *sql.DB, guildID, channelID, userID, unblockerID string) error {
	tempChannel, err := tables.GetTemporaryVoiceChannelByChannelID(db, channelID)
	if err != nil {
		return err
	}

	if tempChannel == nil {
		return fmt.Errorf("kein temporärer Channel")
	}

	// Prüfe ob der Unblocker der Owner ist
	if tempChannel.OwnerID != unblockerID {
		return fmt.Errorf("nur der Owner kann User entblockieren")
	}

	// Parse blocked users
	var blockedUsers []string
	err = json.Unmarshal([]byte(tempChannel.BlockedUsers), &blockedUsers)
	if err != nil {
		return err
	}

	// Entferne User von Blocklist
	newBlockedUsers := []string{}
	found := false
	for _, blocked := range blockedUsers {
		if blocked != userID {
			newBlockedUsers = append(newBlockedUsers, blocked)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("user ist nicht blockiert")
	}

	blockedJSON, err := json.Marshal(newBlockedUsers)
	if err != nil {
		return err
	}

	err = tables.UpdateTemporaryVoiceChannelBlockedUsers(db, channelID, string(blockedJSON))
	if err != nil {
		return err
	}

	// Entferne Channel Permission Override
	err = s.ChannelPermissionDelete(channelID, userID)
	if err != nil {
		// Ignoriere Fehler wenn keine Permission existiert
	}

	logger := logging.NewLogger(db, s, guildID, "voice.create_voice")
	logger.LogInfo("User entblockiert", fmt.Sprintf("Channel: %s, User: %s, Unblocker: %s", channelID, userID, unblockerID), false)

	return nil
}

// UpdateChannelName ändert den Namen eines temporären Channels
func UpdateChannelName(s *discordgo.Session, db *sql.DB, channelID, newName, userID string) error {
	tempChannel, err := tables.GetTemporaryVoiceChannelByChannelID(db, channelID)
	if err != nil {
		return err
	}

	if tempChannel == nil {
		return fmt.Errorf("kein temporärer Channel")
	}

	// Prüfe ob der User der Owner ist
	if tempChannel.OwnerID != userID {
		return fmt.Errorf("nur der Owner kann den Namen ändern")
	}

	// Ändere den Channel-Namen
	_, err = s.ChannelEdit(channelID, &discordgo.ChannelEdit{
		Name: newName,
	})

	if err != nil {
		return err
	}

	logger := logging.NewLogger(db, s, tempChannel.GuildID, "voice.create_voice")
	logger.LogInfo("Channel umbenannt", fmt.Sprintf("Channel: %s, Neuer Name: %s, Owner: %s", channelID, newName, userID), false)

	return nil
}

// UpdateChannelLimit ändert das User-Limit eines temporären Channels
func UpdateChannelLimit(s *discordgo.Session, db *sql.DB, channelID string, newLimit int, userID string) error {
	tempChannel, err := tables.GetTemporaryVoiceChannelByChannelID(db, channelID)
	if err != nil {
		return err
	}

	if tempChannel == nil {
		return fmt.Errorf("kein temporärer Channel")
	}

	// Prüfe ob der User der Owner ist
	if tempChannel.OwnerID != userID {
		return fmt.Errorf("nur der Owner kann das Limit ändern")
	}

	// Ändere das User-Limit
	_, err = s.ChannelEdit(channelID, &discordgo.ChannelEdit{
		UserLimit: newLimit,
	})

	if err != nil {
		return err
	}

	logger := logging.NewLogger(db, s, tempChannel.GuildID, "voice.create_voice")
	logger.LogInfo("Channel Limit geändert", fmt.Sprintf("Channel: %s, Neues Limit: %d, Owner: %s", channelID, newLimit, userID), false)

	return nil
}

// GetBlockedUsers gibt die Liste der blockierten User zurück
func GetBlockedUsers(tempChannel *tables.TemporaryVoiceChannel, blockedUsers *[]string) error {
	return json.Unmarshal([]byte(tempChannel.BlockedUsers), blockedUsers)
}
