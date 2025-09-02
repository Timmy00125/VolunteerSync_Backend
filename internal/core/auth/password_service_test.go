package auth

import (
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestNewPasswordService(t *testing.T) {
	tests := []struct {
		name         string
		cost         int
		expectedCost int
	}{
		{
			name:         "valid cost",
			cost:         14,
			expectedCost: 14,
		},
		{
			name:         "low cost gets upgraded to minimum",
			cost:         8,
			expectedCost: 12,
		},
		{
			name:         "zero cost gets upgraded to minimum",
			cost:         0,
			expectedCost: 12,
		},
		{
			name:         "negative cost gets upgraded to minimum",
			cost:         -1,
			expectedCost: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPasswordService(tt.cost)
			if ps.cost != tt.expectedCost {
				t.Errorf("NewPasswordService() cost = %v, want %v", ps.cost, tt.expectedCost)
			}
		})
	}
}

func TestPasswordService_HashPassword(t *testing.T) {
	ps := NewPasswordService(12)

	t.Run("successful password hashing", func(t *testing.T) {
		password := "SecurePassword123!"

		hash, err := ps.HashPassword(password)
		if err != nil {
			t.Errorf("HashPassword() error = %v, want nil", err)
			return
		}

		if hash == "" {
			t.Error("HashPassword() returned empty hash")
		}

		// Verify the hash can be used to verify the password
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err != nil {
			t.Errorf("Generated hash cannot verify original password: %v", err)
		}
	})

	t.Run("hash is different each time (salt randomization)", func(t *testing.T) {
		password := "TestPassword123"

		hash1, err1 := ps.HashPassword(password)
		hash2, err2 := ps.HashPassword(password)

		if err1 != nil || err2 != nil {
			t.Errorf("HashPassword() errors: %v, %v", err1, err2)
			return
		}

		if hash1 == hash2 {
			t.Error("HashPassword() returned identical hashes for same password (no salt randomization)")
		}
	})

	t.Run("hash cost factor is appropriate (>= 12)", func(t *testing.T) {
		password := "TestPassword123"

		hash, err := ps.HashPassword(password)
		if err != nil {
			t.Errorf("HashPassword() error = %v", err)
			return
		}

		cost, err := bcrypt.Cost([]byte(hash))
		if err != nil {
			t.Errorf("Failed to extract cost from hash: %v", err)
			return
		}

		if cost < 12 {
			t.Errorf("HashPassword() cost = %v, want >= 12", cost)
		}
	})

	t.Run("empty password handling", func(t *testing.T) {
		_, err := ps.HashPassword("")
		if err == nil {
			t.Error("HashPassword() with empty password should return error")
		}

		expectedMsg := "password cannot be empty"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("HashPassword() error = %v, want error containing %q", err, expectedMsg)
		}
	})

	t.Run("long password handling", func(t *testing.T) {
		// bcrypt has a limit of 72 bytes
		longPassword := strings.Repeat("a", 80)

		_, err := ps.HashPassword(longPassword)
		if err == nil {
			t.Error("HashPassword() with long password should return error")
			return
		}

		expectedMsg := "password length exceeds 72 bytes"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("HashPassword() error = %v, want error containing %q", err, expectedMsg)
		}
	})

	t.Run("unicode password handling", func(t *testing.T) {
		password := "密码123ñáéí🔒"

		hash, err := ps.HashPassword(password)
		if err != nil {
			t.Errorf("HashPassword() with unicode password error = %v", err)
			return
		}

		if hash == "" {
			t.Error("HashPassword() returned empty hash for unicode password")
		}

		// Verify the hash works with the original password
		err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		if err != nil {
			t.Errorf("Generated hash cannot verify unicode password: %v", err)
		}
	})
}

