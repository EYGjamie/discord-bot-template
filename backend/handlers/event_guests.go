package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type EventGuestResponse struct {
	ID              int64   `json:"id"`
	EventID         int64   `json:"event_id"`
	UserID          string  `json:"user_id"`
	UserName        string  `json:"user_name"`
	UserDisplayName string  `json:"user_display_name"`
	UserAvatar      string  `json:"user_avatar"`
	RSVPStatus      string  `json:"rsvp_status"`
	RSVPAt          *string `json:"rsvp_at"`
	InvitedAt       string  `json:"invited_at"`
}

type InviteGuestRequest struct {
	UserID string `json:"user_id"`
}

type RSVPRequest struct {
	Status string `json:"status"` // accepted, declined
}

// GetEventGuests holt alle Gäste eines Events
func GetEventGuests(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")

		query := `
			SELECT eg.id, eg.event_id, eg.user_id, 
			       COALESCE(u.name, eg.user_name) as user_name,
			       COALESCE(u.display_name, eg.user_display_name, u.name, eg.user_name) as user_display_name,
			       COALESCE(u.avatar, eg.user_avatar) as user_avatar,
			       eg.rsvp_status, eg.rsvp_at, eg.invited_at
			FROM event_guests eg
			LEFT JOIN users u ON eg.user_id = u.id
			WHERE eg.event_id = $1
			ORDER BY eg.invited_at ASC
		`

		rows, err := db.Query(query, eventID)
		if err != nil {
			log.Printf("Error fetching event guests: %v", err)
			http.Error(w, "Failed to fetch guests", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		guests := []EventGuestResponse{}
		for rows.Next() {
			var guest EventGuestResponse
			var rsvpAt sql.NullTime
			var userAvatar sql.NullString

			err := rows.Scan(
				&guest.ID, &guest.EventID, &guest.UserID,
				&guest.UserName, &guest.UserDisplayName, &userAvatar,
				&guest.RSVPStatus, &rsvpAt, &guest.InvitedAt,
			)
			if err != nil {
				log.Printf("Error scanning guest: %v", err)
				continue
			}

			if userAvatar.Valid {
				guest.UserAvatar = userAvatar.String
			}
			if rsvpAt.Valid {
				rsvpAtStr := rsvpAt.Time.Format(time.RFC3339)
				guest.RSVPAt = &rsvpAtStr
			}

			guests = append(guests, guest)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(guests)
	}
}

// InviteEventGuest lädt einen Gast zu einem Event ein
func InviteEventGuest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")

		var req InviteGuestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.UserID == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}

		// Hole User-Informationen aus der users-Tabelle
		var userName, userDisplayName, userAvatar string
		userQuery := `SELECT name, COALESCE(display_name, name), COALESCE(avatar, '') FROM users WHERE id = $1`
		err := db.QueryRow(userQuery, req.UserID).Scan(&userName, &userDisplayName, &userAvatar)
		if err != nil {
			log.Printf("Error fetching user info: %v", err)
			// Fallback: User nicht in DB, verwende leere Werte
			userName = "Unknown"
			userDisplayName = "Unknown"
			userAvatar = ""
		}

		// Füge Gast hinzu
		insertQuery := `
			INSERT INTO event_guests (event_id, user_id, user_name, user_display_name, user_avatar, rsvp_status)
			VALUES ($1, $2, $3, $4, $5, 'pending')
			ON CONFLICT (event_id, user_id) DO NOTHING
			RETURNING id, invited_at
		`

		var guestID int64
		var invitedAt time.Time
		err = db.QueryRow(insertQuery, eventID, req.UserID, userName, userDisplayName, userAvatar).Scan(&guestID, &invitedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				// Guest already exists
				http.Error(w, "Guest already invited", http.StatusConflict)
				return
			}
			log.Printf("Error inviting guest: %v", err)
			http.Error(w, "Failed to invite guest", http.StatusInternalServerError)
			return
		}

		// Trigger Discord Bot to send DM to invited user
		go sendEventInvitationNotification(mustParseInt64(eventID), req.UserID)

		guest := EventGuestResponse{
			ID:              guestID,
			EventID:         mustParseInt64(eventID),
			UserID:          req.UserID,
			UserName:        userName,
			UserDisplayName: userDisplayName,
			UserAvatar:      userAvatar,
			RSVPStatus:      "pending",
			InvitedAt:       invitedAt.Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(guest)
	}
}

