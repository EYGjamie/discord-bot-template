package middleware

import (
	"database/sql"
	"discord-bot-template/shared/database/tables"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims repräsentiert die JWT Claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	jwt.RegisteredClaims
}

// AuditLogger ist ein Middleware für Audit-Logging
type AuditLogger struct {
	db *sql.DB
}

// NewAuditLogger erstellt eine neue AuditLogger Middleware Instanz
func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Middleware fügt Audit-Logging zu HTTP-Requests hinzu
func (al *AuditLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Response Writer wrappen um Status Code zu erfassen
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// User ID aus JWT Token extrahieren
		userID := al.extractUserID(r)

		// User ID im Context speichern für nachfolgende Handler
		if userID != "" {
			ctx := NewContextWithUserID(r.Context(), userID)
			r = r.WithContext(ctx)
		}

		// Request verarbeiten
		next.ServeHTTP(wrapped, r)

		// Audit Log erstellen
		go al.logRequest(r, wrapped.statusCode, userID, time.Since(startTime))
	})
}

// extractUserID extrahiert die User ID aus dem JWT Token oder Headers
func (al *AuditLogger) extractUserID(r *http.Request) string {
	// 1. Versuche JWT Token aus Authorization Header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Token extrahieren (format: "Bearer <token>")
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			tokenString = authHeader
		}

		// JWT validieren
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "your-secret-key-change-this-in-production"
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err == nil {
			if claims, ok := token.Claims.(*Claims); ok && token.Valid {
				log.Printf("User ID extracted from JWT: %s", claims.UserID)
				return claims.UserID
			}
		} else {
			log.Printf("JWT validation failed: %v", err)
		}
	}

	// 2. Fallback: X-User-ID Header (für Requests die direkt die User ID senden)
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		log.Printf("User ID extracted from X-User-ID header: %s", userID)
		return userID
	}

	// 3. Fallback: Versuche aus Context (wenn bereits von anderer Middleware gesetzt)
	if userID := GetUserIDFromContext(r.Context()); userID != "" {
		log.Printf("User ID extracted from context: %s", userID)
		return userID
	}

	log.Printf("No user ID found for request: %s %s", r.Method, r.URL.Path)
	return ""
}

// logRequest speichert den Request in der Datenbank
func (al *AuditLogger) logRequest(r *http.Request, statusCode int, userID string, duration time.Duration) {
	// Nur relevante API-Calls loggen (nicht static files, health checks, etc.)
	if !al.shouldLog(r.URL.Path, statusCode) {
		return
	}

	// Action Type basierend auf HTTP Method und Path bestimmen
	actionType := al.determineActionType(r.Method, r.URL.Path)
	resourceType, resourceID := al.extractResource(r.URL.Path)

	// IP-Adresse ermitteln
	ipAddress := al.getClientIP(r)

	// Details sammeln
	details := make(map[string]interface{})
	details["duration_ms"] = duration.Milliseconds()
	if len(r.URL.Query()) > 0 {
		details["query_params"] = r.URL.Query()
	}

	// Audit Log erstellen
	auditLog := &tables.WebAppAuditLog{
		UserID:        userID,
		ActionType:    actionType,
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		IPAddress:     ipAddress,
		UserAgent:     r.UserAgent(),
		RequestMethod: r.Method,
		RequestPath:   r.URL.Path,
		StatusCode:    statusCode,
		Details:       details,
	}

	// In Datenbank speichern
	if err := tables.InsertWebAppAuditLog(al.db, auditLog); err != nil {
		log.Printf("Error inserting audit log for %s %s: %v", r.Method, r.URL.Path, err)
	} else {
		log.Printf("Audit log created: User=%s, Action=%s, Resource=%s, Path=%s, Status=%d",
			userID, actionType, resourceType, r.URL.Path, statusCode)
	}
}

