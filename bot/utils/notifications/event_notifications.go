package notifications

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"discord-bot-template/shared/database/tables"

	"github.com/bwmarrin/discordgo"
)

// EventNotificationService handles Discord DM notifications for event invitations
type EventNotificationService struct {
	session *discordgo.Session
	db      *sql.DB
}

// NewEventNotificationService creates a new event notification service
func NewEventNotificationService(session *discordgo.Session, db *sql.DB) *EventNotificationService {
	return &EventNotificationService{
		session: session,
		db:      db,
	}
}

// RegisterEventHandlers registers interaction handlers for event RSVP buttons
func (s *EventNotificationService) RegisterEventHandlers() {
	s.session.AddHandler(func(sess *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionMessageComponent {
			return
		}

		customID := i.MessageComponentData().CustomID

		// Check if it's an event RSVP button
		if len(customID) < 12 {
			return
		}

		var eventID int64
		var action string

		if _, err := fmt.Sscanf(customID, "event_accept_%d", &eventID); err == nil {
			action = "accepted"
		} else if _, err := fmt.Sscanf(customID, "event_decline_%d", &eventID); err == nil {
			action = "declined"
		} else {
			return // Not an event RSVP button
		}

		userID := i.User.ID
		if i.Member != nil {
			userID = i.Member.User.ID
		}

		// Update RSVP status in database
		err := s.UpdateRSVPByUserID(eventID, userID, action)
		if err != nil {
			log.Printf("Failed to update RSVP: %v", err)
			sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Fehler beim Aktualisieren deines RSVP-Status.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}

		// Get event details for confirmation
		var eventTitle string
		err = s.db.QueryRow("SELECT title FROM events WHERE id = $1", eventID).Scan(&eventTitle)
		if err != nil {
			eventTitle = "Event"
		}

		responseText := ""
		if action == "accepted" {
			responseText = fmt.Sprintf("✅ Du hast dem Event **%s** zugesagt!", eventTitle)
		} else {
			responseText = fmt.Sprintf("❌ Du hast dem Event **%s** abgesagt.", eventTitle)
		}

		// Update the original message to show the response
		sess.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    responseText,
				Components: []discordgo.MessageComponent{}, // Remove buttons
			},
		})
	})
}

// UpdateRSVPByUserID updates a guest's RSVP status by their user ID
func (s *EventNotificationService) UpdateRSVPByUserID(eventID int64, userID string, status string) error {
	_, err := s.db.Exec(`
		UPDATE event_guests 
		SET rsvp_status = $1, rsvp_at = $2 
		WHERE event_id = $3 AND user_id = $4
	`, status, time.Now(), eventID, userID)
	return err
}

