package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// NotificationRequest represents a request to send a notification via the bot
type NotificationRequest struct {
	Type              string   `json:"type"` // "task_assignment", "task_update", "task_comment", "task_due_date_change", "task_unassignment", "task_checklist_item"
	TaskID            int      `json:"task_id"`
	AssigneeID        string   `json:"assignee_id,omitempty"`
	AssignedByID      string   `json:"assigned_by_id,omitempty"`
	UpdatedByID       string   `json:"updated_by_id,omitempty"`
	ActorID           string   `json:"actor_id,omitempty"` // ID of the user who triggered the notification (for avatar)
	UserIDs           []string `json:"user_ids,omitempty"` // List of user IDs to notify (for comments, checklist items)
	ChangeDescription string   `json:"change_description,omitempty"`
	CommentContent    string   `json:"comment_content,omitempty"`
	OldDueDate        string   `json:"old_due_date,omitempty"`
	NewDueDate        string   `json:"new_due_date,omitempty"`
	ChecklistItemName string   `json:"checklist_item_name,omitempty"`
}

// SendTaskNotification sends a notification request to the bot API
func SendTaskNotification(notificationType string, taskID int, assigneeID, actorID, changeDescription string) error {
	botAPIURL := os.Getenv("BOT_API_URL")
	if botAPIURL == "" {
		botAPIURL = "http://localhost:8081" // Default bot API URL
	}

	req := NotificationRequest{
		Type:              notificationType,
		TaskID:            taskID,
		ChangeDescription: changeDescription,
		ActorID:           actorID,
	}

	switch notificationType {
	case "task_assignment":
		req.AssigneeID = assigneeID
		req.AssignedByID = actorID
	case "task_update":
		req.UpdatedByID = actorID
	case "task_unassignment":
		req.AssigneeID = assigneeID
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %v", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/notifications/task", botAPIURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Printf("Failed to send task notification: %v", err)
		return nil // Don't fail the request if notification fails
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Bot API returned error status: %d", resp.StatusCode)
	}

	return nil
}

// SendTaskCommentNotification sends a notification when a comment is added
func SendTaskCommentNotification(taskID int, actorID, commentContent string, userIDs []string) error {
	botAPIURL := os.Getenv("BOT_API_URL")
	if botAPIURL == "" {
		botAPIURL = "http://localhost:8081"
	}

	req := NotificationRequest{
		Type:           "task_comment",
		TaskID:         taskID,
		ActorID:        actorID,
		CommentContent: commentContent,
		UserIDs:        userIDs,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %v", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/notifications/task", botAPIURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Printf("Failed to send comment notification: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Bot API returned error status: %d", resp.StatusCode)
	}

	return nil
}

// SendTaskDueDateChangeNotification sends a notification when due date changes
func SendTaskDueDateChangeNotification(taskID int, actorID, oldDueDate, newDueDate string, userIDs []string) error {
	botAPIURL := os.Getenv("BOT_API_URL")
	if botAPIURL == "" {
		botAPIURL = "http://localhost:8081"
	}

	req := NotificationRequest{
		Type:       "task_due_date_change",
		TaskID:     taskID,
		ActorID:    actorID,
		OldDueDate: oldDueDate,
		NewDueDate: newDueDate,
		UserIDs:    userIDs,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %v", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/notifications/task", botAPIURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Printf("Failed to send due date change notification: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Bot API returned error status: %d", resp.StatusCode)
	}

	return nil
}

// SendTaskChecklistItemNotification sends a notification when a checklist item is added
func SendTaskChecklistItemNotification(taskID int, actorID, checklistItemName string, userIDs []string) error {
	botAPIURL := os.Getenv("BOT_API_URL")
	if botAPIURL == "" {
		botAPIURL = "http://localhost:8081"
	}

	req := NotificationRequest{
		Type:              "task_checklist_item",
		TaskID:            taskID,
		ActorID:           actorID,
		ChecklistItemName: checklistItemName,
		UserIDs:           userIDs,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %v", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/notifications/task", botAPIURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Printf("Failed to send checklist item notification: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Bot API returned error status: %d", resp.StatusCode)
	}

	return nil
}