// shouldLog bestimmt ob ein Request geloggt werden soll
func (al *AuditLogger) shouldLog(path string, statusCode int) bool {
	// Nicht loggen: 404 Not Found
	if statusCode == http.StatusNotFound {
		return false
	}

	// Nicht loggen: Static files, health checks, etc.
	excludedPaths := []string{
		"/health",
		"/websocket",
		"/favicon.ico",
		"/static/",
	}

	for _, excluded := range excludedPaths {
		if strings.HasPrefix(path, excluded) {
			return false
		}
	}

	// Nur API-Calls loggen
	return strings.HasPrefix(path, "/api/")
}

// determineActionType bestimmt den Action Type basierend auf Method und Path
func (al *AuditLogger) determineActionType(method, path string) tables.ActionType {
	// Login/Logout
	if strings.Contains(path, "/auth/discord/login") {
		return tables.ActionTypeLogin
	}
	if strings.Contains(path, "/auth/logout") {
		return tables.ActionTypeLogout
	}

	// Moderation
	if strings.Contains(path, "/moderation/warns") && method == "POST" {
		return tables.ActionTypeWarnCreate
	}
	if strings.Contains(path, "/moderation/warns") && method == "DELETE" {
		return tables.ActionTypeWarnDelete
	}
	if strings.Contains(path, "/moderation/notes") && method == "POST" {
		return tables.ActionTypeNoteCreate
	}
	if strings.Contains(path, "/moderation/notes") && method == "DELETE" {
		return tables.ActionTypeNoteDelete
	}

	// Dashboard
	if strings.Contains(path, "/dashboard") {
		return tables.ActionTypeDashboardView
	}

	// Members
	if strings.Contains(path, "/members") {
		if method == "GET" {
			if strings.Contains(path, "search") {
				return tables.ActionTypeMemberSearch
			}
			return tables.ActionTypeView
		}
	}

	// Settings
	if strings.Contains(path, "/settings") {
		if method == "PUT" || method == "PATCH" || method == "POST" {
			return tables.ActionTypeSettingsChange
		}
		return tables.ActionTypeView
	}

	// Default basierend auf HTTP Method
	switch method {
	case "GET":
		return tables.ActionTypeView
	case "POST":
		return tables.ActionTypeCreate
	case "PUT", "PATCH":
		return tables.ActionTypeUpdate
	case "DELETE":
		return tables.ActionTypeDelete
	default:
		return tables.ActionTypeAPICall
	}
}

// extractResource extrahiert ResourceType und ResourceID aus dem Path
func (al *AuditLogger) extractResource(path string) (tables.ResourceType, string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	// /api/members/123 -> MEMBER, 123
	if len(parts) >= 3 && parts[1] == "members" {
		resourceID := ""
		if len(parts) >= 3 {
			resourceID = parts[2]
		}
		return tables.ResourceTypeMember, resourceID
	}

	// /api/moderation/warns/123 -> WARN, 123
	if len(parts) >= 4 && parts[1] == "moderation" {
		if parts[2] == "warns" {
			resourceID := ""
			if len(parts) >= 4 {
				resourceID = parts[3]
			}
			return tables.ResourceTypeWarn, resourceID
		}
		if parts[2] == "notes" {
			resourceID := ""
			if len(parts) >= 4 {
				resourceID = parts[3]
			}
			return tables.ResourceTypeNote, resourceID
		}
	}

	// /api/dashboard -> DASHBOARD
	if len(parts) >= 2 && parts[1] == "dashboard" {
		return tables.ResourceTypeDashboard, ""
	}

	// /api/settings -> SETTINGS
	if len(parts) >= 2 && parts[1] == "settings" {
		return tables.ResourceTypeSettings, ""
	}

	// /api/me -> USER
	if len(parts) >= 2 && parts[1] == "me" {
		return tables.ResourceTypeUser, ""
	}

	return tables.ResourceTypeAPI, ""
}

// getClientIP ermittelt die Client IP-Adresse
func (al *AuditLogger) getClientIP(r *http.Request) string {
	// X-Forwarded-For Header prüfen (bei Proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Nimm die erste IP aus der Liste
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// X-Real-IP Header prüfen
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// RemoteAddr verwenden
	return strings.Split(r.RemoteAddr, ":")[0]
}
