package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUserInfo represents user information from Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
}

// OAuthService handles Google OAuth2 authentication
type OAuthService struct {
	config      *oauth2.Config
	userRepo    UserRepository
	authService *AuthService
	logger      *slog.Logger
	stateCache  map[string]time.Time // In production, use Redis
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(
	config OAuthConfig,
	userRepo UserRepository,
	authService *AuthService,
	logger *slog.Logger,
) *OAuthService {
	if len(config.Scopes) == 0 {
		config.Scopes = []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		}
	}

	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       config.Scopes,
		Endpoint:     google.Endpoint,
	}

	return &OAuthService{
		config:      oauth2Config,
		userRepo:    userRepo,
		authService: authService,
		logger:      logger,
		stateCache:  make(map[string]time.Time),
	}
}

// GetAuthURL generates the Google OAuth authorization URL
func (os *OAuthService) GetAuthURL() (string, string, error) {
	// Generate secure random state
	state, err := os.generateState()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Store state with expiration (5 minutes)
	os.stateCache[state] = time.Now().Add(5 * time.Minute)

	// Clean up expired states
	go os.cleanupExpiredStates()

	authURL := os.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return authURL, state, nil
}

// HandleCallback processes the OAuth callback and authenticates the user
func (os *OAuthService) HandleCallback(ctx context.Context, code, state string) (*AuthResponse, error) {
	// Validate state
	if !os.validateState(state) {
		os.logger.Warn("invalid OAuth state", "state", state)
		return nil, fmt.Errorf("invalid state parameter")
	}

	// Remove used state
	delete(os.stateCache, state)

	// Exchange code for token
	token, err := os.config.Exchange(ctx, code)
	if err != nil {
		os.logger.Error("failed to exchange OAuth code", "error", err)
		return nil, fmt.Errorf("failed to exchange authorization code")
	}

	// Get user info from Google
	userInfo, err := os.getUserInfo(ctx, token)
	if err != nil {
		os.logger.Error("failed to get user info from Google", "error", err)
		return nil, fmt.Errorf("failed to get user information")
	}

	// Check if user exists by Google ID
	existingUser, err := os.userRepo.GetUserByGoogleID(ctx, userInfo.ID)
	if err == nil {
		// Existing user - perform login
		return os.loginExistingUser(ctx, existingUser)
	}

	// Check if user exists by email
	existingUser, err = os.userRepo.GetUserByEmail(ctx, strings.ToLower(userInfo.Email))
	if err == nil {
		// Link Google account to existing user
		return os.linkGoogleAccount(ctx, existingUser, userInfo)
	}

	// Create new user
	return os.createNewUser(ctx, userInfo)
}

// generateState creates a cryptographically secure random state
func (os *OAuthService) generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// validateState checks if the state is valid and not expired
func (os *OAuthService) validateState(state string) bool {
	expiry, exists := os.stateCache[state]
	if !exists {
		return false
	}
	return time.Now().Before(expiry)
}

// cleanupExpiredStates removes expired states from cache
func (os *OAuthService) cleanupExpiredStates() {
	now := time.Now()
	for state, expiry := range os.stateCache {
		if now.After(expiry) {
			delete(os.stateCache, state)
		}
	}
}

// getUserInfo fetches user information from Google API
func (os *OAuthService) getUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := os.config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// loginExistingUser handles login for existing users
func (os *OAuthService) loginExistingUser(ctx context.Context, user *User) (*AuthResponse, error) {
	// Check if account is locked
	if user.IsLocked() {
		os.logger.Warn("OAuth login attempt on locked account", "user_id", user.ID)
		return nil, fmt.Errorf("account is temporarily locked")
	}

	// Update last login
	err := os.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		os.logger.Error("failed to update last login", "user_id", user.ID, "error", err)
	}

	// Generate tokens
	roles := []string{"user"}
	tokenPair, err := os.authService.jwtService.GenerateTokenPair(user.ID, user.Email, roles)
	if err != nil {
		os.logger.Error("failed to generate tokens", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate authentication tokens")
	}

	// Store refresh token
	err = os.authService.storeRefreshToken(ctx, user.ID, tokenPair.RefreshToken)
	if err != nil {
		os.logger.Error("failed to store refresh token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to store refresh token")
	}

	os.logger.Info("user logged in via Google OAuth", "user_id", user.ID, "email", user.Email)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	}, nil
}

// linkGoogleAccount links a Google account to an existing user
func (os *OAuthService) linkGoogleAccount(ctx context.Context, user *User, userInfo *GoogleUserInfo) (*AuthResponse, error) {
	// Update user with Google ID and verify email if Google says it's verified
	user.GoogleID = &userInfo.ID
	if userInfo.VerifiedEmail {
		user.EmailVerified = true
	}
	user.UpdatedAt = time.Now()

	err := os.userRepo.UpdateUser(ctx, user)
	if err != nil {
		os.logger.Error("failed to link Google account", "user_id", user.ID, "google_id", userInfo.ID, "error", err)
		return nil, fmt.Errorf("failed to link Google account")
	}

	os.logger.Info("Google account linked to existing user", "user_id", user.ID, "google_id", userInfo.ID)

	// Proceed with login
	return os.loginExistingUser(ctx, user)
}

// createNewUser creates a new user from Google OAuth information
func (os *OAuthService) createNewUser(ctx context.Context, userInfo *GoogleUserInfo) (*AuthResponse, error) {
	// Create new user
	user := &User{
		ID:            uuid.New().String(),
		Email:         strings.ToLower(userInfo.Email),
		Name:          userInfo.Name,
		GoogleID:      &userInfo.ID,
		EmailVerified: userInfo.VerifiedEmail,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := os.userRepo.CreateUser(ctx, user)
	if err != nil {
		os.logger.Error("failed to create user from Google OAuth", "email", userInfo.Email, "google_id", userInfo.ID, "error", err)
		return nil, fmt.Errorf("failed to create user account")
	}

	// Generate tokens
	roles := []string{"user"}
	tokenPair, err := os.authService.jwtService.GenerateTokenPair(user.ID, user.Email, roles)
	if err != nil {
		os.logger.Error("failed to generate tokens for new user", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate authentication tokens")
	}

	// Store refresh token
	err = os.authService.storeRefreshToken(ctx, user.ID, tokenPair.RefreshToken)
	if err != nil {
		os.logger.Error("failed to store refresh token for new user", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to store refresh token")
	}

	os.logger.Info("new user created via Google OAuth", "user_id", user.ID, "email", user.Email, "google_id", userInfo.ID)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	}, nil
}
