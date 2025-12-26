package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// GetUserIDFromRequest extrahiert die User-ID aus dem Request
// Unterstützt mehrere Methoden: Cookie, Header, Query-Parameter
func GetUserIDFromRequest(r *http.Request) string {
	// 1. Versuche aus Cookie zu lesen
	if cookie, err := r.Cookie("user_id"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 2. Versuche aus Authorization Header (Bearer Token)
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		// Für einfaches Testing: Bearer {discord_id}
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 3. Versuche aus X-User-ID Header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}

	// 4. Versuche aus Query-Parameter (nur für Testing)
	if userID := r.URL.Query().Get("user_id"); userID != "" {
		return userID
	}

	return ""
}

// RequireAuth ist ein Middleware, das sicherstellt, dass ein User eingeloggt ist
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserIDFromRequest(r)
		if userID == "" {
			http.Error(w, "Unauthorized: User not authenticated", http.StatusUnauthorized)
			return
		}

		// User-ID im Context speichern für spätere Verwendung
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserIDFromContext holt die User-ID aus dem Request Context
func GetUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}