// RemoveEventGuest entfernt einen Gast von einem Event
func RemoveEventGuest(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")
		guestID := r.PathValue("guestId")

		query := `DELETE FROM event_guests WHERE event_id = $1 AND id = $2`
		result, err := db.Exec(query, eventID, guestID)
		if err != nil {
			log.Printf("Error removing guest: %v", err)
			http.Error(w, "Failed to remove guest", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Guest not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// UpdateGuestRSVP aktualisiert den RSVP-Status eines Gastes
func UpdateGuestRSVP(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")
		guestID := r.PathValue("guestId")

		var req RSVPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Status != "accepted" && req.Status != "declined" && req.Status != "pending" {
			http.Error(w, "Invalid status. Must be 'accepted', 'declined', or 'pending'", http.StatusBadRequest)
			return
		}

		query := `
			UPDATE event_guests 
			SET rsvp_status = $1, rsvp_at = CURRENT_TIMESTAMP
			WHERE event_id = $2 AND id = $3
			RETURNING id, event_id, user_id, user_name, user_display_name, user_avatar, rsvp_status, rsvp_at, invited_at
		`

		var guest EventGuestResponse
		var rsvpAt sql.NullTime
		var userAvatar sql.NullString

		err := db.QueryRow(query, req.Status, eventID, guestID).Scan(
			&guest.ID, &guest.EventID, &guest.UserID,
			&guest.UserName, &guest.UserDisplayName, &userAvatar,
			&guest.RSVPStatus, &rsvpAt, &guest.InvitedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Guest not found", http.StatusNotFound)
				return
			}
			log.Printf("Error updating RSVP: %v", err)
			http.Error(w, "Failed to update RSVP", http.StatusInternalServerError)
			return
		}

		if userAvatar.Valid {
			guest.UserAvatar = userAvatar.String
		}
		if rsvpAt.Valid {
			rsvpAtStr := rsvpAt.Time.Format(time.RFC3339)
			guest.RSVPAt = &rsvpAtStr
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(guest)
	}
}

// RSVPByUserID ermöglicht einem User, per User-ID zu antworten (für Discord Bot Integration)
func RSVPByUserID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")
		userID := r.PathValue("userId")

		var req RSVPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Status != "accepted" && req.Status != "declined" {
			http.Error(w, "Invalid status. Must be 'accepted' or 'declined'", http.StatusBadRequest)
			return
		}

		query := `
			UPDATE event_guests 
			SET rsvp_status = $1, rsvp_at = CURRENT_TIMESTAMP
			WHERE event_id = $2 AND user_id = $3
			RETURNING id
		`

		var guestID int64
		err := db.QueryRow(query, req.Status, eventID, userID).Scan(&guestID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Guest not found or not invited to this event", http.StatusNotFound)
				return
			}
			log.Printf("Error updating RSVP by user ID: %v", err)
			http.Error(w, "Failed to update RSVP", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"status":  req.Status,
		})
	}
}

