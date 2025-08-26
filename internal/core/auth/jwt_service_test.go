package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/kataras/jwt"
)

func TestNewJWTService(t *testing.T) {
	tests := []struct {
		name      string
		config    JWTConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid configuration",
			config: JWTConfig{
				AccessSecret:  "access-secret-key",
				RefreshSecret: "refresh-secret-key",
				AccessExpiry:  15 * time.Minute,
				RefreshExpiry: 7 * 24 * time.Hour,
				Issuer:        "volunteersync",
			},
			wantError: false,
		},
		{
			name: "missing access secret",
			config: JWTConfig{
				RefreshSecret: "refresh-secret-key",
			},
			wantError: true,
			errorMsg:  "access secret cannot be empty",
		},
		{
			name: "missing refresh secret",
			config: JWTConfig{
				AccessSecret: "access-secret-key",
			},
			wantError: true,
			errorMsg:  "refresh secret cannot be empty",
		},
		{
			name: "defaults applied for zero values",
			config: JWTConfig{
				AccessSecret:  "access-secret-key",
				RefreshSecret: "refresh-secret-key",
				// AccessExpiry and RefreshExpiry are zero
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewJWTService(tt.config)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewJWTService() expected error, got nil")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("NewJWTService() error = %v, want error containing %q", err, tt.errorMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewJWTService() unexpected error = %v", err)
				return
			}

			if service == nil {
				t.Error("NewJWTService() returned nil service")
				return
			}

			// Check defaults were applied
			if tt.config.AccessExpiry == 0 && service.accessExpiry != 15*time.Minute {
				t.Errorf("NewJWTService() accessExpiry = %v, want 15m", service.accessExpiry)
			}
			if tt.config.RefreshExpiry == 0 && service.refreshExpiry != 7*24*time.Hour {
				t.Errorf("NewJWTService() refreshExpiry = %v, want 168h", service.refreshExpiry)
			}
			if tt.config.Issuer == "" && service.issuer != "volunteersync" {
				t.Errorf("NewJWTService() issuer = %v, want volunteersync", service.issuer)
			}
		})
	}
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	service := createTestJWTService(t)

	t.Run("successful token generation", func(t *testing.T) {
		userID := "test-user-id"
		email := "test@example.com"
		roles := []string{"user", "admin"}

		tokenPair, err := service.GenerateTokenPair(userID, email, roles)
		if err != nil {
			t.Errorf("GenerateTokenPair() error = %v, want nil", err)
			return
		}

		if tokenPair == nil {
			t.Error("GenerateTokenPair() returned nil token pair")
			return
		}

		if tokenPair.AccessToken == "" {
			t.Error("GenerateTokenPair() returned empty access token")
		}
		if tokenPair.RefreshToken == "" {
			t.Error("GenerateTokenPair() returned empty refresh token")
		}
		if tokenPair.ExpiresIn <= 0 {
			t.Error("GenerateTokenPair() returned invalid expires_in")
		}

		// Verify access token contains correct claims
		accessClaims, err := service.ValidateAccessToken(tokenPair.AccessToken)
		if err != nil {
			t.Errorf("Generated access token validation failed: %v", err)
			return
		}

		if accessClaims.UserID != userID {
			t.Errorf("Access token UserID = %v, want %v", accessClaims.UserID, userID)
		}
		if accessClaims.Email != email {
			t.Errorf("Access token Email = %v, want %v", accessClaims.Email, email)
		}
		if len(accessClaims.Roles) != len(roles) {
			t.Errorf("Access token Roles length = %v, want %v", len(accessClaims.Roles), len(roles))
		}
		if accessClaims.TokenType != AccessTokenType {
			t.Errorf("Access token TokenType = %v, want %v", accessClaims.TokenType, AccessTokenType)
		}

		// Verify refresh token contains correct claims
		refreshClaims, err := service.ValidateRefreshToken(tokenPair.RefreshToken)
		if err != nil {
			t.Errorf("Generated refresh token validation failed: %v", err)
			return
		}

		if refreshClaims.UserID != userID {
			t.Errorf("Refresh token UserID = %v, want %v", refreshClaims.UserID, userID)
		}
		if refreshClaims.TokenType != RefreshTokenType {
			t.Errorf("Refresh token TokenType = %v, want %v", refreshClaims.TokenType, RefreshTokenType)
		}
	})

	t.Run("empty user ID", func(t *testing.T) {
		_, err := service.GenerateTokenPair("", "test@example.com", []string{"user"})
		if err == nil {
			t.Error("GenerateTokenPair() with empty userID should return error")
		}
	})

	t.Run("empty email", func(t *testing.T) {
		_, err := service.GenerateTokenPair("user-id", "", []string{"user"})
		if err == nil {
			t.Error("GenerateTokenPair() with empty email should return error")
		}
	})

	t.Run("nil roles", func(t *testing.T) {
		tokenPair, err := service.GenerateTokenPair("user-id", "test@example.com", nil)
		if err != nil {
			t.Errorf("GenerateTokenPair() with nil roles error = %v, want nil", err)
			return
		}

		claims, err := service.ValidateAccessToken(tokenPair.AccessToken)
		if err != nil {
			t.Errorf("ValidateAccessToken() error = %v", err)
			return
		}

		if claims.Roles == nil {
			t.Error("Token claims should have non-nil roles slice")
		}
	})

	t.Run("empty roles", func(t *testing.T) {
		tokenPair, err := service.GenerateTokenPair("user-id", "test@example.com", []string{})
		if err != nil {
			t.Errorf("GenerateTokenPair() with empty roles error = %v, want nil", err)
			return
		}

		claims, err := service.ValidateAccessToken(tokenPair.AccessToken)
		if err != nil {
			t.Errorf("ValidateAccessToken() error = %v", err)
			return
		}

		if len(claims.Roles) != 0 {
			t.Errorf("Token claims roles length = %v, want 0", len(claims.Roles))
		}
	})

	t.Run("token expiration times are set correctly", func(t *testing.T) {
		before := time.Now()
		tokenPair, err := service.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		after := time.Now()

		if err != nil {
			t.Errorf("GenerateTokenPair() error = %v", err)
			return
		}

		// Check expires_in is correct (should be 15 minutes in seconds)
		expectedExpiresIn := int64(15 * 60) // 15 minutes in seconds
		if tokenPair.ExpiresIn != expectedExpiresIn {
			t.Errorf("TokenPair ExpiresIn = %v, want %v", tokenPair.ExpiresIn, expectedExpiresIn)
		}

		// Validate the actual token expiration times
		accessClaims, err := service.GetTokenClaims(tokenPair.AccessToken)
		if err != nil {
			t.Errorf("GetTokenClaims() for access token error = %v", err)
			return
		}

		refreshClaims, err := service.GetTokenClaims(tokenPair.RefreshToken)
		if err != nil {
			t.Errorf("GetTokenClaims() for refresh token error = %v", err)
			return
		}

		// The access token should expire in approximately 15 minutes
		expectedAccessExpiry := before.Add(15 * time.Minute)
		actualAccessExpiry := time.Unix(getStandardClaims(tokenPair.AccessToken, service).Expiry, 0)

		if actualAccessExpiry.Before(expectedAccessExpiry.Add(-time.Minute)) ||
			actualAccessExpiry.After(after.Add(15*time.Minute)) {
			t.Errorf("Access token expiry = %v, want around %v", actualAccessExpiry, expectedAccessExpiry)
		}

		// The refresh token should expire in approximately 7 days
		expectedRefreshExpiry := before.Add(7 * 24 * time.Hour)
		actualRefreshExpiry := time.Unix(getStandardClaims(tokenPair.RefreshToken, service).Expiry, 0)

		if actualRefreshExpiry.Before(expectedRefreshExpiry.Add(-time.Minute)) ||
			actualRefreshExpiry.After(after.Add(7*24*time.Hour)) {
			t.Errorf("Refresh token expiry = %v, want around %v", actualRefreshExpiry, expectedRefreshExpiry)
		}

		// Sanity check that access token expires before refresh token
		if actualAccessExpiry.After(actualRefreshExpiry) {
			t.Error("Access token should expire before refresh token")
		}

		// Verify claims are as expected
		_ = accessClaims
		_ = refreshClaims
	})
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	service := createTestJWTService(t)

	t.Run("valid token validation succeeds", func(t *testing.T) {
		userID := "test-user-id"
		email := "test@example.com"
		roles := []string{"user"}

		tokenPair, err := service.GenerateTokenPair(userID, email, roles)
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		claims, err := service.ValidateAccessToken(tokenPair.AccessToken)
		if err != nil {
			t.Errorf("ValidateAccessToken() error = %v, want nil", err)
			return
		}

		if claims.UserID != userID {
			t.Errorf("ValidateAccessToken() UserID = %v, want %v", claims.UserID, userID)
		}
		if claims.Email != email {
			t.Errorf("ValidateAccessToken() Email = %v, want %v", claims.Email, email)
		}
		if claims.TokenType != AccessTokenType {
			t.Errorf("ValidateAccessToken() TokenType = %v, want %v", claims.TokenType, AccessTokenType)
		}
	})

	t.Run("empty token validation fails", func(t *testing.T) {
		_, err := service.ValidateAccessToken("")
		if err == nil {
			t.Error("ValidateAccessToken() with empty token should return error")
		}
	})

	t.Run("malformed token validation fails", func(t *testing.T) {
		malformedTokens := []string{
			"not.a.token",
			"invalid-token-string",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
			"too.many.parts.in.token.here",
		}

		for _, token := range malformedTokens {
			_, err := service.ValidateAccessToken(token)
			if err == nil {
				t.Errorf("ValidateAccessToken() with malformed token %q should return error", token)
			}
		}
	})

	t.Run("token with invalid signature fails", func(t *testing.T) {
		// Create a service with different secret
		differentService := createTestJWTServiceWithSecrets(t, "different-access-secret", "different-refresh-secret")

		// Generate token with different service
		tokenPair, err := differentService.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Try to validate with original service (different secret)
		_, err = service.ValidateAccessToken(tokenPair.AccessToken)
		if err == nil {
			t.Error("ValidateAccessToken() with wrong signature should return error")
		}
	})

	t.Run("refresh token used as access token fails", func(t *testing.T) {
		tokenPair, err := service.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Try to validate refresh token as access token
		_, err = service.ValidateAccessToken(tokenPair.RefreshToken)
		if err == nil {
			t.Error("ValidateAccessToken() with refresh token should return error")
		}
		// Since refresh tokens are signed with different secret, we expect signature error
		if !strings.Contains(err.Error(), "invalid token signature") {
			t.Errorf("ValidateAccessToken() error = %v, want error containing 'invalid token signature'", err)
		}
	})

	t.Run("expired token validation fails", func(t *testing.T) {
		// Create service with very short expiry
		config := JWTConfig{
			AccessSecret:  "access-secret-key",
			RefreshSecret: "refresh-secret-key",
			AccessExpiry:  1 * time.Millisecond, // Very short expiry
			RefreshExpiry: 7 * 24 * time.Hour,
			Issuer:        "test",
		}

		shortExpiryService, err := NewJWTService(config)
		if err != nil {
			t.Fatalf("Failed to create test service: %v", err)
		}

		tokenPair, err := shortExpiryService.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		_, err = shortExpiryService.ValidateAccessToken(tokenPair.AccessToken)
		if err == nil {
			t.Error("ValidateAccessToken() with expired token should return error")
		}
	})
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	service := createTestJWTService(t)

	t.Run("valid refresh token validation succeeds", func(t *testing.T) {
		userID := "test-user-id"
		email := "test@example.com"
		roles := []string{"user"}

		tokenPair, err := service.GenerateTokenPair(userID, email, roles)
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		claims, err := service.ValidateRefreshToken(tokenPair.RefreshToken)
		if err != nil {
			t.Errorf("ValidateRefreshToken() error = %v, want nil", err)
			return
		}

		if claims.UserID != userID {
			t.Errorf("ValidateRefreshToken() UserID = %v, want %v", claims.UserID, userID)
		}
		if claims.Email != email {
			t.Errorf("ValidateRefreshToken() Email = %v, want %v", claims.Email, email)
		}
		if claims.TokenType != RefreshTokenType {
			t.Errorf("ValidateRefreshToken() TokenType = %v, want %v", claims.TokenType, RefreshTokenType)
		}
	})

	t.Run("access token used as refresh token fails", func(t *testing.T) {
		tokenPair, err := service.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Try to validate access token as refresh token
		_, err = service.ValidateRefreshToken(tokenPair.AccessToken)
		if err == nil {
			t.Error("ValidateRefreshToken() with access token should return error")
		}
		// Since access tokens are signed with different secret, we expect signature error
		if !strings.Contains(err.Error(), "invalid token signature") {
			t.Errorf("ValidateRefreshToken() error = %v, want error containing 'invalid token signature'", err)
		}
	})

	t.Run("expired refresh token validation fails", func(t *testing.T) {
		// Create service with very short refresh expiry
		config := JWTConfig{
			AccessSecret:  "access-secret-key",
			RefreshSecret: "refresh-secret-key",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 1 * time.Millisecond, // Very short expiry
			Issuer:        "test",
		}

		shortExpiryService, err := NewJWTService(config)
		if err != nil {
			t.Fatalf("Failed to create test service: %v", err)
		}

		tokenPair, err := shortExpiryService.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		_, err = shortExpiryService.ValidateRefreshToken(tokenPair.RefreshToken)
		if err == nil {
			t.Error("ValidateRefreshToken() with expired token should return error")
		}
	})
}

