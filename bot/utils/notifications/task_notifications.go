package notifications

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"discord-bot-template/shared/database/tables"

	"github.com/bwmarrin/discordgo"
)

// TaskNotificationService handles Discord DM notifications for task-related events
type TaskNotificationService struct {
	session *discordgo.Session
	db      *sql.DB
}

// NewTaskNotificationService creates a new task notification service
func NewTaskNotificationService(session *discordgo.Session, db *sql.DB) *TaskNotificationService {
	return &TaskNotificationService{
		session: session,
		db:      db,
	}
}

// shouldSendNotification checks if a user should receive notifications based on their settings
func (s *TaskNotificationService) shouldSendNotification(userID, guildID string, boardID int, notificationType string) (bool, error) {
	// Get global settings
	var globalSettings tables.NotificationSettings
	err := s.db.QueryRow(`
		SELECT task_notifications_enabled, notify_on_assignment, notify_on_task_update, 
		       notify_on_comment, notify_on_due_date_change, notify_on_unassignment, notify_on_checklist_item
		FROM notification_settings
		WHERE user_id = $1 AND guild_id = $2
	`, userID, guildID).Scan(
		&globalSettings.TaskNotificationsEnabled,
		&globalSettings.NotifyOnAssignment,
		&globalSettings.NotifyOnTaskUpdate,
		&globalSettings.NotifyOnComment,
		&globalSettings.NotifyOnDueDateChange,
		&globalSettings.NotifyOnUnassignment,
		&globalSettings.NotifyOnChecklistItem,
	)

	// If no settings exist, default to enabled
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	// Check if task notifications are globally disabled
	if !globalSettings.TaskNotificationsEnabled {
		return false, nil
	}

	// Check notification type
	switch notificationType {
	case "assignment":
		if !globalSettings.NotifyOnAssignment {
			return false, nil
		}
	case "update":
		if !globalSettings.NotifyOnTaskUpdate {
			return false, nil
		}
	case "comment":
		if !globalSettings.NotifyOnComment {
			return false, nil
		}
	case "due_date_change":
		if !globalSettings.NotifyOnDueDateChange {
			return false, nil
		}
	case "unassignment":
		if !globalSettings.NotifyOnUnassignment {
			return false, nil
		}
	case "checklist_item":
		if !globalSettings.NotifyOnChecklistItem {
			return false, nil
		}
	}

	// Check board-specific settings
	var boardSettings tables.BoardNotificationSettings
	err = s.db.QueryRow(`
		SELECT notifications_enabled, notify_on_assignment, notify_on_task_update,
		       notify_on_comment, notify_on_due_date_change, notify_on_unassignment, notify_on_checklist_item
		FROM board_notification_settings
		WHERE user_id = $1 AND board_id = $2
	`, userID, boardID).Scan(
		&boardSettings.NotificationsEnabled,
		&boardSettings.NotifyOnAssignment,
		&boardSettings.NotifyOnTaskUpdate,
		&boardSettings.NotifyOnComment,
		&boardSettings.NotifyOnDueDateChange,
		&boardSettings.NotifyOnUnassignment,
		&boardSettings.NotifyOnChecklistItem,
	)

	// If no board-specific settings exist, use global settings
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	// Check board-specific settings
	if !boardSettings.NotificationsEnabled {
		return false, nil
	}

	switch notificationType {
	case "assignment":
		if !boardSettings.NotifyOnAssignment {
			return false, nil
		}
	case "update":
		if !boardSettings.NotifyOnTaskUpdate {
			return false, nil
		}
	case "comment":
		if !boardSettings.NotifyOnComment {
			return false, nil
		}
	case "due_date_change":
		if !boardSettings.NotifyOnDueDateChange {
			return false, nil
		}
	case "unassignment":
		if !boardSettings.NotifyOnUnassignment {
			return false, nil
		}
	case "checklist_item":
		if !boardSettings.NotifyOnChecklistItem {
			return false, nil
		}
	}

	return true, nil
}

