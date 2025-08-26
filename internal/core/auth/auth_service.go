package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuthService handles user authentication operations
type AuthService struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	passwordService  *PasswordService
	jwtService       *JWTService
	logger           *slog.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo UserRepository,
	refreshTokenRepo RefreshTokenRepository,
	passwordService *PasswordService,
	jwtService *JWTService,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		passwordService:  passwordService,
		jwtService:       jwtService,
		logger:           logger,
	}
}

// Register creates a new user account
func (as *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// Validate input
	if err := as.validateRegisterRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if email already exists
	exists, err := as.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		as.logger.Error("failed to check email existence", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to check email availability")
	}
	if exists {
		return nil, fmt.Errorf("email address already registered")
	}

	// Validate password strength
	if err := as.passwordService.ValidatePasswordStrength(req.Password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash password
	hashedPassword, err := as.passwordService.HashPassword(req.Password)
	if err != nil {
		as.logger.Error("failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to process password")
	}

	// Create user
	user := &User{
		ID:                  uuid.New().String(),
		Email:               strings.ToLower(strings.TrimSpace(req.Email)),
		Name:                strings.TrimSpace(req.Name),
		PasswordHash:        &hashedPassword,
		EmailVerified:       false,
		FailedLoginAttempts: 0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	err = as.userRepo.CreateUser(ctx, user)
	if err != nil {
		as.logger.Error("failed to create user", "email", req.Email, "error", err)
		return nil, fmt.Errorf("failed to create user account")
	}

	// Generate tokens
	tokenPair, err := as.jwtService.GenerateTokenPair(user.ID, user.Email, []string{"user"})
	if err != nil {
		as.logger.Error("failed to generate tokens", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate authentication tokens")
	}

	// Store refresh token
	err = as.storeRefreshToken(ctx, user.ID, tokenPair.RefreshToken)
	if err != nil {
		as.logger.Error("failed to store refresh token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to store refresh token")
	}

	as.logger.Info("user registered successfully", "user_id", user.ID, "email", user.Email)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	}, nil
}

// Login authenticates a user with email and password
func (as *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// Validate input
	if err := as.validateLoginRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get user by email
	user, err := as.userRepo.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		as.logger.Error("failed to get user by email", "email", req.Email, "error", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is locked
	if user.IsLocked() {
		as.logger.Warn("login attempt on locked account", "user_id", user.ID, "locked_until", user.LockedUntil)
		return nil, fmt.Errorf("account is temporarily locked due to too many failed attempts")
	}

	// Verify password
	if user.PasswordHash == nil {
		as.logger.Warn("login attempt on account without password", "user_id", user.ID)
		return nil, fmt.Errorf("invalid credentials")
	}

	err = as.passwordService.VerifyPassword(*user.PasswordHash, req.Password)
	if err != nil {
		// Handle failed login attempt
		return as.handleFailedLogin(ctx, user)
	}

	// Reset failed login attempts on successful login
	if user.FailedLoginAttempts > 0 {
		err = as.userRepo.UpdateUserLoginAttempts(ctx, user.ID, 0, nil)
		if err != nil {
			as.logger.Error("failed to reset login attempts", "user_id", user.ID, "error", err)
		}
	}

	// Update last login
	err = as.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		as.logger.Error("failed to update last login", "user_id", user.ID, "error", err)
	}

	// Generate tokens
	roles := []string{"user"}
	// Add additional roles based on user properties if needed

	tokenPair, err := as.jwtService.GenerateTokenPair(user.ID, user.Email, roles)
	if err != nil {
		as.logger.Error("failed to generate tokens", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate authentication tokens")
	}

	// Store refresh token
	err = as.storeRefreshToken(ctx, user.ID, tokenPair.RefreshToken)
	if err != nil {
		as.logger.Error("failed to store refresh token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to store refresh token")
	}

	as.logger.Info("user logged in successfully", "user_id", user.ID, "email", user.Email)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	}, nil
}

// RefreshToken generates new tokens using a valid refresh token
func (as *AuthService) RefreshToken(ctx context.Context, refreshTokenString string) (*AuthResponse, error) {
	if refreshTokenString == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	// Validate refresh token
	claims, err := as.jwtService.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		as.logger.Warn("invalid refresh token", "error", err)
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if token exists in database
	tokenHash := as.jwtService.HashRefreshToken(refreshTokenString)
	storedToken, err := as.refreshTokenRepo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		as.logger.Warn("refresh token not found in database", "user_id", claims.UserID)
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if token is still valid
	if !storedToken.IsValid() {
		as.logger.Warn("refresh token is expired or revoked", "user_id", claims.UserID)
		return nil, fmt.Errorf("refresh token is expired or revoked")
	}

	// Get user
	user, err := as.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		as.logger.Error("failed to get user for token refresh", "user_id", claims.UserID, "error", err)
		return nil, fmt.Errorf("user not found")
	}

	// Check if user account is locked
	if user.IsLocked() {
		as.logger.Warn("token refresh attempt on locked account", "user_id", user.ID)
		return nil, fmt.Errorf("account is temporarily locked")
	}

	// Revoke old refresh token
	err = as.refreshTokenRepo.RevokeRefreshToken(ctx, tokenHash)
	if err != nil {
		as.logger.Error("failed to revoke old refresh token", "user_id", user.ID, "error", err)
	}

	// Generate new tokens
	tokenPair, err := as.jwtService.GenerateTokenPair(user.ID, user.Email, claims.Roles)
	if err != nil {
		as.logger.Error("failed to generate new tokens", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate new tokens")
	}

	// Store new refresh token
	err = as.storeRefreshToken(ctx, user.ID, tokenPair.RefreshToken)
	if err != nil {
		as.logger.Error("failed to store new refresh token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to store new refresh token")
	}

	as.logger.Info("tokens refreshed successfully", "user_id", user.ID)

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	}, nil
}

// Logout revokes all refresh tokens for a user
func (as *AuthService) Logout(ctx context.Context, userID string) error {
	err := as.refreshTokenRepo.RevokeAllUserTokens(ctx, userID)
	if err != nil {
		as.logger.Error("failed to revoke user tokens", "user_id", userID, "error", err)
		return fmt.Errorf("failed to logout user")
	}

	as.logger.Info("user logged out successfully", "user_id", userID)
	return nil
}

// GetUserByID retrieves user information by ID
func (as *AuthService) GetUserByID(ctx context.Context, userID string) (*User, error) {
	user, err := as.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		as.logger.Error("failed to get user by ID", "user_id", userID, "error", err)
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// ValidateAccessToken validates an access token and returns user information
func (as *AuthService) ValidateAccessToken(tokenString string) (*UserClaims, error) {
	return as.jwtService.ValidateAccessToken(tokenString)
}

// Helper methods

func (as *AuthService) validateRegisterRequest(req *RegisterRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (as *AuthService) validateLoginRequest(req *LoginRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (as *AuthService) handleFailedLogin(ctx context.Context, user *User) (*AuthResponse, error) {
	attempts := user.FailedLoginAttempts + 1
	var lockedUntil *time.Time

	// Lock account after 5 failed attempts for 30 minutes
	if attempts >= 5 {
		lockTime := time.Now().Add(30 * time.Minute)
		lockedUntil = &lockTime
	}

	err := as.userRepo.UpdateUserLoginAttempts(ctx, user.ID, attempts, lockedUntil)
	if err != nil {
		as.logger.Error("failed to update login attempts", "user_id", user.ID, "error", err)
	}

	as.logger.Warn("failed login attempt", "user_id", user.ID, "attempts", attempts, "locked", lockedUntil != nil)

	if lockedUntil != nil {
		return nil, fmt.Errorf("account locked due to too many failed attempts. Try again after 30 minutes")
	}

	return nil, fmt.Errorf("invalid credentials")
}

func (as *AuthService) storeRefreshToken(ctx context.Context, userID, token string) error {
	tokenHash := as.jwtService.HashRefreshToken(token)
	refreshToken := &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt: time.Now(),
	}

	return as.refreshTokenRepo.CreateRefreshToken(ctx, refreshToken)
}
