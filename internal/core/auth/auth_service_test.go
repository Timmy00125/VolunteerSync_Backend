package auth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

// Mock implementations for testing

type MockUserRepository struct {
	users          map[string]*User
	emailToUserID  map[string]string
	googleToUserID map[string]string
	shouldError    bool
	errorMsg       string
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:          make(map[string]*User),
		emailToUserID:  make(map[string]string),
		googleToUserID: make(map[string]string),
	}
}

func (m *MockUserRepository) SetError(shouldError bool, msg string) {
	m.shouldError = shouldError
	m.errorMsg = msg
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *User) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	// Check if email already exists
	if _, exists := m.emailToUserID[user.Email]; exists {
		return errors.New("email already exists")
	}

	m.users[user.ID] = user
	m.emailToUserID[user.Email] = user.ID
	if user.GoogleID != nil {
		m.googleToUserID[*user.GoogleID] = user.ID
	}
	return nil
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	user, exists := m.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	userID, exists := m.emailToUserID[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return m.users[userID], nil
}

func (m *MockUserRepository) GetUserByGoogleID(ctx context.Context, googleID string) (*User, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	userID, exists := m.googleToUserID[googleID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return m.users[userID], nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *User) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	if _, exists := m.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdateUserLoginAttempts(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	user, exists := m.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	user.FailedLoginAttempts = attempts
	user.LockedUntil = lockedUntil
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	user, exists := m.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	now := time.Now()
	user.LastLogin = &now
	return nil
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	if m.shouldError {
		return false, errors.New(m.errorMsg)
	}

	_, exists := m.emailToUserID[email]
	return exists, nil
}

type MockRefreshTokenRepository struct {
	tokens      map[string]*RefreshToken
	shouldError bool
	errorMsg    string
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		tokens: make(map[string]*RefreshToken),
	}
}

func (m *MockRefreshTokenRepository) SetError(shouldError bool, msg string) {
	m.shouldError = shouldError
	m.errorMsg = msg
}

func (m *MockRefreshTokenRepository) CreateRefreshToken(ctx context.Context, token *RefreshToken) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	m.tokens[token.TokenHash] = token
	return nil
}

func (m *MockRefreshTokenRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	token, exists := m.tokens[tokenHash]
	if !exists {
		return nil, errors.New("token not found")
	}
	return token, nil
}

func (m *MockRefreshTokenRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	token, exists := m.tokens[tokenHash]
	if !exists {
		return errors.New("token not found")
	}

	now := time.Now()
	token.RevokedAt = &now
	return nil
}