// NotifyTaskAssignment sends a DM to a user when they are assigned to a task
func (s *TaskNotificationService) NotifyTaskAssignment(taskID int, assigneeID, assignedByID string) error {
	guildID := os.Getenv("GUILD_ID")

	// Don't notify if user assigned themselves
	if assigneeID == assignedByID {
		return nil
	}

	// Get task details
	var task tables.Task
	var boardID int
	var boardName string
	err := s.db.QueryRow(`
		SELECT t.id, t.board_id, t.title, t.description, t.status, b.name
		FROM tasks t
		JOIN boards b ON t.board_id = b.id
		WHERE t.id = $1
	`, taskID).Scan(&task.ID, &boardID, &task.Title, &task.Description, &task.Status, &boardName)
	if err != nil {
		return fmt.Errorf("failed to get task details: %v", err)
	}

	// Check if user should receive notification
	shouldNotify, err := s.shouldSendNotification(assigneeID, guildID, boardID, "assignment")
	if err != nil {
		return fmt.Errorf("failed to check notification settings: %v", err)
	}
	if !shouldNotify {
		log.Printf("User %s has disabled assignment notifications", assigneeID)
		return nil
	}

	// Get assigner details
	var assignerName, assignerAvatar string
	err = s.db.QueryRow(`SELECT display_name, avatar_url FROM users WHERE id = $1`, assignedByID).Scan(&assignerName, &assignerAvatar)
	if err != nil {
		assignerName = "Unknown User"
		assignerAvatar = ""
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "📋 Neue Aufgabe zugewiesen",
		Description: fmt.Sprintf("**%s** hat dir eine Aufgabe im Board **%s** zugewiesen.", assignerName, boardName),
		Color:       0x5865F2, // Discord Blurple
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Aufgabe",
				Value:  task.Title,
				Inline: false,
			},
			{
				Name:   "Status",
				Value:  string(task.Status),
				Inline: true,
			},
			{
				Name:   "Board",
				Value:  boardName,
				Inline: true,
			},
		},
	}

	// Add user avatar as thumbnail
	if assignerAvatar != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: assignerAvatar,
		}
	}

	if task.Description != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Beschreibung",
			Value:  truncateString(task.Description, 1024),
			Inline: false,
		})
	}

	// Send DM
	return s.sendDM(assigneeID, embed)
}

// NotifyTaskUpdate sends a DM to the assignee when their task is updated by someone else
func (s *TaskNotificationService) NotifyTaskUpdate(taskID int, updatedByID, changeDescription string) error {
	guildID := os.Getenv("GUILD_ID")

	// Get task details including assignee
	var task tables.Task
	var boardID int
	var boardName string
	var assigneeID *string
	err := s.db.QueryRow(`
		SELECT t.id, t.board_id, t.title, t.assignee_id, t.status, b.name
		FROM tasks t
		JOIN boards b ON t.board_id = b.id
		WHERE t.id = $1
	`, taskID).Scan(&task.ID, &boardID, &task.Title, &assigneeID, &task.Status, &boardName)
	if err != nil {
		return fmt.Errorf("failed to get task details: %v", err)
	}

	// Don't notify if task is not assigned or if the updater is the assignee
	if assigneeID == nil || *assigneeID == updatedByID {
		return nil
	}

	// Check if user should receive notification
	shouldNotify, err := s.shouldSendNotification(*assigneeID, guildID, boardID, "update")
	if err != nil {
		return fmt.Errorf("failed to check notification settings: %v", err)
	}
	if !shouldNotify {
		log.Printf("User %s has disabled update notifications", *assigneeID)
		return nil
	}

	// Get updater details
	var updaterName, updaterAvatar string
	err = s.db.QueryRow(`SELECT display_name, avatar_url FROM users WHERE id = $1`, updatedByID).Scan(&updaterName, &updaterAvatar)
	if err != nil {
		updaterName = "Unknown User"
		updaterAvatar = ""
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "✏️ Aufgabe aktualisiert",
		Description: fmt.Sprintf("**%s** hat deine Aufgabe im Board **%s** bearbeitet.", updaterName, boardName),
		Color:       0xFEE75C, // Yellow
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Aufgabe",
				Value:  task.Title,
				Inline: false,
			},
			{
				Name:   "Status",
				Value:  string(task.Status),
				Inline: true,
			},
			{
				Name:   "Board",
				Value:  boardName,
				Inline: true,
			},
		},
	}

	// Add user avatar as thumbnail
	if updaterAvatar != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: updaterAvatar,
		}
	}

	if changeDescription != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Änderung",
			Value:  truncateString(changeDescription, 1024),
			Inline: false,
		})
	}

	// Send DM
	return s.sendDM(*assigneeID, embed)
}

