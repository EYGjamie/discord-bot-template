package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// DiscordLogin initiates the Discord OAuth flow
func (h *AuthHandler) DiscordLogin(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("DISCORD_CLIENT_ID")
	redirectURL := os.Getenv("DISCORD_REDIRECT_URL")

	// Debug: Log the values (remove in production!)
	if clientID == "" {
		http.Error(w, "DISCORD_CLIENT_ID not set", http.StatusInternalServerError)
		return
	}
	if redirectURL == "" {
		http.Error(w, "DISCORD_REDIRECT_URL not set", http.StatusInternalServerError)
		return
	}

	// Build Discord authorization URL
	authURL := "https://discord.com/api/oauth2/authorize?" +
		"client_id=" + clientID +
		"&redirect_uri=" + url.QueryEscape(redirectURL) +
		"&response_type=code" +
		"&scope=identify%20email%20guilds"

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// DiscordCallback handles the OAuth callback from Discord
func (h *AuthHandler) DiscordCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// TODO: Exchange code for access token with Discord API
	// TODO: Fetch user data from Discord
	// TODO: Save/update user in database
	// TODO: Generate JWT token

	// For now: Mock token for testing
	mockToken := "mock-jwt-token-12345"

	// Redirect to frontend with token
	frontendURL := "http://localhost:5173/auth/callback?token=" + mockToken
	http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Successfully logged out",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentUser returns mock user data
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// TODO: Validate JWT token from Authorization header
	// TODO: Fetch user from database

	// Mock response
	user := map[string]interface{}{
		"id":       "123456789",
		"username": "TestUser",
		"email":    "test@example.com",
		"avatar":   "https://cdn.discordapp.com/avatars/123456789/avatar.png",
		"is_admin": false,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