func TestPasswordService_VerifyPassword(t *testing.T) {
	ps := NewPasswordService(12)

	// Generate a known hash for testing
	password := "TestPassword123!"
	hash, err := ps.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to generate test hash: %v", err)
	}

	t.Run("correct password verification returns true", func(t *testing.T) {
		err := ps.VerifyPassword(hash, password)
		if err != nil {
			t.Errorf("VerifyPassword() with correct password error = %v, want nil", err)
		}
	})

	t.Run("incorrect password verification returns false", func(t *testing.T) {
		wrongPassword := "WrongPassword123!"

		err := ps.VerifyPassword(hash, wrongPassword)
		if err == nil {
			t.Error("VerifyPassword() with incorrect password should return error")
		}
	})

	t.Run("verification against invalid hash fails gracefully", func(t *testing.T) {
		invalidHash := "invalid-hash-string"

		err := ps.VerifyPassword(invalidHash, password)
		if err == nil {
			t.Error("VerifyPassword() with invalid hash should return error")
		}
	})

	t.Run("empty hashed password handling", func(t *testing.T) {
		err := ps.VerifyPassword("", password)
		if err == nil {
			t.Error("VerifyPassword() with empty hash should return error")
		}

		expectedMsg := "hashed password cannot be empty"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("VerifyPassword() error = %v, want error containing %q", err, expectedMsg)
		}
	})

	t.Run("empty password handling", func(t *testing.T) {
		err := ps.VerifyPassword(hash, "")
		if err == nil {
			t.Error("VerifyPassword() with empty password should return error")
		}

		expectedMsg := "password cannot be empty"
		if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("VerifyPassword() error = %v, want error containing %q", err, expectedMsg)
		}
	})

	t.Run("timing attack resistance", func(t *testing.T) {
		// This test ensures that verification takes approximately the same time
		// regardless of whether the password is correct or incorrect
		numTests := 50
		correctTimes := make([]time.Duration, numTests)
		incorrectTimes := make([]time.Duration, numTests)

		for i := 0; i < numTests; i++ {
			// Test correct password timing
			start := time.Now()
			ps.VerifyPassword(hash, password)
			correctTimes[i] = time.Since(start)

			// Test incorrect password timing
			start = time.Now()
			ps.VerifyPassword(hash, "wrongpassword")
			incorrectTimes[i] = time.Since(start)
		}

		// Calculate average times
		var correctTotal, incorrectTotal time.Duration
		for i := 0; i < numTests; i++ {
			correctTotal += correctTimes[i]
			incorrectTotal += incorrectTimes[i]
		}

		correctAvg := correctTotal / time.Duration(numTests)
		incorrectAvg := incorrectTotal / time.Duration(numTests)

		// The times should be within 50% of each other (bcrypt naturally provides timing resistance)
		diff := float64(correctAvg - incorrectAvg)
		if diff < 0 {
			diff = -diff
		}

		avgTime := float64(correctAvg+incorrectAvg) / 2
		if avgTime == 0 {
			t.Error("Average verification time is zero, timing test invalid")
			return
		}

		percentDiff := (diff / avgTime) * 100

		// Allow up to 50% difference due to natural variation
		if percentDiff > 50 {
			t.Errorf("Timing difference too large: correct=%v, incorrect=%v, diff=%.1f%%",
				correctAvg, incorrectAvg, percentDiff)
		}

		// Log timing info for manual review
		t.Logf("Timing analysis: correct=%v, incorrect=%v, diff=%.1f%%",
			correctAvg, incorrectAvg, percentDiff)
	})

	t.Run("case sensitivity", func(t *testing.T) {
		// Passwords should be case sensitive
		upperPassword := strings.ToUpper(password)

		err := ps.VerifyPassword(hash, upperPassword)
		if err == nil {
			t.Error("VerifyPassword() should be case sensitive")
		}
	})

	t.Run("null byte handling", func(t *testing.T) {
		// Test that null bytes in password don't cause issues
		passwordWithNull := "password\x00"
		hashWithNull, err := ps.HashPassword(passwordWithNull)
		if err != nil {
			t.Errorf("Failed to hash password with null byte: %v", err)
			return
		}

		// Verify with exact same password
		err = ps.VerifyPassword(hashWithNull, passwordWithNull)
		if err != nil {
			t.Errorf("VerifyPassword() with null byte password error = %v", err)
		}

		// Verify that password without null byte fails
		err = ps.VerifyPassword(hashWithNull, "password")
		if err == nil {
			t.Error("VerifyPassword() should distinguish passwords with and without null bytes")
		}
	})
}

func TestPasswordService_ValidatePasswordStrength(t *testing.T) {
	ps := NewPasswordService(12)

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "strong password",
			password: "StrongPassword123!",
			wantErr:  false,
		},
		{
			name:     "password with mixed case, numbers, and symbols",
			password: "MySecure@Pass123",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "ThisIsAVeryLongPasswordThatShouldBeSecure123!@#",
			wantErr:  false,
		},
		{
			name:     "password too short",
			password: "short",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "password too long",
			password: strings.Repeat("a", 129),
			wantErr:  true,
			errMsg:   "no more than 128 characters",
		},
		{
			name:     "password without uppercase",
			password: "lowercase123!",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "password without lowercase",
			password: "UPPERCASE123!",
			wantErr:  true,
			errMsg:   "lowercase letter",
		},
		{
			name:     "password without numbers",
			password: "NoNumbers!",
			wantErr:  true,
			errMsg:   "number",
		},
		{
			name:     "password without special characters",
			password: "NoSpecialChars123",
			wantErr:  true,
			errMsg:   "special character",
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
			errMsg:   "required",
		},
		{
			name:     "whitespace only password",
			password: "   ",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ps.ValidatePasswordStrength(tt.password)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePasswordStrength() error = nil, want error")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePasswordStrength() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePasswordStrength() error = %v, want nil", err)
				}
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkPasswordService_HashPassword(b *testing.B) {
	ps := NewPasswordService(12)
	password := "TestPassword123!"

	for b.Loop() {
		_, err := ps.HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword failed: %v", err)
		}
	}
}

func BenchmarkPasswordService_VerifyPassword(b *testing.B) {
	ps := NewPasswordService(12)
	password := "TestPassword123!"
	hash, err := ps.HashPassword(password)
	if err != nil {
		b.Fatalf("Failed to generate test hash: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		err := ps.VerifyPassword(hash, password)
		if err != nil {
			b.Fatalf("VerifyPassword failed: %v", err)
		}
	}
}