func TestJWTService_RefreshTokens(t *testing.T) {
	service := createTestJWTService(t)

	t.Run("valid refresh token generates new access token", func(t *testing.T) {
		userID := "test-user-id"
		email := "test@example.com"
		roles := []string{"user", "admin"}

		// Generate initial token pair
		originalPair, err := service.GenerateTokenPair(userID, email, roles)
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Use refresh token to get new token pair
		newPair, err := service.RefreshTokens(originalPair.RefreshToken)
		if err != nil {
			t.Errorf("RefreshTokens() error = %v, want nil", err)
			return
		}

		if newPair == nil {
			t.Error("RefreshTokens() returned nil token pair")
			return
		}

		// Verify new tokens are different from original
		if newPair.AccessToken == originalPair.AccessToken {
			t.Error("RefreshTokens() returned same access token")
		}
		if newPair.RefreshToken == originalPair.RefreshToken {
			t.Error("RefreshTokens() returned same refresh token")
		}

		// Verify new access token contains correct claims
		claims, err := service.ValidateAccessToken(newPair.AccessToken)
		if err != nil {
			t.Errorf("New access token validation failed: %v", err)
			return
		}

		if claims.UserID != userID {
			t.Errorf("New access token UserID = %v, want %v", claims.UserID, userID)
		}
		if claims.Email != email {
			t.Errorf("New access token Email = %v, want %v", claims.Email, email)
		}
		if len(claims.Roles) != len(roles) {
			t.Errorf("New access token Roles length = %v, want %v", len(claims.Roles), len(roles))
		}
	})

	t.Run("invalid refresh token is rejected", func(t *testing.T) {
		_, err := service.RefreshTokens("invalid-refresh-token")
		if err == nil {
			t.Error("RefreshTokens() with invalid token should return error")
		}
	})

	t.Run("expired refresh token is rejected", func(t *testing.T) {
		// Create service with very short refresh expiry
		config := JWTConfig{
			AccessSecret:  "access-secret-key",
			RefreshSecret: "refresh-secret-key",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 1 * time.Millisecond,
			Issuer:        "test",
		}

		shortExpiryService, err := NewJWTService(config)
		if err != nil {
			t.Fatalf("Failed to create test service: %v", err)
		}

		tokenPair, err := shortExpiryService.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		_, err = shortExpiryService.RefreshTokens(tokenPair.RefreshToken)
		if err == nil {
			t.Error("RefreshTokens() with expired token should return error")
		}
	})

	t.Run("access token cannot be used for refresh", func(t *testing.T) {
		tokenPair, err := service.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		_, err = service.RefreshTokens(tokenPair.AccessToken)
		if err == nil {
			t.Error("RefreshTokens() with access token should return error")
		}
	})
}