// NotifyEventInvitation sends a DM to a user when they're invited to an event
func (s *EventNotificationService) NotifyEventInvitation(eventID int64, guestUserID string) error {
	// Get event details
	var event tables.Event
	var tagsJSON string
	err := s.db.QueryRow(`
		SELECT id, guild_id, title, description, start_date, end_date, start_time, end_time, 
		       is_all_day, color, location, COALESCE(tags, '[]'), created_by
		FROM events WHERE id = $1
	`, eventID).Scan(
		&event.ID, &event.GuildID, &event.Title, &event.Description,
		&event.StartDate, &event.EndDate, &event.StartTime, &event.EndTime,
		&event.IsAllDay, &event.Color, &event.Location, &tagsJSON, &event.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to get event: %v", err)
	}

	// Get inviter info (name and avatar)
	inviterName := "Jemand"
	inviterAvatarURL := ""
	if member, err := s.session.GuildMember(event.GuildID, event.CreatedBy); err == nil {
		// Prefer DisplayName > Nick > GlobalName > Username
		if member.User.GlobalName != "" {
			inviterName = member.User.GlobalName
		} else if member.Nick != "" {
			inviterName = member.Nick
		} else {
			inviterName = member.User.Username
		}
		// Get avatar URL
		if member.Avatar != "" {
			inviterAvatarURL = fmt.Sprintf("https://cdn.discordapp.com/guilds/%s/users/%s/avatars/%s.png", event.GuildID, event.CreatedBy, member.Avatar)
		} else if member.User.Avatar != "" {
			inviterAvatarURL = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", event.CreatedBy, member.User.Avatar)
		}
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📅 Event-Einladung: %s", event.Title),
		Description: fmt.Sprintf("**%s** hat dich zu einem Event eingeladen!", inviterName),
		Color:       0x5865F2, // Discord Blurple
		Fields:      []*discordgo.MessageEmbedField{},
		Timestamp:   time.Now().Format(time.RFC3339),
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Bitte antworte mit Zusage oder Absage",
		},
	}

	// Add inviter avatar as thumbnail
	if inviterAvatarURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: inviterAvatarURL,
		}
	}

	// Add event details
	if event.Description != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📝 Beschreibung",
			Value:  truncateString(event.Description, 200),
			Inline: false,
		})
	}

	// Format date/time nicely
	dateStr := formatDateNice(event.StartDate)
	if event.EndDate != event.StartDate {
		dateStr = fmt.Sprintf("%s - %s", formatDateNice(event.StartDate), formatDateNice(event.EndDate))
	}
	if !event.IsAllDay && event.StartTime != "" {
		timeStr := formatTimeNice(event.StartTime)
		if event.EndTime != "" && event.EndTime != event.StartTime {
			timeStr = fmt.Sprintf("%s - %s", formatTimeNice(event.StartTime), formatTimeNice(event.EndTime))
		}
		dateStr = fmt.Sprintf("%s, %s Uhr", dateStr, timeStr)
	} else if event.IsAllDay {
		dateStr = fmt.Sprintf("%s (Ganztägig)", dateStr)
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "📆 Datum & Zeit",
		Value:  dateStr,
		Inline: true,
	})

	if event.Location != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📍 Ort",
			Value:  event.Location,
			Inline: true,
		})
	}

	// Create RSVP buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "✅ Zusagen",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("event_accept_%d", eventID),
				},
				discordgo.Button{
					Label:    "❌ Absagen",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("event_decline_%d", eventID),
				},
			},
		},
	}

	// Create DM channel
	channel, err := s.session.UserChannelCreate(guestUserID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel for user %s: %v", guestUserID, err)
	}

	// Send message with embed and buttons
	_, err = s.session.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embed:      embed,
		Components: components,
	})
	if err != nil {
		return fmt.Errorf("failed to send DM to user %s: %v", guestUserID, err)
	}

	log.Printf("Sent event invitation to user %s for event %d", guestUserID, eventID)
	return nil
}

// NotifyEventUpdate sends a DM to all guests when an event is updated
func (s *EventNotificationService) NotifyEventUpdate(eventID int64, updateMessage string) error {
	// Get all guests for this event
	rows, err := s.db.Query(`
		SELECT user_id FROM event_guests WHERE event_id = $1
	`, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event guests: %v", err)
	}
	defer rows.Close()

	// Get event details
	var eventTitle string
	var eventDate string
	err = s.db.QueryRow(`
		SELECT title, start_date FROM events WHERE id = $1
	`, eventID).Scan(&eventTitle, &eventDate)
	if err != nil {
		return fmt.Errorf("failed to get event: %v", err)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📅 Event aktualisiert: %s", eventTitle),
		Description: updateMessage,
		Color:       0xFFA500, // Orange
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📆 Datum",
				Value:  eventDate,
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		s.sendDM(userID, embed)
	}

	return nil
}

