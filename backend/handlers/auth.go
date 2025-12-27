package handlers

import (
	"database/sql"
	"discord-bot-template/shared/database/tables"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
}

type DiscordTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	jwt.RegisteredClaims
}

func generateJWT(user *DiscordUser) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-secret-key-change-this-in-production"
	}

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

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
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
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

	clientID := os.Getenv("DISCORD_CLIENT_ID")
	clientSecret := os.Getenv("DISCORD_CLIENT_SECRET")
	redirectURL := os.Getenv("DISCORD_REDIRECT_URL")

	// Exchange code for access token
	tokenURL := "https://discord.com/api/oauth2/token"
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURL)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get access token from Discord", http.StatusInternalServerError)
		return
	}

	var tokenResp DiscordTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		http.Error(w, "Failed to decode token response", http.StatusInternalServerError)
		return
	}

	// Fetch user data from Discord
	userReq, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		http.Error(w, "Failed to create user request", http.StatusInternalServerError)
		return
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		http.Error(w, "Failed to fetch user data from Discord", http.StatusInternalServerError)
		return
	}
	defer userResp.Body.Close()

	if userResp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to get user data from Discord", http.StatusInternalServerError)
		return
	}

	var discordUser DiscordUser
	if err := json.NewDecoder(userResp.Body).Decode(&discordUser); err != nil {
		http.Error(w, "Failed to decode user data", http.StatusInternalServerError)
		return
	}

	// Check if user exists in database (from users table)
	user, err := tables.GetUserByID(h.db, discordUser.ID)
	if err != nil {
		// User doesn't exist in database yet - this is okay for web login
		// We'll still create a JWT token for them
		fmt.Printf("User %s not found in database (might not be a guild member yet)\n", discordUser.ID)
	}

	// Generate JWT token
	jwtToken, err := generateJWT(&discordUser)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Redirect to frontend with token
	frontendURL := "http://localhost:5173/auth/callback?token=" + jwtToken
	if user != nil {
		// Add user_id parameter for easier frontend handling
		frontendURL += "&user_id=" + user.ID
	}
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

// GetCurrentUser returns the authenticated user's data
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get JWT token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	// Extract token (format: "Bearer <token>")
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	claims, err := validateJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Try to fetch full user data from database
	user, err := tables.GetUserByID(h.db, claims.UserID)

	// Build response
	response := map[string]interface{}{
		"id":            claims.UserID,
		"discord_id":    claims.UserID,
		"username":      claims.Username,
		"email":         claims.Email,
		"avatar":        claims.Avatar,
		"discriminator": "0", // Discord removed discriminators
		"is_admin":      false,
		"is_moderator":  false,
		"created_at":    time.Now().Format(time.RFC3339),
		"updated_at":    time.Now().Format(time.RFC3339),
		"roles":         []map[string]interface{}{},
	}

	// If user exists in database, add more details
	if err == nil && user != nil {
		response["discord_id"] = user.ID
		response["username"] = user.Name
		response["global_name"] = user.GlobalName
		response["display_name"] = user.DisplayName
		response["avatar_url"] = user.AvatarURL
		response["nick"] = user.Nick
		response["joined_at"] = user.JoinedAt
		response["created_at"] = user.CreatedAt.Format(time.RFC3339)
		response["updated_at"] = user.UpdatedAt.Format(time.RFC3339)

		// Fetch user roles
		roles, roleErr := tables.GetUserRoles(h.db, claims.UserID)
		if roleErr == nil && len(roles) > 0 {
			// Convert roles to response format
			roleList := make([]map[string]interface{}, len(roles))
			for i, role := range roles {
				roleList[i] = map[string]interface{}{
					"id":       role.ID,
					"name":     role.Name,
					"color":    role.Color,
					"position": role.Position,
				}
			}
			response["roles"] = roleList

			// Check for moderator/admin permissions
			response["is_moderator"] = h.checkIsModerator(roles)
			response["is_admin"] = h.checkIsAdmin(roles)
		}
	} else {
		// Generate avatar URL from Discord CDN
		if claims.Avatar != "" {
			response["avatar_url"] = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", claims.UserID, claims.Avatar)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// checkIsModerator checks if user has any moderator role
func (h *AuthHandler) checkIsModerator(roles []*tables.Role) bool {
	modRoleIDs := os.Getenv("MODERATOR_ROLE_IDS")
	if modRoleIDs == "" {
		return false
	}

	modIDs := splitAndTrim(modRoleIDs)
	for _, role := range roles {
		for _, modID := range modIDs {
			if role.ID == modID {
				return true
			}
		}
	}

	// Also check admin roles (admins are also moderators)
	return h.checkIsAdmin(roles)
}

// checkIsAdmin checks if user has any admin role
func (h *AuthHandler) checkIsAdmin(roles []*tables.Role) bool {
	adminRoleIDs := os.Getenv("ADMIN_ROLE_IDS")
	if adminRoleIDs == "" {
		return false
	}

	adminIDs := splitAndTrim(adminRoleIDs)
	for _, role := range roles {
		for _, adminID := range adminIDs {
			if role.ID == adminID {
				return true
			}
		}
	}
	return false
}

// splitAndTrim splits a comma-separated string and trims whitespace
func splitAndTrim(s string) []string {
	result := []string{}
	current := ""
	for _, char := range s {
		if char == ',' {
			if trimmed := trimSpace(current); trimmed != "" {
				result = append(result, trimmed)
			}
			current = ""
		} else {
			current += string(char)
		}
	}
	if trimmed := trimSpace(current); trimmed != "" {
		result = append(result, trimmed)
	}
	return result
}

// trimSpace removes leading and trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