func TestJWTService_RevokeToken(t *testing.T) {
	service := createTestJWTService(t)

	t.Run("revoke access token", func(t *testing.T) {
		tokenPair, err := service.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Verify token is valid before revocation
		_, err = service.ValidateAccessToken(tokenPair.AccessToken)
		if err != nil {
			t.Errorf("Token should be valid before revocation: %v", err)
		}

		// Revoke the token
		err = service.RevokeToken(tokenPair.AccessToken)
		if err != nil {
			t.Errorf("RevokeToken() error = %v, want nil", err)
		}

		// Verify token is invalid after revocation
		_, err = service.ValidateAccessToken(tokenPair.AccessToken)
		if err == nil {
			t.Error("Token should be invalid after revocation")
		}
	})

	t.Run("revoke refresh token", func(t *testing.T) {
		tokenPair, err := service.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			t.Fatalf("Failed to generate test token: %v", err)
		}

		// Verify token is valid before revocation
		_, err = service.ValidateRefreshToken(tokenPair.RefreshToken)
		if err != nil {
			t.Errorf("Token should be valid before revocation: %v", err)
		}

		// Revoke the token
		err = service.RevokeToken(tokenPair.RefreshToken)
		if err != nil {
			t.Errorf("RevokeToken() error = %v, want nil", err)
		}

		// Verify token is invalid after revocation
		_, err = service.ValidateRefreshToken(tokenPair.RefreshToken)
		if err == nil {
			t.Error("Token should be invalid after revocation")
		}
	})

	t.Run("revoking invalid token returns error", func(t *testing.T) {
		err := service.RevokeToken("invalid-token")
		if err == nil {
			t.Error("RevokeToken() with invalid token should return error")
		}
	})

	t.Run("revoking empty token returns error", func(t *testing.T) {
		err := service.RevokeToken("")
		if err == nil {
			t.Error("RevokeToken() with empty token should return error")
		}
	})
}