func (m *MockRefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	now := time.Now()
	for _, token := range m.tokens {
		if token.UserID == userID {
			token.RevokedAt = &now
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	now := time.Now()
	for hash, token := range m.tokens {
		if token.ExpiresAt.Before(now) {
			delete(m.tokens, hash)
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) CountActiveTokensForUser(ctx context.Context, userID string) (int, error) {
	if m.shouldError {
		return 0, errors.New(m.errorMsg)
	}

	count := 0
	now := time.Now()
	for _, token := range m.tokens {
		if token.UserID == userID && token.RevokedAt == nil && token.ExpiresAt.After(now) {
			count++
		}
	}
	return count, nil
}

// MockUserRoleService for testing role assignment
type MockUserRoleService struct {
	userRoles   map[string][]string
	shouldError bool
	errorMsg    string
}

func NewMockUserRoleService() *MockUserRoleService {
	return &MockUserRoleService{
		userRoles: make(map[string][]string),
	}
}

func (m *MockUserRoleService) SetError(shouldError bool, msg string) {
	m.shouldError = shouldError
	m.errorMsg = msg
}

func (m *MockUserRoleService) SetUserRoles(ctx context.Context, userID string, roles []string, assignedBy string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	m.userRoles[userID] = roles
	return nil
}

func (m *MockUserRoleService) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	roles, exists := m.userRoles[userID]
	if !exists {
		return []string{}, nil
	}
	return roles, nil
}

// Test helper functions

func createTestAuthService(t *testing.T) (*AuthService, *MockUserRepository, *MockRefreshTokenRepository, *MockUserRoleService) {
	userRepo := NewMockUserRepository()
	refreshTokenRepo := NewMockRefreshTokenRepository()
	userRoleService := NewMockUserRoleService()
	passwordService := NewPasswordService(12)

	jwtConfig := JWTConfig{
		AccessSecret:  "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "test",
	}

	jwtService, err := NewJWTService(jwtConfig)
	if err != nil {
		t.Fatalf("Failed to create JWT service: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	authService := NewAuthService(userRepo, refreshTokenRepo, passwordService, jwtService, userRoleService, logger)
	return authService, userRepo, refreshTokenRepo, userRoleService
}

// AuthService Tests

func TestAuthService_Register(t *testing.T) {
	authService, userRepo, _, _ := createTestAuthService(t)
	ctx := context.Background()

	t.Run("successful user registration", func(t *testing.T) {
		req := &RegisterRequest{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "SecurePassword123!",
		}

		response, err := authService.Register(ctx, req)
		if err != nil {
			t.Errorf("Register() error = %v, want nil", err)
			return
		}

		if response == nil {
			t.Error("Register() returned nil response")
			return
		}

		// Verify response contains tokens
		if response.AccessToken == "" {
			t.Error("Register() response missing access token")
		}
		if response.RefreshToken == "" {
			t.Error("Register() response missing refresh token")
		}
		if response.ExpiresIn <= 0 {
			t.Error("Register() response invalid expires_in")
		}

		// Verify user was created
		if response.User == nil {
			t.Error("Register() response missing user")
			return
		}

		if response.User.Email != req.Email {
			t.Errorf("Register() user email = %v, want %v", response.User.Email, req.Email)
		}
		if response.User.Name != req.Name {
			t.Errorf("Register() user name = %v, want %v", response.User.Name, req.Name)
		}
		if response.User.EmailVerified {
			t.Error("Register() user should not be email verified initially")
		}

		// Verify password was hashed
		if response.User.PasswordHash == nil {
			t.Error("Register() user missing password hash")
		} else if *response.User.PasswordHash == req.Password {
			t.Error("Register() password was not hashed")
		}

		// Verify user exists in repository
		storedUser, err := userRepo.GetUserByEmail(ctx, req.Email)
		if err != nil {
			t.Errorf("Failed to retrieve created user: %v", err)
		} else if storedUser.ID != response.User.ID {
			t.Error("Created user ID mismatch")
		}
	})

	t.Run("duplicate email registration fails", func(t *testing.T) {
		// First registration
		req1 := &RegisterRequest{
			Name:     "John Doe",
			Email:    "duplicate@example.com",
			Password: "SecurePassword123!",
		}

		_, err := authService.Register(ctx, req1)
		if err != nil {
			t.Errorf("First registration failed: %v", err)
			return
		}

		// Second registration with same email
		req2 := &RegisterRequest{
			Name:     "Jane Doe",
			Email:    "duplicate@example.com",
			Password: "AnotherPassword123!",
		}

		_, err = authService.Register(ctx, req2)
		if err == nil {
			t.Error("Register() with duplicate email should return error")
		}

		if !strings.Contains(err.Error(), "already registered") {
			t.Errorf("Register() error = %v, want error containing 'already registered'", err)
		}
	})

	t.Run("invalid input validation", func(t *testing.T) {
		tests := []struct {
			name string
			req  *RegisterRequest
		}{
			{
				name: "nil request",
				req:  nil,
			},
			{
				name: "empty name",
				req: &RegisterRequest{
					Name:     "",
					Email:    "test@example.com",
					Password: "SecurePassword123!",
				},
			},
			{
				name: "empty email",
				req: &RegisterRequest{
					Name:     "John Doe",
					Email:    "",
					Password: "SecurePassword123!",
				},
			},
			{
				name: "empty password",
				req: &RegisterRequest{
					Name:     "John Doe",
					Email:    "test@example.com",
					Password: "",
				},
			},
			{
				name: "whitespace name",
				req: &RegisterRequest{
					Name:     "   ",
					Email:    "test@example.com",
					Password: "SecurePassword123!",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := authService.Register(ctx, tt.req)
				if err == nil {
					t.Errorf("Register() with %s should return error", tt.name)
				}
			})
		}
	})

	t.Run("password strength requirements", func(t *testing.T) {
		weakPasswords := []string{
			"weak",                   // too short
			"nouppercaseletters123!", // no uppercase
			"NOLOWERCASELETTERS123!", // no lowercase
			"NoNumbers!",             // no numbers
			"NoSpecialChars123",      // no special characters
		}

		for _, password := range weakPasswords {
			req := &RegisterRequest{
				Name:     "Test User",
				Email:    "test" + password + "@example.com",
				Password: password,
			}

			_, err := authService.Register(ctx, req)
			if err == nil {
				t.Errorf("Register() with weak password %q should return error", password)
			}
		}
	})

	t.Run("response includes valid tokens", func(t *testing.T) {
		req := &RegisterRequest{
			Name:     "Token Test User",
			Email:    "tokentest@example.com",
			Password: "SecurePassword123!",
		}

		response, err := authService.Register(ctx, req)
		if err != nil {
			t.Errorf("Register() error = %v", err)
			return
		}

		// Validate access token
		claims, err := authService.ValidateAccessToken(response.AccessToken)
		if err != nil {
			t.Errorf("Register() generated invalid access token: %v", err)
			return
		}

		if claims.UserID != response.User.ID {
			t.Errorf("Access token UserID = %v, want %v", claims.UserID, response.User.ID)
		}
		if claims.Email != response.User.Email {
			t.Errorf("Access token Email = %v, want %v", claims.Email, response.User.Email)
		}
	})

	t.Run("repository error handling", func(t *testing.T) {
		// Test email exists check error
		userRepo.SetError(true, "database connection failed")

		req := &RegisterRequest{
			Name:     "Error Test User",
			Email:    "errortest@example.com",
			Password: "SecurePassword123!",
		}

		_, err := authService.Register(ctx, req)
		if err == nil {
			t.Error("Register() should return error when repository fails")
		}

		// Reset error
		userRepo.SetError(false, "")
	})
}

func TestAuthService_Login(t *testing.T) {
	authService, userRepo, _, _ := createTestAuthService(t)
	ctx := context.Background()

	// Create a test user
	testUser := &User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Name:  "Test User",
	}

	// Hash a known password
	passwordService := NewPasswordService(12)
	hashedPassword, err := passwordService.HashPassword("TestPassword123!")
	if err != nil {
		t.Fatalf("Failed to hash test password: %v", err)
	}
	testUser.PasswordHash = &hashedPassword
	testUser.EmailVerified = true
	testUser.CreatedAt = time.Now()
	testUser.UpdatedAt = time.Now()

	err = userRepo.CreateUser(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("successful login with valid credentials", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "test@example.com",
			Password: "TestPassword123!",
		}

		response, err := authService.Login(ctx, req)
		if err != nil {
			t.Errorf("Login() error = %v, want nil", err)
			return
		}

		if response == nil {
			t.Error("Login() returned nil response")
			return
		}

		// Verify response contains tokens
		if response.AccessToken == "" {
			t.Error("Login() response missing access token")
		}
		if response.RefreshToken == "" {
			t.Error("Login() response missing refresh token")
		}

		// Verify user information
		if response.User == nil {
			t.Error("Login() response missing user")
			return
		}

		if response.User.ID != testUser.ID {
			t.Errorf("Login() user ID = %v, want %v", response.User.ID, testUser.ID)
		}

		// Verify last login was updated
		updatedUser, err := userRepo.GetUserByID(ctx, testUser.ID)
		if err != nil {
			t.Errorf("Failed to get updated user: %v", err)
		} else if updatedUser.LastLogin == nil {
			t.Error("Login() should update last login time")
		}
	})

	t.Run("failed login with invalid credentials", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "test@example.com",
			Password: "WrongPassword123!",
		}

		_, err := authService.Login(ctx, req)
		if err == nil {
			t.Error("Login() with wrong password should return error")
		}

		if !strings.Contains(err.Error(), "invalid credentials") {
			t.Errorf("Login() error = %v, want error containing 'invalid credentials'", err)
		}

		// Verify failed login attempts were incremented
		user, err := userRepo.GetUserByID(ctx, testUser.ID)
		if err != nil {
			t.Errorf("Failed to get user after failed login: %v", err)
		} else if user.FailedLoginAttempts == 0 {
			t.Error("Login() should increment failed login attempts")
		}
	})

	t.Run("account lockout after failed attempts", func(t *testing.T) {
		// Create a new user for this test
		lockedUser := &User{
			ID:                  "locked-user-id",
			Email:               "locked@example.com",
			Name:                "Locked User",
			PasswordHash:        &hashedPassword,
			EmailVerified:       true,
			FailedLoginAttempts: 0,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		err := userRepo.CreateUser(ctx, lockedUser)
		if err != nil {
			t.Fatalf("Failed to create locked test user: %v", err)
		}

		req := &LoginRequest{
			Email:    "locked@example.com",
			Password: "WrongPassword123!",
		}

		// Attempt login 5 times with wrong password
		for i := range 5 {
			_, err := authService.Login(ctx, req)
			if err == nil {
				t.Errorf("Login attempt %d should fail with wrong password", i+1)
			}
		}

		// The 5th attempt should lock the account
		user, err := userRepo.GetUserByID(ctx, lockedUser.ID)
		if err != nil {
			t.Errorf("Failed to get user after failed attempts: %v", err)
		} else {
			if user.FailedLoginAttempts != 5 {
				t.Errorf("User should have 5 failed attempts, got %d", user.FailedLoginAttempts)
			}
			if user.LockedUntil == nil {
				t.Error("User should be locked after 5 failed attempts")
			}
		}

		// Verify that even correct password fails when locked
		correctReq := &LoginRequest{
			Email:    "locked@example.com",
			Password: "TestPassword123!",
		}

		_, err = authService.Login(ctx, correctReq)
		if err == nil {
			t.Error("Login() should fail when account is locked")
		}

		if !strings.Contains(err.Error(), "locked") {
			t.Errorf("Login() error = %v, want error containing 'locked'", err)
		}
	})

	t.Run("response format matches schema", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "test@example.com",
			Password: "TestPassword123!",
		}

		// Reset failed attempts for clean test
		err := userRepo.UpdateUserLoginAttempts(ctx, testUser.ID, 0, nil)
		if err != nil {
			t.Errorf("Failed to reset login attempts: %v", err)
		}

		response, err := authService.Login(ctx, req)
		if err != nil {
			t.Errorf("Login() error = %v", err)
			return
		}

		// Verify response structure
		if response.AccessToken == "" {
			t.Error("Login() response missing AccessToken")
		}
		if response.RefreshToken == "" {
			t.Error("Login() response missing RefreshToken")
		}
		if response.ExpiresIn <= 0 {
			t.Error("Login() response invalid ExpiresIn")
		}
		if response.User == nil {
			t.Error("Login() response missing User")
		}

		// Verify token can be validated
		claims, err := authService.ValidateAccessToken(response.AccessToken)
		if err != nil {
			t.Errorf("Login() generated invalid access token: %v", err)
		} else {
			if claims.UserID != testUser.ID {
				t.Errorf("Access token UserID = %v, want %v", claims.UserID, testUser.ID)
			}
		}
	})

	t.Run("login validation", func(t *testing.T) {
		tests := []struct {
			name string
			req  *LoginRequest
		}{
			{
				name: "nil request",
				req:  nil,
			},
			{
				name: "empty email",
				req: &LoginRequest{
					Email:    "",
					Password: "TestPassword123!",
				},
			},
			{
				name: "empty password",
				req: &LoginRequest{
					Email:    "test@example.com",
					Password: "",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := authService.Login(ctx, tt.req)
				if err == nil {
					t.Errorf("Login() with %s should return error", tt.name)
				}
			})
		}
	})

	t.Run("nonexistent user login fails", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "TestPassword123!",
		}

		_, err := authService.Login(ctx, req)
		if err == nil {
			t.Error("Login() with nonexistent user should return error")
		}

		if !strings.Contains(err.Error(), "invalid credentials") {
			t.Errorf("Login() error = %v, want error containing 'invalid credentials'", err)
		}
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	authService, userRepo, refreshTokenRepo, _ := createTestAuthService(t)
	ctx := context.Background()

	// Create a test user and generate tokens
	testUser := &User{
		ID:            "test-user-id",
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := userRepo.CreateUser(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Generate initial tokens
	registerReq := &RegisterRequest{
		Name:     "Refresh Test User",
		Email:    "refresh@example.com",
		Password: "TestPassword123!",
	}

	registerResponse, err := authService.Register(ctx, registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	t.Run("successful token refresh", func(t *testing.T) {
		response, err := authService.RefreshToken(ctx, registerResponse.RefreshToken)
		if err != nil {
			t.Errorf("RefreshToken() error = %v, want nil", err)
			return
		}

		if response == nil {
			t.Error("RefreshToken() returned nil response")
			return
		}

		// Verify new tokens are different
		if response.AccessToken == registerResponse.AccessToken {
			t.Error("RefreshToken() should return new access token")
		}
		if response.RefreshToken == registerResponse.RefreshToken {
			t.Error("RefreshToken() should return new refresh token")
		}

		// Verify new tokens are valid
		claims, err := authService.ValidateAccessToken(response.AccessToken)
		if err != nil {
			t.Errorf("RefreshToken() generated invalid access token: %v", err)
		} else {
			if claims.UserID != registerResponse.User.ID {
				t.Errorf("New access token UserID = %v, want %v", claims.UserID, registerResponse.User.ID)
			}
		}

		// Verify old refresh token is revoked
		oldTokenHash := authService.jwtService.HashRefreshToken(registerResponse.RefreshToken)
		oldToken, err := refreshTokenRepo.GetRefreshToken(ctx, oldTokenHash)
		if err != nil {
			t.Errorf("Failed to get old refresh token: %v", err)
		} else if oldToken.RevokedAt == nil {
			t.Error("Old refresh token should be revoked")
		}
	})

	t.Run("failed refresh with invalid token", func(t *testing.T) {
		_, err := authService.RefreshToken(ctx, "invalid-refresh-token")
		if err == nil {
			t.Error("RefreshToken() with invalid token should return error")
		}

		if !strings.Contains(err.Error(), "invalid refresh token") {
			t.Errorf("RefreshToken() error = %v, want error containing 'invalid refresh token'", err)
		}
	})

	t.Run("empty refresh token", func(t *testing.T) {
		_, err := authService.RefreshToken(ctx, "")
		if err == nil {
			t.Error("RefreshToken() with empty token should return error")
		}

		if !strings.Contains(err.Error(), "required") {
			t.Errorf("RefreshToken() error = %v, want error containing 'required'", err)
		}
	})

	t.Run("token rotation behavior", func(t *testing.T) {
		// Generate fresh tokens
		freshRegisterReq := &RegisterRequest{
			Name:     "Rotation Test User",
			Email:    "rotation@example.com",
			Password: "TestPassword123!",
		}

		freshResponse, err := authService.Register(ctx, freshRegisterReq)
		if err != nil {
			t.Fatalf("Failed to register rotation test user: %v", err)
		}

		originalRefreshToken := freshResponse.RefreshToken

		// First refresh
		firstRefresh, err := authService.RefreshToken(ctx, originalRefreshToken)
		if err != nil {
			t.Errorf("First RefreshToken() error = %v", err)
			return
		}

		// Second refresh with first refresh token
		secondRefresh, err := authService.RefreshToken(ctx, firstRefresh.RefreshToken)
		if err != nil {
			t.Errorf("Second RefreshToken() error = %v", err)
			return
		}

		// All tokens should be different
		if originalRefreshToken == firstRefresh.RefreshToken {
			t.Error("First refresh should return new refresh token")
		}
		if firstRefresh.RefreshToken == secondRefresh.RefreshToken {
			t.Error("Second refresh should return new refresh token")
		}
		if originalRefreshToken == secondRefresh.RefreshToken {
			t.Error("Second refresh token should differ from original")
		}

		// Original refresh token should not work anymore
		_, err = authService.RefreshToken(ctx, originalRefreshToken)
		if err == nil {
			t.Error("Original refresh token should be invalid after being used")
		}
	})
}

func TestAuthService_Logout(t *testing.T) {
	authService, _, refreshTokenRepo, _ := createTestAuthService(t)
	ctx := context.Background()

	// Create test user and tokens
	registerReq := &RegisterRequest{
		Name:     "Logout Test User",
		Email:    "logout@example.com",
		Password: "TestPassword123!",
	}

	registerResponse, err := authService.Register(ctx, registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	userID := registerResponse.User.ID

	t.Run("successful logout", func(t *testing.T) {
		err := authService.Logout(ctx, userID)
		if err != nil {
			t.Errorf("Logout() error = %v, want nil", err)
		}

		// Verify refresh token is revoked
		tokenHash := authService.jwtService.HashRefreshToken(registerResponse.RefreshToken)
		token, err := refreshTokenRepo.GetRefreshToken(ctx, tokenHash)
		if err != nil {
			t.Errorf("Failed to get refresh token: %v", err)
		} else if token.RevokedAt == nil {
			t.Error("Refresh token should be revoked after logout")
		}
	})

	t.Run("logout with repository error", func(t *testing.T) {
		refreshTokenRepo.SetError(true, "database connection failed")

		err := authService.Logout(ctx, userID)
		if err == nil {
			t.Error("Logout() should return error when repository fails")
		}

		refreshTokenRepo.SetError(false, "")
	})
}

func TestAuthService_GetUserByID(t *testing.T) {
	authService, userRepo, _, _ := createTestAuthService(t)
	ctx := context.Background()

	// Create test user
	testUser := &User{
		ID:            "get-user-test-id",
		Email:         "getuser@example.com",
		Name:          "Get User Test",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := userRepo.CreateUser(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	t.Run("successful user retrieval", func(t *testing.T) {
		user, err := authService.GetUserByID(ctx, testUser.ID)
		if err != nil {
			t.Errorf("GetUserByID() error = %v, want nil", err)
			return
		}

		if user == nil {
			t.Error("GetUserByID() returned nil user")
			return
		}

		if user.ID != testUser.ID {
			t.Errorf("GetUserByID() ID = %v, want %v", user.ID, testUser.ID)
		}
		if user.Email != testUser.Email {
			t.Errorf("GetUserByID() Email = %v, want %v", user.Email, testUser.Email)
		}
	})

	t.Run("nonexistent user", func(t *testing.T) {
		_, err := authService.GetUserByID(ctx, "nonexistent-user-id")
		if err == nil {
			t.Error("GetUserByID() with nonexistent ID should return error")
		}

		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("GetUserByID() error = %v, want error containing 'not found'", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		userRepo.SetError(true, "database connection failed")

		_, err := authService.GetUserByID(ctx, testUser.ID)
		if err == nil {
			t.Error("GetUserByID() should return error when repository fails")
		}

		userRepo.SetError(false, "")
	})
}

func TestAuthService_ValidateAccessToken(t *testing.T) {
	authService, _, _, _ := createTestAuthService(t)
	ctx := context.Background()

	// Generate test tokens
	registerReq := &RegisterRequest{
		Name:     "Validate Test User",
		Email:    "validate@example.com",
		Password: "TestPassword123!",
	}

	registerResponse, err := authService.Register(ctx, registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	t.Run("valid token validation", func(t *testing.T) {
		claims, err := authService.ValidateAccessToken(registerResponse.AccessToken)
		if err != nil {
			t.Errorf("ValidateAccessToken() error = %v, want nil", err)
			return
		}

		if claims == nil {
			t.Error("ValidateAccessToken() returned nil claims")
			return
		}

		if claims.UserID != registerResponse.User.ID {
			t.Errorf("ValidateAccessToken() UserID = %v, want %v", claims.UserID, registerResponse.User.ID)
		}
		if claims.Email != registerResponse.User.Email {
			t.Errorf("ValidateAccessToken() Email = %v, want %v", claims.Email, registerResponse.User.Email)
		}
	})

	t.Run("invalid token validation", func(t *testing.T) {
		_, err := authService.ValidateAccessToken("invalid-token")
		if err == nil {
			t.Error("ValidateAccessToken() with invalid token should return error")
		}
	})

	t.Run("empty token validation", func(t *testing.T) {
		_, err := authService.ValidateAccessToken("")
		if err == nil {
			t.Error("ValidateAccessToken() with empty token should return error")
		}
	})
}

// Role-based Registration Tests

func TestAuthService_RegisterWithRoles(t *testing.T) {
	authService, userRepo, _, userRoleService := createTestAuthService(t)
	ctx := context.Background()

	t.Run("successful registration with VOLUNTEER role", func(t *testing.T) {
		role := RoleVolunteer
		req := &RegisterRequest{
			Name:     "John Volunteer",
			Email:    "john.volunteer@example.com",
			Password: "SecurePassword123!",
			Role:     &role,
		}

		response, err := authService.Register(ctx, req)
		if err != nil {
			t.Errorf("Register() with VOLUNTEER role error = %v, want nil", err)
			return
		}

		if response == nil {
			t.Error("Register() returned nil response")
			return
		}

		// Verify user was created
		if response.User == nil {
			t.Error("Register() response missing user")
			return
		}

		// Verify role was assigned
		assignedRoles, err := userRoleService.GetUserRoles(ctx, response.User.ID)
		if err != nil {
			t.Errorf("Failed to get user roles: %v", err)
			return
		}

		if len(assignedRoles) != 1 || assignedRoles[0] != RoleVolunteer {
			t.Errorf("Expected user to have VOLUNTEER role, got: %v", assignedRoles)
		}

		// Verify tokens are valid
		if response.AccessToken == "" {
			t.Error("Register() response missing access token")
		}
		if response.RefreshToken == "" {
			t.Error("Register() response missing refresh token")
		}
	})

	t.Run("successful registration with ORGANIZER role", func(t *testing.T) {
		role := RoleOrganizer
		req := &RegisterRequest{
			Name:     "Jane Organizer",
			Email:    "jane.organizer@example.com",
			Password: "SecurePassword123!",
			Role:     &role,
		}

		response, err := authService.Register(ctx, req)
		if err != nil {
			t.Errorf("Register() with ORGANIZER role error = %v, want nil", err)
			return
		}

		// Verify role was assigned
		assignedRoles, err := userRoleService.GetUserRoles(ctx, response.User.ID)
		if err != nil {
			t.Errorf("Failed to get user roles: %v", err)
			return
		}

		if len(assignedRoles) != 1 || assignedRoles[0] != RoleOrganizer {
			t.Errorf("Expected user to have ORGANIZER role, got: %v", assignedRoles)
		}
	})

	t.Run("registration with no role defaults to VOLUNTEER", func(t *testing.T) {
		req := &RegisterRequest{
			Name:     "Default User",
			Email:    "default@example.com",
			Password: "SecurePassword123!",
			Role:     nil, // No role specified
		}

		response, err := authService.Register(ctx, req)
		if err != nil {
			t.Errorf("Register() with no role error = %v, want nil", err)
			return
		}

		// Verify default role was assigned
		assignedRoles, err := userRoleService.GetUserRoles(ctx, response.User.ID)
		if err != nil {
			t.Errorf("Failed to get user roles: %v", err)
			return
		}

		if len(assignedRoles) != 1 || assignedRoles[0] != RoleVolunteer {
			t.Errorf("Expected user to have default VOLUNTEER role, got: %v", assignedRoles)
		}
	})

	t.Run("registration with empty role defaults to VOLUNTEER", func(t *testing.T) {
		emptyRole := ""
		req := &RegisterRequest{
			Name:     "Empty Role User",
			Email:    "empty@example.com",
			Password: "SecurePassword123!",
			Role:     &emptyRole,
		}

		response, err := authService.Register(ctx, req)
		if err != nil {
			t.Errorf("Register() with empty role error = %v, want nil", err)
			return
		}

		// Verify default role was assigned
		assignedRoles, err := userRoleService.GetUserRoles(ctx, response.User.ID)
		if err != nil {
			t.Errorf("Failed to get user roles: %v", err)
			return
		}

		if len(assignedRoles) != 1 || assignedRoles[0] != RoleVolunteer {
			t.Errorf("Expected user to have default VOLUNTEER role, got: %v", assignedRoles)
		}
	})

	t.Run("registration with invalid role fails validation", func(t *testing.T) {
		invalidRole := "ADMIN" // ADMIN is not allowed for self-registration
		req := &RegisterRequest{
			Name:     "Invalid Role User",
			Email:    "invalid@example.com",
			Password: "SecurePassword123!",
			Role:     &invalidRole,
		}

		response, err := authService.Register(ctx, req)
		if err == nil {
			t.Error("Register() with ADMIN role should fail validation")
			return
		}

		if response != nil {
			t.Error("Register() with invalid role should return nil response")
		}

		// Verify error message mentions invalid role
		if !strings.Contains(err.Error(), "invalid role") {
			t.Errorf("Expected error to mention 'invalid role', got: %v", err)
		}
	})

	t.Run("registration with unrecognized role fails validation", func(t *testing.T) {
		invalidRole := "INVALID_ROLE"
		req := &RegisterRequest{
			Name:     "Invalid Role User",
			Email:    "invalid2@example.com",
			Password: "SecurePassword123!",
			Role:     &invalidRole,
		}

		response, err := authService.Register(ctx, req)
		if err == nil {
			t.Error("Register() with unrecognized role should fail validation")
			return
		}

		if response != nil {
			t.Error("Register() with invalid role should return nil response")
		}
	})

	t.Run("role assignment failure doesn't prevent user creation", func(t *testing.T) {
		// Set user role service to return error
		userRoleService.SetError(true, "role assignment service unavailable")

		role := RoleVolunteer
		req := &RegisterRequest{
			Name:     "Role Error User",
			Email:    "roleerror@example.com",
			Password: "SecurePassword123!",
			Role:     &role,
		}

		response, err := authService.Register(ctx, req)
		if err != nil {
			t.Errorf("Register() should succeed even if role assignment fails, got error: %v", err)
			return
		}

		// User should still be created
		if response == nil || response.User == nil {
			t.Error("Register() should create user even if role assignment fails")
			return
		}

		// Verify user exists in repository
		user, err := userRepo.GetUserByID(ctx, response.User.ID)
		if err != nil {
			t.Errorf("User should exist in repository: %v", err)
		}
		if user.Email != req.Email {
			t.Errorf("Expected user email %s, got %s", req.Email, user.Email)
		}

		// Reset error state
		userRoleService.SetError(false, "")
	})
}

func TestAuthService_LoginWithRoles(t *testing.T) {
	authService, userRepo, _, userRoleService := createTestAuthService(t)
	ctx := context.Background()

	// Create a test user with ORGANIZER role
	user := &User{
		ID:                  "test-organizer-id",
		Email:               "organizer@example.com",
		Name:                "Test Organizer",
		EmailVerified:       false,
		FailedLoginAttempts: 0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	passwordService := NewPasswordService(12)
	hashedPassword, _ := passwordService.HashPassword("TestPassword123!")
	user.PasswordHash = &hashedPassword

	userRepo.CreateUser(ctx, user)
	userRoleService.SetUserRoles(ctx, user.ID, []string{RoleOrganizer}, user.ID)

	t.Run("login returns user roles in token", func(t *testing.T) {
		req := &LoginRequest{
			Email:    "organizer@example.com",
			Password: "TestPassword123!",
		}

		response, err := authService.Login(ctx, req)
		if err != nil {
			t.Errorf("Login() error = %v, want nil", err)
			return
		}

		if response == nil {
			t.Error("Login() returned nil response")
			return
		}

		// Validate the access token contains correct roles
		claims, err := authService.ValidateAccessToken(response.AccessToken)
		if err != nil {
			t.Errorf("Failed to validate access token: %v", err)
			return
		}

		if len(claims.Roles) != 1 || claims.Roles[0] != RoleOrganizer {
			t.Errorf("Expected token to contain ORGANIZER role, got: %v", claims.Roles)
		}
	})

	t.Run("login with role service error uses fallback", func(t *testing.T) {
		// Set user role service to return error
		userRoleService.SetError(true, "role service unavailable")

		req := &LoginRequest{
			Email:    "organizer@example.com",
			Password: "TestPassword123!",
		}

		response, err := authService.Login(ctx, req)
		if err != nil {
			t.Errorf("Login() should succeed even with role service error, got: %v", err)
			return
		}

		// Should get fallback role
		claims, err := authService.ValidateAccessToken(response.AccessToken)
		if err != nil {
			t.Errorf("Failed to validate access token: %v", err)
			return
		}

		if len(claims.Roles) != 1 || claims.Roles[0] != RoleVolunteer {
			t.Errorf("Expected token to contain fallback VOLUNTEER role, got: %v", claims.Roles)
		}

		// Reset error state
		userRoleService.SetError(false, "")
	})
}

func TestAuthService_RefreshTokenWithRoles(t *testing.T) {
	authService, userRepo, _, userRoleService := createTestAuthService(t)
	ctx := context.Background()

	// Create a test user
	user := &User{
		ID:                  "test-refresh-user-id",
		Email:               "refresh@example.com",
		Name:                "Test Refresh User",
		EmailVerified:       false,
		FailedLoginAttempts: 0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	passwordService := NewPasswordService(12)
	hashedPassword, _ := passwordService.HashPassword("TestPassword123!")
	user.PasswordHash = &hashedPassword

	userRepo.CreateUser(ctx, user)
	userRoleService.SetUserRoles(ctx, user.ID, []string{RoleVolunteer}, user.ID)

	// Login to get initial tokens
	loginReq := &LoginRequest{
		Email:    "refresh@example.com",
		Password: "TestPassword123!",
	}

	loginResponse, err := authService.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}

	t.Run("token refresh retrieves updated roles", func(t *testing.T) {
		// Update user roles
		userRoleService.SetUserRoles(ctx, user.ID, []string{RoleOrganizer}, user.ID)

		// Refresh token
		refreshResponse, err := authService.RefreshToken(ctx, loginResponse.RefreshToken)
		if err != nil {
			t.Errorf("RefreshToken() error = %v, want nil", err)
			return
		}

		// Validate the new access token contains updated roles
		claims, err := authService.ValidateAccessToken(refreshResponse.AccessToken)
		if err != nil {
			t.Errorf("Failed to validate refreshed access token: %v", err)
			return
		}

		if len(claims.Roles) != 1 || claims.Roles[0] != RoleOrganizer {
			t.Errorf("Expected refreshed token to contain updated ORGANIZER role, got: %v", claims.Roles)
		}
	})

	t.Run("token refresh with role service error uses existing roles", func(t *testing.T) {
		// Get fresh tokens for this test
		freshLoginResponse, err := authService.Login(ctx, loginReq)
		if err != nil {
			t.Fatalf("Failed to get fresh login tokens: %v", err)
		}

		// Set user role service to return error
		userRoleService.SetError(true, "role service unavailable")

		// Refresh token
		refreshResponse, err := authService.RefreshToken(ctx, freshLoginResponse.RefreshToken)
		if err != nil {
			t.Errorf("RefreshToken() should succeed even with role service error, got: %v", err)
			return
		}

		// Should use existing roles from the old token
		claims, err := authService.ValidateAccessToken(refreshResponse.AccessToken)
		if err != nil {
			t.Errorf("Failed to validate access token: %v", err)
			return
		}

		// Should have at least one role (fallback)
		if len(claims.Roles) == 0 {
			t.Error("Expected token to contain at least one role")
		}

		// Reset error state
		userRoleService.SetError(false, "")
	})
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		wantErr  bool
		errorMsg string
	}{
		{
			name:    "valid VOLUNTEER role",
			role:    RoleVolunteer,
			wantErr: false,
		},
		{
			name:    "valid ORGANIZER role",
			role:    RoleOrganizer,
			wantErr: false,
		},
		{
			name:    "empty role is allowed",
			role:    "",
			wantErr: false,
		},
		{
			name:     "ADMIN role is not allowed for registration",
			role:     RoleAdmin,
			wantErr:  true,
			errorMsg: "invalid role",
		},
		{
			name:     "invalid role string",
			role:     "INVALID_ROLE",
			wantErr:  true,
			errorMsg: "invalid role",
		},
		{
			name:     "lowercase role not accepted",
			role:     "volunteer",
			wantErr:  true,
			errorMsg: "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRole(tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("ValidateRole() error = %v, want error containing %s", err, tt.errorMsg)
			}
		})
	}
}

func TestIsValidRegistrationRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{
			name: "VOLUNTEER is valid",
			role: RoleVolunteer,
			want: true,
		},
		{
			name: "ORGANIZER is valid",
			role: RoleOrganizer,
			want: true,
		},
		{
			name: "empty role is valid",
			role: "",
			want: true,
		},
		{
			name: "ADMIN is not valid for registration",
			role: RoleAdmin,
			want: false,
		},
		{
			name: "invalid role is not valid",
			role: "INVALID",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRegistrationRole(tt.role); got != tt.want {
				t.Errorf("IsValidRegistrationRole() = %v, want %v", got, tt.want)
			}
		})
	}
}
