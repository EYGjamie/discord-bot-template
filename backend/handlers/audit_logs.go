package handlers

import (
	"database/sql"
	"discord-bot-template/shared/database/tables"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// GetAuditLogs returns audit logs with pagination and filters
func GetAuditLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		query := r.URL.Query()
		userID := query.Get("user_id")
		actionTypeStr := query.Get("action_type")
		limitStr := query.Get("limit")
		offsetStr := query.Get("offset")

		// Default values
		limit := 50
		offset := 0

		// Parse limit
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
				if limit > 1000 {
					limit = 1000 // Max limit
				}
			}
		}

		// Parse offset
		if offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		// Parse action type
		var actionType tables.ActionType
		if actionTypeStr != "" {
			actionType = tables.ActionType(actionTypeStr)
		}

		// Get logs from database
		logs, err := tables.GetWebAppAuditLogs(db, userID, actionType, limit, offset)
		if err != nil {
			http.Error(w, "Failed to fetch audit logs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get total count
		total, err := tables.CountAuditLogs(db, userID, actionType)
		if err != nil {
			http.Error(w, "Failed to count audit logs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Build response
		response := map[string]interface{}{
			"logs":   logs,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// GetUserAuditLogs returns audit logs for a specific user
func GetUserAuditLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from URL path
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 4 {
			http.Error(w, "Invalid URL path", http.StatusBadRequest)
			return
		}
		userID := pathParts[3]

		// Parse query parameters
		query := r.URL.Query()
		actionTypeStr := query.Get("action_type")
		limitStr := query.Get("limit")
		offsetStr := query.Get("offset")

		// Default values
		limit := 50
		offset := 0

		// Parse limit
		if limitStr != "" {
			if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
				if limit > 1000 {
					limit = 1000 // Max limit
				}
			}
		}

		// Parse offset
		if offsetStr != "" {
			if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		// Parse action type
		var actionType tables.ActionType
		if actionTypeStr != "" {
			actionType = tables.ActionType(actionTypeStr)
		}

		// Get logs from database
		logs, err := tables.GetWebAppAuditLogs(db, userID, actionType, limit, offset)
		if err != nil {
			http.Error(w, "Failed to fetch user audit logs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get total count
		total, err := tables.CountAuditLogs(db, userID, actionType)
		if err != nil {
			http.Error(w, "Failed to count user audit logs: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Build response
		response := map[string]interface{}{
			"user_id": userID,
			"logs":    logs,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