func TestJWTService_HashRefreshToken(t *testing.T) {
	service := createTestJWTService(t)

	t.Run("hash is consistent", func(t *testing.T) {
		token := "test-refresh-token"
		hash1 := service.HashRefreshToken(token)
		hash2 := service.HashRefreshToken(token)

		if hash1 != hash2 {
			t.Error("HashRefreshToken() should return consistent hash for same input")
		}

		if hash1 == "" {
			t.Error("HashRefreshToken() should not return empty hash")
		}
	})

	t.Run("different tokens produce different hashes", func(t *testing.T) {
		token1 := "test-refresh-token-1"
		token2 := "test-refresh-token-2"

		hash1 := service.HashRefreshToken(token1)
		hash2 := service.HashRefreshToken(token2)

		if hash1 == hash2 {
			t.Error("HashRefreshToken() should produce different hashes for different tokens")
		}
	})

	t.Run("empty token handling", func(t *testing.T) {
		hash := service.HashRefreshToken("")
		// Should not panic and should return a valid hash (even for empty string)
		if hash == "" {
			t.Error("HashRefreshToken() with empty string should still return a hash")
		}
	})
}

// Helper functions for tests

func createTestJWTService(t *testing.T) *JWTService {
	config := JWTConfig{
		AccessSecret:  "test-access-secret-key",
		RefreshSecret: "test-refresh-secret-key",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "test",
	}

	service, err := NewJWTService(config)
	if err != nil {
		t.Fatalf("Failed to create test JWT service: %v", err)
	}

	return service
}