// NotifyEventCancellation sends a DM to all guests when an event is cancelled
func (s *EventNotificationService) NotifyEventCancellation(eventID int64) error {
	// Get event details before it's deleted
	var eventTitle, eventDate, eventTime, location string
	var isAllDay bool
	err := s.db.QueryRow(`
		SELECT title, start_date, COALESCE(start_time, ''), is_all_day, COALESCE(location, '')
		FROM events WHERE id = $1
	`, eventID).Scan(&eventTitle, &eventDate, &eventTime, &isAllDay, &location)
	if err != nil {
		return fmt.Errorf("failed to get event: %v", err)
	}

	// Get all guests
	rows, err := s.db.Query(`
		SELECT user_id FROM event_guests WHERE event_id = $1
	`, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event guests: %v", err)
	}
	defer rows.Close()

	dateStr := eventDate
	if !isAllDay && eventTime != "" {
		dateStr = fmt.Sprintf("%s um %s", eventDate, eventTime)
	}

	embed := &discordgo.MessageEmbed{
		Title:       "❌ Event abgesagt",
		Description: fmt.Sprintf("Das Event **%s** wurde abgesagt.", eventTitle),
		Color:       0xDC143C, // Crimson
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📆 Geplantes Datum",
				Value:  dateStr,
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if location != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📍 Geplanter Ort",
			Value:  location,
			Inline: true,
		})
	}

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		s.sendDM(userID, embed)
	}

	return nil
}

// NotifyEventReminder sends a reminder to all accepted guests
func (s *EventNotificationService) NotifyEventReminder(eventID int64, reminderTime string) error {
	// Get event details
	var event tables.Event
	err := s.db.QueryRow(`
		SELECT id, title, start_date, start_time, end_time, is_all_day, location
		FROM events WHERE id = $1
	`, eventID).Scan(
		&event.ID, &event.Title, &event.StartDate, &event.StartTime, &event.EndTime, &event.IsAllDay, &event.Location,
	)
	if err != nil {
		return fmt.Errorf("failed to get event: %v", err)
	}

	// Get accepted guests
	rows, err := s.db.Query(`
		SELECT user_id FROM event_guests WHERE event_id = $1 AND rsvp_status = 'accepted'
	`, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event guests: %v", err)
	}
	defer rows.Close()

	dateStr := event.StartDate
	if !event.IsAllDay && event.StartTime != "" {
		dateStr = fmt.Sprintf("%s um %s", event.StartDate, event.StartTime)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("⏰ Erinnerung: %s", event.Title),
		Description: fmt.Sprintf("Das Event beginnt %s!", reminderTime),
		Color:       0x5865F2, // Discord Blurple
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "📆 Datum & Zeit",
				Value:  dateStr,
				Inline: true,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if event.Location != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "📍 Ort",
			Value:  event.Location,
			Inline: true,
		})
	}

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		s.sendDM(userID, embed)
	}

	return nil
}

// sendDM sends a direct message to a user
func (s *EventNotificationService) sendDM(userID string, embed *discordgo.MessageEmbed) error {
	channel, err := s.session.UserChannelCreate(userID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel for user %s: %v", userID, err)
	}

	_, err = s.session.ChannelMessageSendEmbed(channel.ID, embed)
	if err != nil {
		return fmt.Errorf("failed to send DM to user %s: %v", userID, err)
	}

	log.Printf("Sent event notification to user %s", userID)
	return nil
}

// formatDateNice formats a date string (YYYY-MM-DD or ISO) to a nice German format
func formatDateNice(dateStr string) string {
	// Try parsing different date formats
	var t time.Time
	var err error

	// Try ISO format first (2026-01-31T00:00:00Z)
	t, err = time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// Try simple date format (2026-01-31)
		t, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return dateStr // Return as-is if parsing fails
		}
	}

	// German weekday names
	weekdays := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
	// German month names
	months := []string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"}

	return fmt.Sprintf("%s, %d. %s %d", weekdays[t.Weekday()], t.Day(), months[t.Month()-1], t.Year())
}

// formatTimeNice formats a time string (HH:MM:SS or HH:MM) to HH:MM
func formatTimeNice(timeStr string) string {
	// Remove seconds if present
	if len(timeStr) >= 5 {
		return timeStr[:5]
	}
	return timeStr
}
