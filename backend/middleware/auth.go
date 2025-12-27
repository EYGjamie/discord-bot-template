package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// validateJWT validates a JWT token and returns the claims
func validateJWT(tokenString string) (*Claims, error) {
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

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, fmt.Errorf("token expired")
		}
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetUserIDFromRequest extrahiert die User-ID aus dem Request
// Unterstützt mehrere Methoden: JWT Token, Cookie, Header
func GetUserIDFromRequest(r *http.Request) string {
	// 1. Versuche aus Authorization Header (Bearer JWT Token)
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := validateJWT(tokenString)
			if err == nil && claims.UserID != "" {
				return claims.UserID
			}
		}
	}

	// 2. Versuche aus Cookie zu lesen
	if cookie, err := r.Cookie("user_id"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 3. Versuche aus X-User-ID Header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
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
		ctx := NewContextWithUserID(r.Context(), userID)
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

// NewContextWithUserID erstellt einen neuen Context mit User ID
func NewContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