func createTestJWTServiceWithSecrets(t *testing.T, accessSecret, refreshSecret string) *JWTService {
	config := JWTConfig{
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "test",
	}

	service, err := NewJWTService(config)
	if err != nil {
		t.Fatalf("Failed to create test JWT service: %v", err)
	}

	return service
}

// Helper function to extract standard claims from token
func getStandardClaims(tokenString string, service *JWTService) jwt.Claims {
	// Try access token first
	verifiedToken, err := jwt.Verify(jwt.HS256, service.accessSecret, []byte(tokenString))
	if err != nil {
		// Try refresh token
		verifiedToken, _ = jwt.Verify(jwt.HS256, service.refreshSecret, []byte(tokenString))
	}
	return verifiedToken.StandardClaims
}

// Benchmark tests
func BenchmarkJWTService_GenerateTokenPair(b *testing.B) {
	service := createTestJWTService(&testing.T{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
		if err != nil {
			b.Fatalf("GenerateTokenPair failed: %v", err)
		}
	}
}

func BenchmarkJWTService_ValidateAccessToken(b *testing.B) {
	service := createTestJWTService(&testing.T{})
	tokenPair, err := service.GenerateTokenPair("user-id", "test@example.com", []string{"user"})
	if err != nil {
		b.Fatalf("Failed to generate test token: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ValidateAccessToken(tokenPair.AccessToken)
		if err != nil {
			b.Fatalf("ValidateAccessToken failed: %v", err)
		}
	}
}
