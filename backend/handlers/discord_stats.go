package handlers

import (
	"database/sql"
	"discord-bot-template/shared/database/tables"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// DiscordStatsHandler verwaltet Discord-Statistiken
type DiscordStatsHandler struct {
	db            *sql.DB
	botAPIBaseURL string
	guildID       string
	roleID        string // Die zu überwachende Rolle
}

// NewDiscordStatsHandler erstellt einen neuen Discord Stats Handler
func NewDiscordStatsHandler(db *sql.DB) *DiscordStatsHandler {
	botAPIBaseURL := os.Getenv("BOT_API_URL")
	if botAPIBaseURL == "" {
		botAPIBaseURL = "http://localhost:8090" // Default fallback
	}

	guildID := os.Getenv("GUILD_ID")
	roleID := os.Getenv("DISCORD_TRACKED_ROLE_ID") // Optional: Rolle die getrackt werden soll

	return &DiscordStatsHandler{
		db:            db,
		botAPIBaseURL: botAPIBaseURL,
		guildID:       guildID,
		roleID:        roleID,
	}
}

// BotAPIResponse Strukturen für Bot API Responses
type MemberCountResponse struct {
	GuildID          string `json:"guild_id"`
	MemberCount      int    `json:"member_count"`
	CachedMembers    int    `json:"cached_members"`
	ApproximateCount int    `json:"approximate_count"`
}

type RoleMemberCountResponse struct {
	GuildID       string `json:"guild_id"`
	RoleID        string `json:"role_id"`
	RoleExists    bool   `json:"role_exists"`
	RoleName      string `json:"role_name"`
	MemberCount   int    `json:"member_count"`
	CachedMembers int    `json:"cached_members"`
	TotalMembers  int    `json:"total_members"`
}

type ChannelCountResponse struct {
	GuildID          string `json:"guild_id"`
	TotalChannels    int    `json:"total_channels"`
	TextChannels     int    `json:"text_channels"`
	VoiceChannels    int    `json:"voice_channels"`
	CategoryChannels int    `json:"category_channels"`
	OtherChannels    int    `json:"other_channels"`
}

type VoiceUserCountResponse struct {
	GuildID             string                            `json:"guild_id"`
	TotalVoiceUsers     int                               `json:"total_voice_users"`
	ActiveVoiceChannels int                               `json:"active_voice_channels"`
	ChannelDetails      map[string]map[string]interface{} `json:"channel_details"`
}

// FetchAndSaveStatistics holt alle Statistiken vom Bot und speichert sie
func (h *DiscordStatsHandler) FetchAndSaveStatistics(source string) (*tables.DiscordStatistic, error) {
	if h.guildID == "" {
		return nil, fmt.Errorf("GUILD_ID not configured")
	}

	log.Printf("[Discord Stats] Starting fetch for guild %s from %s", h.guildID, h.botAPIBaseURL)

	// Hole alle Statistiken parallel
	memberCountChan := make(chan *MemberCountResponse)
	roleMemberCountChan := make(chan *RoleMemberCountResponse)
	channelCountChan := make(chan *ChannelCountResponse)
	voiceUserCountChan := make(chan *VoiceUserCountResponse)
	errorChan := make(chan error, 4)

	// Member Count
	go func() {
		url := fmt.Sprintf("%s/api/guild/member-count?guild_id=%s", h.botAPIBaseURL, h.guildID)
		log.Printf("[Discord Stats] Fetching member count from: %s", url)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("[Discord Stats] Error fetching member count: %v", err)
			errorChan <- err
			memberCountChan <- nil
			return
		}
		defer resp.Body.Close()

		var data MemberCountResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			log.Printf("[Discord Stats] Error decoding member count: %v", err)
			errorChan <- err
			memberCountChan <- nil
			return
		}
		log.Printf("[Discord Stats] Member count received: %d", data.MemberCount)
		memberCountChan <- &data
	}()

	// Role Member Count (wenn roleID konfiguriert)
	go func() {
		if h.roleID == "" {
			roleMemberCountChan <- nil
			return
		}

		resp, err := http.Get(fmt.Sprintf("%s/api/guild/role-member-count?guild_id=%s&role_id=%s",
			h.botAPIBaseURL, h.guildID, h.roleID))
		if err != nil {
			errorChan <- err
			roleMemberCountChan <- nil
			return
		}
		defer resp.Body.Close()

		var data RoleMemberCountResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			errorChan <- err
			roleMemberCountChan <- nil
			return
		}
		roleMemberCountChan <- &data
	}()

	// Channel Count
	go func() {
		resp, err := http.Get(fmt.Sprintf("%s/api/guild/channel-count?guild_id=%s", h.botAPIBaseURL, h.guildID))
		if err != nil {
			errorChan <- err
			channelCountChan <- nil
			return
		}
		defer resp.Body.Close()

		var data ChannelCountResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			errorChan <- err
			channelCountChan <- nil
			return
		}
		channelCountChan <- &data
	}()

	// Voice User Count
	go func() {
		resp, err := http.Get(fmt.Sprintf("%s/api/guild/voice-user-count?guild_id=%s", h.botAPIBaseURL, h.guildID))
		if err != nil {
			errorChan <- err
			voiceUserCountChan <- nil
			return
		}
		defer resp.Body.Close()

		var data VoiceUserCountResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			errorChan <- err
			voiceUserCountChan <- nil
			return
		}
		voiceUserCountChan <- &data
	}()

	// Warte auf alle Responses
	memberCount := <-memberCountChan
	roleMemberCount := <-roleMemberCountChan
	channelCount := <-channelCountChan
	voiceUserCount := <-voiceUserCountChan

	// Prüfe auf Fehler
	select {
	case err := <-errorChan:
		log.Printf("Error fetching statistics: %v", err)
		// Fahre trotzdem fort mit den Daten die wir haben
	default:
	}

	// Erstelle Statistik-Objekt
	stat := &tables.DiscordStatistic{
		GuildID:   h.guildID,
		Timestamp: time.Now(),
		Source:    source,
	}

	if memberCount != nil {
		stat.MemberCount = memberCount.MemberCount
	}

	if roleMemberCount != nil {
		stat.RoleMemberCount = roleMemberCount.MemberCount
		stat.RoleID = h.roleID
	}

	if channelCount != nil {
		stat.TotalChannels = channelCount.TotalChannels
		stat.TextChannels = channelCount.TextChannels
		stat.VoiceChannels = channelCount.VoiceChannels
		stat.CategoryChannels = channelCount.CategoryChannels
	}

	if voiceUserCount != nil {
		stat.VoiceUserCount = voiceUserCount.TotalVoiceUsers
		stat.ActiveVoiceChannels = voiceUserCount.ActiveVoiceChannels
	}

	// Berechne Total Voice Time bis jetzt
	var totalVoiceTime sql.NullInt64
	err := h.db.QueryRow(`
		SELECT CAST(COALESCE(SUM(EXTRACT(EPOCH FROM (COALESCE(left_at, NOW()) - joined_at))), 0) AS BIGINT)
		FROM user_voice_logs
	`).Scan(&totalVoiceTime)
	if err != nil {
		log.Printf("Error getting total voice time: %v", err)
		stat.TotalVoiceTime = 0
	} else if totalVoiceTime.Valid {
		stat.TotalVoiceTime = int64(totalVoiceTime.Int64)
	} else {
		stat.TotalVoiceTime = 0
	}

	// Speichere in Datenbank
	if err := tables.SaveDiscordStatistic(h.db, stat); err != nil {
		return nil, fmt.Errorf("failed to save statistics: %v", err)
	}

	log.Printf("Discord statistics saved: source=%s, members=%d, channels=%d, voice_users=%d, voice_time=%d",
		source, stat.MemberCount, stat.TotalChannels, stat.VoiceUserCount, stat.TotalVoiceTime)

	return stat, nil
}

