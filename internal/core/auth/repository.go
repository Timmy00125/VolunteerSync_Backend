package auth

import (
	"context"
	"time"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// CreateUser creates a new user in the database
	CreateUser(ctx context.Context, user *User) error

	// GetUserByID retrieves a user by their ID
	GetUserByID(ctx context.Context, id string) (*User, error)

	// GetUserByEmail retrieves a user by their email address
	GetUserByEmail(ctx context.Context, email string) (*User, error)

	// GetUserByGoogleID retrieves a user by their Google OAuth ID
	GetUserByGoogleID(ctx context.Context, googleID string) (*User, error)

	// UpdateUser updates an existing user's information
	UpdateUser(ctx context.Context, user *User) error

	// UpdateUserLoginAttempts updates failed login attempts and potential lockout
	UpdateUserLoginAttempts(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error

	// UpdateLastLogin updates the user's last login timestamp
	UpdateLastLogin(ctx context.Context, userID string) error

	// EmailExists checks if an email is already registered
	EmailExists(ctx context.Context, email string) (bool, error)
}

// RefreshTokenRepository defines the interface for refresh token operations
type RefreshTokenRepository interface {
	// CreateRefreshToken stores a new refresh token
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error

	// GetRefreshToken retrieves a refresh token by its hash
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)

	// RevokeRefreshToken marks a refresh token as revoked
	RevokeRefreshToken(ctx context.Context, tokenHash string) error

	// RevokeAllUserTokens revokes all refresh tokens for a user
	RevokeAllUserTokens(ctx context.Context, userID string) error

	// DeleteExpiredTokens removes expired tokens from storage
	DeleteExpiredTokens(ctx context.Context) error

	// CountActiveTokensForUser counts active refresh tokens for a user
	CountActiveTokensForUser(ctx context.Context, userID string) (int, error)
}