// GetEventLabels holt alle Event-Labels für die Guild
func GetEventLabels(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		query := `SELECT id, guild_id, name, color, created_at FROM event_labels WHERE guild_id = $1 ORDER BY name`
		rows, err := db.Query(query, guildID)
		if err != nil {
			log.Printf("Error fetching event labels: %v", err)
			http.Error(w, "Failed to fetch labels", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type LabelResponse struct {
			ID        int64  `json:"id"`
			GuildID   string `json:"guild_id"`
			Name      string `json:"name"`
			Color     string `json:"color"`
			CreatedAt string `json:"created_at"`
		}

		labels := []LabelResponse{}
		for rows.Next() {
			var label LabelResponse
			if err := rows.Scan(&label.ID, &label.GuildID, &label.Name, &label.Color, &label.CreatedAt); err != nil {
				log.Printf("Error scanning label: %v", err)
				continue
			}
			labels = append(labels, label)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(labels)
	}
}

// CreateEventLabel erstellt ein neues Event-Label
func CreateEventLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		guildID := os.Getenv("GUILD_ID")
		if guildID == "" {
			http.Error(w, "Server configuration error", http.StatusInternalServerError)
			return
		}

		var req struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if req.Color == "" {
			req.Color = "blue"
		}

		query := `INSERT INTO event_labels (guild_id, name, color) VALUES ($1, $2, $3) RETURNING id, created_at`
		var id int64
		var createdAt string
		err := db.QueryRow(query, guildID, req.Name, req.Color).Scan(&id, &createdAt)
		if err != nil {
			log.Printf("Error creating event label: %v", err)
			http.Error(w, "Failed to create label", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         id,
			"guild_id":   guildID,
			"name":       req.Name,
			"color":      req.Color,
			"created_at": createdAt,
		})
	}
}

// UpdateEventLabel aktualisiert ein Event-Label
func UpdateEventLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		labelID := r.PathValue("labelId")

		var req struct {
			Name  string `json:"name"`
			Color string `json:"color"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		query := `UPDATE event_labels SET name = $1, color = $2 WHERE id = $3 RETURNING id, guild_id, name, color, created_at`
		var id int64
		var guildID, name, color, createdAt string
		err := db.QueryRow(query, req.Name, req.Color, labelID).Scan(&id, &guildID, &name, &color, &createdAt)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Label not found", http.StatusNotFound)
				return
			}
			log.Printf("Error updating event label: %v", err)
			http.Error(w, "Failed to update label", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":         id,
			"guild_id":   guildID,
			"name":       name,
			"color":      color,
			"created_at": createdAt,
		})
	}
}

// DeleteEventLabel löscht ein Event-Label
func DeleteEventLabel(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		labelID := r.PathValue("labelId")

		// Hole Label-Name für Cleanup
		var labelName string
		db.QueryRow(`SELECT name FROM event_labels WHERE id = $1`, labelID).Scan(&labelName)

		// Lösche das Label
		result, err := db.Exec(`DELETE FROM event_labels WHERE id = $1`, labelID)
		if err != nil {
			log.Printf("Error deleting event label: %v", err)
			http.Error(w, "Failed to delete label", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Label not found", http.StatusNotFound)
			return
		}

		// TODO: Entferne Label von allen Events die es verwenden

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetEventChecklist holt alle Checklist-Items eines Events
func GetEventChecklist(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")

		query := `SELECT id, event_id, text, is_completed, position, created_at, updated_at 
		          FROM event_checklist_items WHERE event_id = $1 ORDER BY position`
		rows, err := db.Query(query, eventID)
		if err != nil {
			log.Printf("Error fetching checklist: %v", err)
			http.Error(w, "Failed to fetch checklist", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type ChecklistItem struct {
			ID          int64  `json:"id"`
			EventID     int64  `json:"event_id"`
			Text        string `json:"text"`
			IsCompleted bool   `json:"is_completed"`
			Position    int    `json:"position"`
			CreatedAt   string `json:"created_at"`
			UpdatedAt   string `json:"updated_at"`
		}

		items := []ChecklistItem{}
		for rows.Next() {
			var item ChecklistItem
			if err := rows.Scan(&item.ID, &item.EventID, &item.Text, &item.IsCompleted, &item.Position, &item.CreatedAt, &item.UpdatedAt); err != nil {
				log.Printf("Error scanning checklist item: %v", err)
				continue
			}
			items = append(items, item)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	}
}

// CreateChecklistItem erstellt ein neues Checklist-Item
func CreateEventChecklistItem(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")

		var req struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Text == "" {
			http.Error(w, "text is required", http.StatusBadRequest)
			return
		}

		// Get max position
		var maxPos int
		db.QueryRow(`SELECT COALESCE(MAX(position), 0) FROM event_checklist_items WHERE event_id = $1`, eventID).Scan(&maxPos)

		query := `INSERT INTO event_checklist_items (event_id, text, position) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
		var id int64
		var createdAt, updatedAt string
		err := db.QueryRow(query, eventID, req.Text, maxPos+1).Scan(&id, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("Error creating checklist item: %v", err)
			http.Error(w, "Failed to create checklist item", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           id,
			"event_id":     mustParseInt64(eventID),
			"text":         req.Text,
			"is_completed": false,
			"position":     maxPos + 1,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
		})
	}
}

