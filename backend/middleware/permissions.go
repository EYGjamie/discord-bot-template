package middleware

import (
	"context"
	"database/sql"
	"discord-bot-template/shared/database/tables"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

// Permission levels
const (
	PermissionPublic    = "public"    // Everyone can access
	PermissionMember    = "member"    // Only guild members
	PermissionModerator = "moderator" // Only moderators
	PermissionAdmin     = "admin"     // Only admins
)

const PermissionsKey contextKey = "permissions"

// UserPermissions holds the user's role information and permissions
type UserPermissions struct {
	UserID      string
	Roles       []*tables.Role
	IsModerator bool
	IsAdmin     bool
	HighestRole *tables.Role
}

// PermissionChecker provides methods to check user permissions
type PermissionChecker struct {
	db *sql.DB
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(db *sql.DB) *PermissionChecker {
	return &PermissionChecker{db: db}
}

// GetUserPermissions fetches and calculates user permissions based on their Discord roles
func (pc *PermissionChecker) GetUserPermissions(userID string) (*UserPermissions, error) {
	// Get user's roles from database
	roles, err := tables.GetUserRoles(pc.db, userID)
	if err != nil {
		// User might not have any roles yet
		return &UserPermissions{
			UserID:      userID,
			Roles:       []*tables.Role{},
			IsModerator: false,
			IsAdmin:     false,
		}, nil
	}

	perms := &UserPermissions{
		UserID:      userID,
		Roles:       roles,
		IsModerator: false,
		IsAdmin:     false,
	}

	// Get moderator and admin role IDs from environment
	modRoleIDs := strings.Split(os.Getenv("MODERATOR_ROLE_IDS"), ",")
	adminRoleIDs := strings.Split(os.Getenv("ADMIN_ROLE_IDS"), ",")

	// Find highest role and check for moderator/admin
	var highestRole *tables.Role
	for _, role := range roles {
		// Check if user has moderator role
		for _, modRoleID := range modRoleIDs {
			if strings.TrimSpace(modRoleID) != "" && role.ID == strings.TrimSpace(modRoleID) {
				perms.IsModerator = true
			}
		}

		// Check if user has admin role
		for _, adminRoleID := range adminRoleIDs {
			if strings.TrimSpace(adminRoleID) != "" && role.ID == strings.TrimSpace(adminRoleID) {
				perms.IsAdmin = true
			}
		}

		// Track highest role by position
		if highestRole == nil || role.Position > highestRole.Position {
			highestRole = role
		}
	}

	perms.HighestRole = highestRole

	// Admins are also moderators
	if perms.IsAdmin {
		perms.IsModerator = true
	}

	return perms, nil
}

// HasPermission checks if the user has the required permission level
func (up *UserPermissions) HasPermission(required string) bool {
	switch required {
	case PermissionPublic:
		return true
	case PermissionMember:
		return len(up.Roles) > 0 // User must be in the guild
	case PermissionModerator:
		return up.IsModerator
	case PermissionAdmin:
		return up.IsAdmin
	default:
		return false
	}
}

// RequirePermission is a middleware that checks if user has required permission level
func (pc *PermissionChecker) RequirePermission(level string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (set by RequireAuth middleware)
			userID := GetUserIDFromContext(r.Context())
			if userID == "" {
				http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
				return
			}

			// Get user permissions
			perms, err := pc.GetUserPermissions(userID)
			if err != nil {
				http.Error(w, "Failed to check permissions", http.StatusInternalServerError)
				return
			}

			// Check if user has required permission
			if !perms.HasPermission(level) {
				response := map[string]interface{}{
					"error":    "Forbidden: Insufficient permissions",
					"required": level,
					"current": map[string]bool{
						"is_moderator": perms.IsModerator,
						"is_admin":     perms.IsAdmin,
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(response)
				return
			}

			// Add permissions to context for later use
			ctx := context.WithValue(r.Context(), PermissionsKey, perms)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// GetPermissionsFromContext retrieves permissions from request context
func GetPermissionsFromContext(ctx context.Context) *UserPermissions {
	if perms, ok := ctx.Value(PermissionsKey).(*UserPermissions); ok {
		return perms
	}
	return nil
}

// WithPermissions is a middleware that adds user permissions to the context without enforcing any level
func (pc *PermissionChecker) WithPermissions(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserIDFromContext(r.Context())
		if userID != "" {
			perms, err := pc.GetUserPermissions(userID)
			if err == nil {
				ctx := context.WithValue(r.Context(), PermissionsKey, perms)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}