// sendDM sends a direct message to a user
func (s *TaskNotificationService) sendDM(userID string, embed *discordgo.MessageEmbed) error {
	// Create DM channel
	channel, err := s.session.UserChannelCreate(userID)
	if err != nil {
		return fmt.Errorf("failed to create DM channel for user %s: %v", userID, err)
	}

	// Send message
	_, err = s.session.ChannelMessageSendEmbed(channel.ID, embed)
	if err != nil {
		return fmt.Errorf("failed to send DM to user %s: %v", userID, err)
	}

	log.Printf("Sent task notification to user %s", userID)
	return nil
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// NotifyTaskComment sends a DM to users when a comment is added to a task
func (s *TaskNotificationService) NotifyTaskComment(taskID int, actorID, commentContent string, userIDs []string) error {
	guildID := os.Getenv("GUILD_ID")

	// Get task details
	var task tables.Task
	var boardID int
	var boardName string
	err := s.db.QueryRow(`
		SELECT t.id, t.board_id, t.title, b.name
		FROM tasks t
		JOIN boards b ON t.board_id = b.id
		WHERE t.id = $1
	`, taskID).Scan(&task.ID, &boardID, &task.Title, &boardName)
	if err != nil {
		return fmt.Errorf("failed to get task details: %v", err)
	}

	// Get actor details
	var actorName, actorAvatar string
	err = s.db.QueryRow(`SELECT display_name, avatar_url FROM users WHERE id = $1`, actorID).Scan(&actorName, &actorAvatar)
	if err != nil {
		actorName = "Unknown User"
		actorAvatar = ""
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "💬 Neuer Kommentar",
		Description: fmt.Sprintf("**%s** hat einen Kommentar zu deiner Aufgabe im Board **%s** hinzugefügt.", actorName, boardName),
		Color:       0x57F287, // Green
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Aufgabe",
				Value:  task.Title,
				Inline: false,
			},
			{
				Name:   "Kommentar",
				Value:  truncateString(commentContent, 1024),
				Inline: false,
			},
		},
	}

	// Add user avatar as thumbnail
	if actorAvatar != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: actorAvatar,
		}
	}

	// Send DM to all users
	for _, userID := range userIDs {
		if userID == actorID {
			continue // Don't notify the commenter
		}

		// Check if user should receive notification
		shouldNotify, err := s.shouldSendNotification(userID, guildID, boardID, "comment")
		if err != nil {
			log.Printf("Failed to check notification settings for user %s: %v", userID, err)
			continue
		}
		if !shouldNotify {
			log.Printf("User %s has disabled comment notifications", userID)
			continue
		}

		s.sendDM(userID, embed)
	}

	return nil
}

// NotifyTaskDueDateChange sends a DM to users when a task's due date changes
func (s *TaskNotificationService) NotifyTaskDueDateChange(taskID int, actorID, oldDueDate, newDueDate string, userIDs []string) error {
	guildID := os.Getenv("GUILD_ID")

	// Get task details
	var task tables.Task
	var boardID int
	var boardName string
	err := s.db.QueryRow(`
		SELECT t.id, t.board_id, t.title, b.name
		FROM tasks t
		JOIN boards b ON t.board_id = b.id
		WHERE t.id = $1
	`, taskID).Scan(&task.ID, &boardID, &task.Title, &boardName)
	if err != nil {
		return fmt.Errorf("failed to get task details: %v", err)
	}

	// Get actor details
	var actorName, actorAvatar string
	err = s.db.QueryRow(`SELECT display_name, avatar_url FROM users WHERE id = $1`, actorID).Scan(&actorName, &actorAvatar)
	if err != nil {
		actorName = "Unknown User"
		actorAvatar = ""
	}

	// Create change description
	changeDesc := ""
	if oldDueDate == "" && newDueDate != "" {
		changeDesc = fmt.Sprintf("Fälligkeitsdatum gesetzt: **%s**", newDueDate)
	} else if oldDueDate != "" && newDueDate == "" {
		changeDesc = "Fälligkeitsdatum entfernt"
	} else {
		changeDesc = fmt.Sprintf("**%s** → **%s**", oldDueDate, newDueDate)
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "📅 Fälligkeitsdatum geändert",
		Description: fmt.Sprintf("**%s** hat das Fälligkeitsdatum deiner Aufgabe im Board **%s** geändert.", actorName, boardName),
		Color:       0xED4245, // Red
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Aufgabe",
				Value:  task.Title,
				Inline: false,
			},
			{
				Name:   "Änderung",
				Value:  changeDesc,
				Inline: false,
			},
		},
	}

	// Add user avatar as thumbnail
	if actorAvatar != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: actorAvatar,
		}
	}

	// Send DM to all users
	for _, userID := range userIDs {
		if userID == actorID {
			continue // Don't notify the actor
		}

		// Check if user should receive notification
		shouldNotify, err := s.shouldSendNotification(userID, guildID, boardID, "due_date_change")
		if err != nil {
			log.Printf("Failed to check notification settings for user %s: %v", userID, err)
			continue
		}
		if !shouldNotify {
			log.Printf("User %s has disabled due date change notifications", userID)
			continue
		}

		s.sendDM(userID, embed)
	}

	return nil
}

