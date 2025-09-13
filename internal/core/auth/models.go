package auth

import (
	"errors"
	"time"
)

// Common errors for auth package
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrAccountLocked      = errors.New("account locked")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrInvalidRole        = errors.New("invalid role")
)

// Valid user roles for registration
const (
	RoleVolunteer = "VOLUNTEER"
	RoleOrganizer = "ORGANIZER"
	RoleAdmin     = "ADMIN"
)

// ValidRegistrationRoles are roles that users can self-assign during registration
var ValidRegistrationRoles = []string{RoleVolunteer, RoleOrganizer}

// User represents a user in the system
type User struct {
	ID                  string     `json:"id" db:"id"`
	Email               string     `json:"email" db:"email"`
	Name                string     `json:"name" db:"name"`
	PasswordHash        *string    `json:"-" db:"password_hash"`
	EmailVerified       bool       `json:"email_verified" db:"email_verified"`
	GoogleID            *string    `json:"google_id" db:"google_id"`
	LastLogin           *time.Time `json:"last_login" db:"last_login"`
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until" db:"locked_until"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// RefreshToken represents a refresh token stored in the database
type RefreshToken struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	TokenHash string     `json:"-" db:"token_hash"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	RevokedAt *time.Time `json:"revoked_at" db:"revoked_at"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Name     string  `json:"name" validate:"required,min=2,max=100"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=8,max=128"`
	Role     *string `json:"role,omitempty" validate:"omitempty,oneof=VOLUNTEER ORGANIZER"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	User         *User  `json:"user"`
}

// IsLocked checks if the user account is currently locked
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// ShouldLockAccount determines if account should be locked based on failed attempts
func (u *User) ShouldLockAccount() bool {
	return u.FailedLoginAttempts >= 5
}

// IsRefreshTokenValid checks if refresh token is still valid
func (rt *RefreshToken) IsValid() bool {
	if rt.RevokedAt != nil {
		return false
	}
	return time.Now().Before(rt.ExpiresAt)
}

// ValidateRole checks if a role is valid for user registration
func ValidateRole(role string) error {
	if role == "" {
		return nil // Role is optional
	}

	for _, validRole := range ValidRegistrationRoles {
		if role == validRole {
			return nil
		}
	}
	return ErrInvalidRole
}

// IsValidRegistrationRole checks if a role is valid for self-registration
func IsValidRegistrationRole(role string) bool {
	return ValidateRole(role) == nil
}