// GetCurrentStats HTTP Handler - Holt aktuelle Statistiken und speichert sie
func GetCurrentStats(db *sql.DB) http.HandlerFunc {
	handler := NewDiscordStatsHandler(db)
	return func(w http.ResponseWriter, r *http.Request) {
		stat, err := handler.FetchAndSaveStatistics("manual")
		if err != nil {
			log.Printf("Error fetching statistics: %v", err)
			http.Error(w, fmt.Sprintf("Failed to fetch statistics: %v", err), http.StatusInternalServerError)
			return
		}

		// Hole zusätzliche Statistiken aus der Datenbank
		additionalStats := getAdditionalStats(db)

		// Kombiniere die Stats
		response := map[string]interface{}{
			"current_stats":      stat,
			"user_max":           additionalStats["user_max"],
			"total_messages":     additionalStats["total_messages"],
			"total_voice_time":   additionalStats["total_voice_time"],
			"avg_voice_time_day": additionalStats["avg_voice_time_day"],
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// getAdditionalStats holt zusätzliche Statistiken aus der Datenbank
func getAdditionalStats(db *sql.DB) map[string]interface{} {
	stats := make(map[string]interface{})

	// User Max (höchste Anzahl an Usern in der users Tabelle zu einem Zeitpunkt)
	var userMax int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM users WHERE joined_at IS NOT NULL
	`).Scan(&userMax)
	if err != nil {
		log.Printf("Error getting user max: %v", err)
		userMax = 0
	}
	stats["user_max"] = userMax

	// Total Messages
	var totalMessages int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM user_messages_logs
	`).Scan(&totalMessages)
	if err != nil {
		log.Printf("Error getting total messages: %v", err)
		totalMessages = 0
	}
	stats["total_messages"] = totalMessages

	// Total Voice Time (in Sekunden)
	var totalVoiceTime sql.NullInt64
	err = db.QueryRow(`
		SELECT CAST(COALESCE(SUM(EXTRACT(EPOCH FROM (left_at - joined_at))), 0) AS BIGINT)
		FROM user_voice_logs 
		WHERE left_at IS NOT NULL
	`).Scan(&totalVoiceTime)
	if err != nil {
		log.Printf("Error getting total voice time: %v", err)
		totalVoiceTime.Int64 = 0
	}
	if !totalVoiceTime.Valid {
		totalVoiceTime.Int64 = 0
	}
	stats["total_voice_time"] = totalVoiceTime.Int64

	// Durchschnittliche Voice Time pro Tag (letzte 30 Tage)
	var avgVoiceTimeDay sql.NullFloat64
	err = db.QueryRow(`
		SELECT AVG(daily_time) FROM (
			SELECT 
				DATE(joined_at) as day,
				SUM(EXTRACT(EPOCH FROM (COALESCE(left_at, NOW()) - joined_at))) as daily_time
			FROM user_voice_logs
			WHERE joined_at >= NOW() - INTERVAL '30 days'
			GROUP BY DATE(joined_at)
		) as daily_stats
	`).Scan(&avgVoiceTimeDay)
	if err != nil {
		log.Printf("Error getting avg voice time: %v", err)
		avgVoiceTimeDay.Float64 = 0
	}
	if !avgVoiceTimeDay.Valid {
		avgVoiceTimeDay.Float64 = 0
	}
	stats["avg_voice_time_day"] = int64(avgVoiceTimeDay.Float64)

	return stats
}

// GetHistoricalStats HTTP Handler - Gibt historische Statistiken zurück
func GetHistoricalStats(db *sql.DB) http.HandlerFunc {
	handler := NewDiscordStatsHandler(db)
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse Query Parameter
		queryParams := r.URL.Query()
		limitStr := queryParams.Get("limit")
		sinceStr := queryParams.Get("since") // Unix timestamp oder RFC3339

		var limit int
		var since *time.Time

		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
		}

		if sinceStr != "" {
			// Versuche Unix timestamp
			if unixTime, err := strconv.ParseInt(sinceStr, 10, 64); err == nil {
				t := time.Unix(unixTime, 0)
				since = &t
			} else {
				// Versuche RFC3339
				if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
					since = &t
				} else {
					http.Error(w, "Invalid since parameter", http.StatusBadRequest)
					return
				}
			}
		}

		statistics, err := tables.GetDiscordStatistics(db, handler.guildID, since, limit)
		if err != nil {
			log.Printf("Error getting historical statistics: %v", err)
			http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statistics)
	}
}