// NotifyTaskUnassignment sends a DM to a user when they are unassigned from a task
func (s *TaskNotificationService) NotifyTaskUnassignment(taskID int, unassignedUserID, actorID string) error {
	guildID := os.Getenv("GUILD_ID")

	// Don't notify if user unassigned themselves
	if unassignedUserID == actorID {
		return nil
	}

	// Get task details
	var task tables.Task
	var boardID int
	var boardName string
	err := s.db.QueryRow(`
		SELECT t.id, t.board_id, t.title, b.name
		FROM tasks t
		JOIN boards b ON t.board_id = b.id
		WHERE t.id = $1
	`, taskID).Scan(&task.ID, &boardID, &task.Title, &boardName)
	if err != nil {
		return fmt.Errorf("failed to get task details: %v", err)
	}

	// Check if user should receive notification
	shouldNotify, err := s.shouldSendNotification(unassignedUserID, guildID, boardID, "unassignment")
	if err != nil {
		return fmt.Errorf("failed to check notification settings: %v", err)
	}
	if !shouldNotify {
		log.Printf("User %s has disabled unassignment notifications", unassignedUserID)
		return nil
	}

	// Get actor details
	var actorName, actorAvatar string
	err = s.db.QueryRow(`SELECT display_name, avatar_url FROM users WHERE id = $1`, actorID).Scan(&actorName, &actorAvatar)
	if err != nil {
		actorName = "Unknown User"
		actorAvatar = ""
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "➖ Aufgabenzuweisung entfernt",
		Description: fmt.Sprintf("**%s** hat dich von einer Aufgabe im Board **%s** entfernt.", actorName, boardName),
		Color:       0x95A5A6, // Gray
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Aufgabe",
				Value:  task.Title,
				Inline: false,
			},
			{
				Name:   "Board",
				Value:  boardName,
				Inline: true,
			},
		},
	}

	// Add user avatar as thumbnail
	if actorAvatar != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: actorAvatar,
		}
	}

	// Send DM
	return s.sendDM(unassignedUserID, embed)
}

// NotifyTaskChecklistItem sends a DM to users when a checklist item is added
func (s *TaskNotificationService) NotifyTaskChecklistItem(taskID int, actorID, checklistItemName string, userIDs []string) error {
	guildID := os.Getenv("GUILD_ID")

	// Get task details
	var task tables.Task
	var boardID int
	var boardName string
	err := s.db.QueryRow(`
		SELECT t.id, t.board_id, t.title, b.name
		FROM tasks t
		JOIN boards b ON t.board_id = b.id
		WHERE t.id = $1
	`, taskID).Scan(&task.ID, &boardID, &task.Title, &boardName)
	if err != nil {
		return fmt.Errorf("failed to get task details: %v", err)
	}

	// Get actor details
	var actorName, actorAvatar string
	err = s.db.QueryRow(`SELECT display_name, avatar_url FROM users WHERE id = $1`, actorID).Scan(&actorName, &actorAvatar)
	if err != nil {
		actorName = "Unknown User"
		actorAvatar = ""
	}

	// Create embed
	embed := &discordgo.MessageEmbed{
		Title:       "✅ Neues Checklisten-Element",
		Description: fmt.Sprintf("**%s** hat ein Checklisten-Element zu deiner Aufgabe im Board **%s** hinzugefügt.", actorName, boardName),
		Color:       0xF1C40F, // Gold
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Aufgabe",
				Value:  task.Title,
				Inline: false,
			},
			{
				Name:   "Checklisten-Element",
				Value:  truncateString(checklistItemName, 1024),
				Inline: false,
			},
		},
	}

	// Add user avatar as thumbnail
	if actorAvatar != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: actorAvatar,
		}
	}

	// Send DM to all users
	for _, userID := range userIDs {
		if userID == actorID {
			continue // Don't notify the actor
		}

		// Check if user should receive notification
		shouldNotify, err := s.shouldSendNotification(userID, guildID, boardID, "checklist_item")
		if err != nil {
			log.Printf("Failed to check notification settings for user %s: %v", userID, err)
			continue
		}
		if !shouldNotify {
			log.Printf("User %s has disabled checklist item notifications", userID)
			continue
		}

		s.sendDM(userID, embed)
	}

	return nil
}
