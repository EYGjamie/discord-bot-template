package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"discord-bot-template/bot/utils/notifications"

	"github.com/bwmarrin/discordgo"
)

// EventNotificationRequest represents an event notification request
type EventNotificationRequest struct {
	Type          string `json:"type"` // "event_invitation", "event_update", "event_cancellation", "event_reminder"
	EventID       int64  `json:"event_id"`
	GuestUserID   string `json:"guest_user_id,omitempty"`
	UpdateMessage string `json:"update_message,omitempty"`
	ReminderTime  string `json:"reminder_time,omitempty"` // e.g., "in 1 Stunde", "morgen"
}

// EventNotificationHandler handles incoming event notification requests
func EventNotificationHandler(session *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req EventNotificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Create notification service
		notificationService := notifications.NewEventNotificationService(session, db)

		// Handle notification based on type
		var err error
		switch req.Type {
		case "event_invitation":
			if req.GuestUserID == "" {
				http.Error(w, "Missing guest_user_id", http.StatusBadRequest)
				return
			}
			err = notificationService.NotifyEventInvitation(req.EventID, req.GuestUserID)

		case "event_update":
			if req.UpdateMessage == "" {
				req.UpdateMessage = "Das Event wurde aktualisiert."
			}
			err = notificationService.NotifyEventUpdate(req.EventID, req.UpdateMessage)

		case "event_cancellation":
			err = notificationService.NotifyEventCancellation(req.EventID)

		case "event_reminder":
			if req.ReminderTime == "" {
				req.ReminderTime = "bald"
			}
			err = notificationService.NotifyEventReminder(req.EventID, req.ReminderTime)

		default:
			http.Error(w, "Invalid notification type", http.StatusBadRequest)
			return
		}

		if err != nil {
			log.Printf("Error sending event notification: %v", err)
			http.Error(w, "Failed to send notification", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
	}
}
