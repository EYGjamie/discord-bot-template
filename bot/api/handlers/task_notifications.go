package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"discord-bot-template/bot/utils/notifications"

	"github.com/bwmarrin/discordgo"
)

// TaskNotificationRequest represents a task notification request
type TaskNotificationRequest struct {
	Type              string   `json:"type"` // "task_assignment", "task_update", "task_comment", "task_due_date_change", "task_unassignment", "task_checklist_item"
	TaskID            int      `json:"task_id"`
	AssigneeID        string   `json:"assignee_id,omitempty"`
	AssignedByID      string   `json:"assigned_by_id,omitempty"`
	UpdatedByID       string   `json:"updated_by_id,omitempty"`
	ActorID           string   `json:"actor_id,omitempty"`
	UserIDs           []string `json:"user_ids,omitempty"`
	ChangeDescription string   `json:"change_description,omitempty"`
	CommentContent    string   `json:"comment_content,omitempty"`
	OldDueDate        string   `json:"old_due_date,omitempty"`
	NewDueDate        string   `json:"new_due_date,omitempty"`
	ChecklistItemName string   `json:"checklist_item_name,omitempty"`
}

// TaskNotificationHandler handles incoming task notification requests
func TaskNotificationHandler(session *discordgo.Session, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TaskNotificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Create notification service
		notificationService := notifications.NewTaskNotificationService(session, db)

		// Handle notification based on type
		var err error
		switch req.Type {
		case "task_assignment":
			if req.AssigneeID == "" || req.AssignedByID == "" {
				http.Error(w, "Missing assignee_id or assigned_by_id", http.StatusBadRequest)
				return
			}
			err = notificationService.NotifyTaskAssignment(req.TaskID, req.AssigneeID, req.AssignedByID)

		case "task_update":
			if req.UpdatedByID == "" {
				http.Error(w, "Missing updated_by_id", http.StatusBadRequest)
				return
			}
			err = notificationService.NotifyTaskUpdate(req.TaskID, req.UpdatedByID, req.ChangeDescription)

		case "task_comment":
			if req.ActorID == "" || len(req.UserIDs) == 0 {
				http.Error(w, "Missing actor_id or user_ids", http.StatusBadRequest)
				return
			}
			err = notificationService.NotifyTaskComment(req.TaskID, req.ActorID, req.CommentContent, req.UserIDs)

		case "task_due_date_change":
			if req.ActorID == "" || len(req.UserIDs) == 0 {
				http.Error(w, "Missing actor_id or user_ids", http.StatusBadRequest)
				return
			}
			err = notificationService.NotifyTaskDueDateChange(req.TaskID, req.ActorID, req.OldDueDate, req.NewDueDate, req.UserIDs)

		case "task_unassignment":
			if req.AssigneeID == "" || req.ActorID == "" {
				http.Error(w, "Missing assignee_id or actor_id", http.StatusBadRequest)
				return
			}
			err = notificationService.NotifyTaskUnassignment(req.TaskID, req.AssigneeID, req.ActorID)

		case "task_checklist_item":
			if req.ActorID == "" || len(req.UserIDs) == 0 {
				http.Error(w, "Missing actor_id or user_ids", http.StatusBadRequest)
				return
			}
			err = notificationService.NotifyTaskChecklistItem(req.TaskID, req.ActorID, req.ChecklistItemName, req.UserIDs)

		default:
			http.Error(w, "Invalid notification type", http.StatusBadRequest)
			return
		}

		if err != nil {
			log.Printf("Error sending task notification: %v", err)
			http.Error(w, "Failed to send notification", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