// UpdateChecklistItem aktualisiert ein Checklist-Item
func UpdateEventChecklistItem(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")
		itemID := r.PathValue("itemId")

		var req struct {
			Text        *string `json:"text"`
			IsCompleted *bool   `json:"is_completed"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Build dynamic update query
		query := `UPDATE event_checklist_items SET updated_at = CURRENT_TIMESTAMP`
		args := []interface{}{}
		argNum := 1

		if req.Text != nil {
			query += `, text = $` + strconv.Itoa(argNum)
			args = append(args, *req.Text)
			argNum++
		}
		if req.IsCompleted != nil {
			query += `, is_completed = $` + strconv.Itoa(argNum)
			args = append(args, *req.IsCompleted)
			argNum++
		}

		query += ` WHERE event_id = $` + strconv.Itoa(argNum) + ` AND id = $` + strconv.Itoa(argNum+1)
		query += ` RETURNING id, event_id, text, is_completed, position, created_at, updated_at`
		args = append(args, eventID, itemID)

		var id, eventIDOut int64
		var text, createdAt, updatedAt string
		var isCompleted bool
		var position int

		err := db.QueryRow(query, args...).Scan(&id, &eventIDOut, &text, &isCompleted, &position, &createdAt, &updatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Checklist item not found", http.StatusNotFound)
				return
			}
			log.Printf("Error updating checklist item: %v", err)
			http.Error(w, "Failed to update checklist item", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           id,
			"event_id":     eventIDOut,
			"text":         text,
			"is_completed": isCompleted,
			"position":     position,
			"created_at":   createdAt,
			"updated_at":   updatedAt,
		})
	}
}

// DeleteChecklistItem löscht ein Checklist-Item
func DeleteEventChecklistItem(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eventID := r.PathValue("id")
		itemID := r.PathValue("itemId")

		result, err := db.Exec(`DELETE FROM event_checklist_items WHERE event_id = $1 AND id = $2`, eventID, itemID)
		if err != nil {
			log.Printf("Error deleting checklist item: %v", err)
			http.Error(w, "Failed to delete checklist item", http.StatusInternalServerError)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			http.Error(w, "Checklist item not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func mustParseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// sendEventInvitationNotification sendet eine Einladung an einen Gast via Bot API
func sendEventInvitationNotification(eventID int64, guestUserID string) {
	botAPIURL := os.Getenv("BOT_API_URL")
	if botAPIURL == "" {
		botAPIURL = "http://bot:8081"
	}

	payload := map[string]interface{}{
		"type":          "event_invitation",
		"event_id":      eventID,
		"guest_user_id": guestUserID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling event notification payload: %v", err)
		return
	}

	resp, err := http.Post(botAPIURL+"/api/notifications/event", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending event invitation notification: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Bot API returned status %d for event invitation", resp.StatusCode)
	} else {
		log.Printf("Successfully sent event invitation to user %s for event %d", guestUserID, eventID)
	}
}