// GetStatisticsInRange HTTP Handler - Gibt Statistiken in einem Zeitbereich zurück
func GetStatisticsInRange(db *sql.DB) http.HandlerFunc {
	handler := NewDiscordStatsHandler(db)
	return func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		startStr := queryParams.Get("start")
		endStr := queryParams.Get("end")

		if startStr == "" || endStr == "" {
			http.Error(w, "start and end parameters are required", http.StatusBadRequest)
			return
		}

		var startTime, endTime time.Time
		var err error

		// Parse start time
		if unixTime, e := strconv.ParseInt(startStr, 10, 64); e == nil {
			startTime = time.Unix(unixTime, 0)
		} else if t, e := time.Parse(time.RFC3339, startStr); e == nil {
			startTime = t
		} else {
			http.Error(w, "Invalid start parameter", http.StatusBadRequest)
			return
		}

		// Parse end time
		if unixTime, e := strconv.ParseInt(endStr, 10, 64); e == nil {
			endTime = time.Unix(unixTime, 0)
		} else if t, e := time.Parse(time.RFC3339, endStr); e == nil {
			endTime = t
		} else {
			http.Error(w, "Invalid end parameter", http.StatusBadRequest)
			return
		}

		statistics, err := tables.GetStatisticsInTimeRange(db, handler.guildID, startTime, endTime)
		if err != nil {
			log.Printf("Error getting statistics in range: %v", err)
			http.Error(w, "Failed to get statistics", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statistics)
	}
}

// ProxyBotAPI erstellt einen Proxy-Handler für direkte Bot API Calls
func (h *DiscordStatsHandler) ProxyBotAPI(endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Forwarde Request an Bot API
		url := fmt.Sprintf("%s%s?%s", h.botAPIBaseURL, endpoint, r.URL.RawQuery)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error proxying to bot API: %v", err)
			http.Error(w, "Failed to contact bot API", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Kopiere Response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}
