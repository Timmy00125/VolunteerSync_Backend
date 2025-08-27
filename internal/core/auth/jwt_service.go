package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/jwt"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessTokenType  TokenType = "access"
	RefreshTokenType TokenType = "refresh"
)

// UserClaims represents the custom claims for JWT tokens
type UserClaims struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Roles     []string  `json:"roles"`
	TokenType TokenType `json:"token_type"`
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds until access token expires
}

// JWTService handles JWT token operations
type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
	blocklist     *jwt.Blocklist
}

// JWTConfig represents configuration for JWT service
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

// NewJWTService creates a new JWT service with the provided configuration
func NewJWTService(config JWTConfig) (*JWTService, error) {
	if config.AccessSecret == "" {
		return nil, fmt.Errorf("access secret cannot be empty")
	}
	if config.RefreshSecret == "" {
		return nil, fmt.Errorf("refresh secret cannot be empty")
	}
	if config.AccessExpiry <= 0 {
		config.AccessExpiry = 15 * time.Minute
	}
	if config.RefreshExpiry <= 0 {
		config.RefreshExpiry = 7 * 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "volunteersync"
	}

	// Initialize blocklist for token revocation
	blocklist := jwt.NewBlocklist(1 * time.Hour)

	return &JWTService{
		accessSecret:  []byte(config.AccessSecret),
		refreshSecret: []byte(config.RefreshSecret),
		accessExpiry:  config.AccessExpiry,
		refreshExpiry: config.RefreshExpiry,
		issuer:        config.Issuer,
		blocklist:     blocklist,
	}, nil
}

// GenerateTokenPair generates both access and refresh tokens for a user
func (js *JWTService) GenerateTokenPair(userID, email string, roles []string) (*TokenPair, error) {
	if err := js.validateTokenInputs(userID, email); err != nil {
		return nil, err
	}

	now := time.Now()
	if roles == nil {
		roles = []string{}
	}

	// Generate access token
	accessToken, err := js.generateAccessToken(userID, email, roles, now)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := js.generateRefreshToken(userID, email, roles, now)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
		ExpiresIn:    int64(js.accessExpiry.Seconds()),
	}, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (js *JWTService) ValidateAccessToken(tokenString string) (*UserClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	verifiedToken, err := jwt.Verify(jwt.HS256, js.accessSecret, []byte(tokenString), js.blocklist)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	var claims UserClaims
	err = verifiedToken.Claims(&claims)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token claims: %w", err)
	}

	// Verify token type
	if claims.TokenType != AccessTokenType {
		return nil, fmt.Errorf("invalid token type: expected access token")
	}

	return &claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (js *JWTService) ValidateRefreshToken(tokenString string) (*UserClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	verifiedToken, err := jwt.Verify(jwt.HS256, js.refreshSecret, []byte(tokenString), js.blocklist)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	var claims UserClaims
	err = verifiedToken.Claims(&claims)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token claims: %w", err)
	}

	// Verify token type
	if claims.TokenType != RefreshTokenType {
		return nil, fmt.Errorf("invalid token type: expected refresh token")
	}

	return &claims, nil
}

// RefreshTokens validates a refresh token and generates a new token pair
func (js *JWTService) RefreshTokens(refreshTokenString string) (*TokenPair, error) {
	claims, err := js.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return js.GenerateTokenPair(claims.UserID, claims.Email, claims.Roles)
}

// RevokeToken adds a token to the blocklist to prevent its use
func (js *JWTService) RevokeToken(tokenString string) error {
	if tokenString == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Try to verify token to get its claims (works for both access and refresh tokens)
	var verifiedToken *jwt.VerifiedToken
	var err error

	// Try access token first
	verifiedToken, err = jwt.Verify(jwt.HS256, js.accessSecret, []byte(tokenString))
	if err != nil {
		// Try refresh token
		verifiedToken, err = jwt.Verify(jwt.HS256, js.refreshSecret, []byte(tokenString))
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}
	}

	// Add to blocklist
	js.blocklist.InvalidateToken(verifiedToken.Token, verifiedToken.StandardClaims)
	return nil
}

// HashRefreshToken creates a SHA-256 hash of the refresh token for storage
func (js *JWTService) HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GetTokenClaims extracts claims from any valid token without verification
// Used for testing and debugging purposes - validates signature but ignores expiry
func (js *JWTService) GetTokenClaims(tokenString string) (*UserClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	// Try access token first
	verifiedToken, err := jwt.Verify(jwt.HS256, js.accessSecret, []byte(tokenString))
	if err != nil {
		// Try refresh token
		verifiedToken, err = jwt.Verify(jwt.HS256, js.refreshSecret, []byte(tokenString))
		if err != nil {
			return nil, fmt.Errorf("invalid token: %w", err)
		}
	}

	var claims UserClaims
	err = verifiedToken.Claims(&claims)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token claims: %w", err)
	}

	return &claims, nil
}

// validateTokenInputs validates the required inputs for token generation
func (js *JWTService) validateTokenInputs(userID, email string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	return nil
}

// generateAccessToken creates an access token with the provided claims
func (js *JWTService) generateAccessToken(userID, email string, roles []string, now time.Time) ([]byte, error) {
	accessClaims := UserClaims{
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		TokenType: AccessTokenType,
	}

	standardClaims := jwt.Claims{
		Issuer:   js.issuer,
		Subject:  userID,
		IssuedAt: now.Unix(),
		Expiry:   now.Add(js.accessExpiry).Unix(),
		ID:       uuid.New().String(),
	}

	accessToken, err := jwt.Sign(jwt.HS256, js.accessSecret, accessClaims, standardClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

// generateRefreshToken creates a refresh token with the provided claims
func (js *JWTService) generateRefreshToken(userID, email string, roles []string, now time.Time) ([]byte, error) {
	refreshClaims := UserClaims{
		UserID:    userID,
		Email:     email,
		Roles:     roles,
		TokenType: RefreshTokenType,
	}

	refreshStandardClaims := jwt.Claims{
		Issuer:   js.issuer,
		Subject:  userID,
		IssuedAt: now.Unix(),
		Expiry:   now.Add(js.refreshExpiry).Unix(),
		ID:       uuid.New().String(),
	}

	refreshToken, err := jwt.Sign(jwt.HS256, js.refreshSecret, refreshClaims, refreshStandardClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return refreshToken, nil
}
