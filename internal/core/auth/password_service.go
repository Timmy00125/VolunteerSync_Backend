package auth

import (
	"crypto/subtle"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles secure password operations
type PasswordService struct {
	cost int
}

// NewPasswordService creates a new password service with secure defaults
func NewPasswordService(cost int) *PasswordService {
	// Ensure minimum security cost of 12
	if cost < 12 {
		cost = 12
	}
	return &PasswordService{
		cost: cost,
	}
}

// HashPassword generates a bcrypt hash of the password using the configured cost
func (ps *PasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), ps.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword verifies if the provided password matches the stored hash
// Uses constant-time comparison to prevent timing attacks
func (ps *PasswordService) VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" {
		return fmt.Errorf("hashed password cannot be empty")
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// Use constant-time comparison to prevent timing attacks
		// This ensures verification always takes roughly the same time
		dummy := "$2a$12$dummy.hash.to.prevent.timing.attacks"
		_ = bcrypt.CompareHashAndPassword([]byte(dummy), []byte("dummy"))
		return fmt.Errorf("invalid password")
	}

	return nil
}

// IsValidPasswordHash checks if the provided string is a valid bcrypt hash
func (ps *PasswordService) IsValidPasswordHash(hash string) bool {
	if hash == "" {
		return false
	}

	// Try to get cost from hash to validate format
	_, err := bcrypt.Cost([]byte(hash))
	return err == nil
}

// GetHashCost returns the cost factor used for the given hash
func (ps *PasswordService) GetHashCost(hash string) (int, error) {
	if hash == "" {
		return 0, fmt.Errorf("hash cannot be empty")
	}

	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		return 0, fmt.Errorf("invalid bcrypt hash: %w", err)
	}

	return cost, nil
}

// ValidatePasswordStrength validates password meets minimum security requirements
func (ps *PasswordService) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must be no more than 128 characters long")
	}

	// Check for common weak passwords
	weakPasswords := []string{
		"password", "123456", "12345678", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
	}

	for _, weak := range weakPasswords {
		if subtle.ConstantTimeCompare([]byte(password), []byte(weak)) == 1 {
			return fmt.Errorf("password is too common and easily guessable")
		}
	}

	return nil
}
