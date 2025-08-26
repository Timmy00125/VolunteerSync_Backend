package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/volunteersync/backend/internal/core/auth"
)

// Mock auth service for testing
type MockAuthService struct {
	shouldError     bool
	errorMsg        string
	claims          *auth.UserClaims
	user            *auth.User
	shouldUserError bool
	userErrorMsg    string
}

func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		user: &auth.User{
			ID:            "test-user-id",
			Email:         "test@example.com",
			Name:          "Test User",
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}
}

func (m *MockAuthService) SetError(shouldError bool, msg string) {
	m.shouldError = shouldError
	m.errorMsg = msg
}

func (m *MockAuthService) SetUserError(shouldError bool, msg string) {
	m.shouldUserError = shouldError
	m.userErrorMsg = msg
}

func (m *MockAuthService) SetClaims(claims *auth.UserClaims) {
	m.claims = claims
}

func (m *MockAuthService) SetUser(user *auth.User) {
	m.user = user
}

func (m *MockAuthService) ValidateAccessToken(token string) (*auth.UserClaims, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	if m.claims != nil {
		return m.claims, nil
	}
	return &auth.UserClaims{
		UserID:    "test-user-id",
		Email:     "test@example.com",
		Roles:     []string{"user"},
		TokenType: auth.AccessTokenType,
	}, nil
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	if m.shouldUserError {
		return nil, errors.New(m.userErrorMsg)
	}
	if m.user != nil {
		return m.user, nil
	}
	return &auth.User{
		ID:            userID,
		Email:         "test@example.com",
		Name:          "Test User",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

// Implement other methods required by AuthService interface
func (m *MockAuthService) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.AuthResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) Login(ctx context.Context, req *auth.LoginRequest) (*auth.AuthResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*auth.AuthResponse, error) {
	return nil, errors.New("not implemented")
}

func (m *MockAuthService) Logout(ctx context.Context, userID string) error {
	return errors.New("not implemented")
}

func createTestAuthMiddleware(t *testing.T) (*AuthMiddleware, *MockAuthService) {
	mockAuthService := NewMockAuthService()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	middleware := NewAuthMiddleware(mockAuthService, logger)
	return middleware, mockAuthService
}

func TestNewAuthMiddleware(t *testing.T) {
	mockAuthService := NewMockAuthService()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	middleware := NewAuthMiddleware(mockAuthService, logger)

	if middleware == nil {
		t.Error("NewAuthMiddleware() returned nil")
	}

	if middleware.logger != logger {
		t.Error("NewAuthMiddleware() logger not set correctly")
	}
}

func TestAuthMiddleware_RequireAuth_ValidRequest(t *testing.T) {
	middleware, mockAuthService := createTestAuthMiddleware(t)

	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		token      string
		authHeader string
		wantStatus int
	}{
		{
			name:       "valid Bearer token",
			token:      "valid-jwt-token",
			authHeader: "Bearer valid-jwt-token",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockAuthService.SetError(false, "")
			mockAuthService.SetUserError(false, "")
			mockAuthService.SetClaims(&auth.UserClaims{
				UserID:    "test-user-id",
				Email:     "test@example.com",
				Roles:     []string{"user"},
				TokenType: auth.AccessTokenType,
			})

			// Create test request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/protected", nil)
			c.Request.Header.Set("Authorization", tt.authHeader)

			// Add a test handler that will only be called if auth succeeds
			handlerCalled := false
			testHandler := func(c *gin.Context) {
				handlerCalled = true

				// Test that user context is properly set
				userClaims := GetUserClaimsFromContext(c.Request.Context())
				if userClaims == nil {
					t.Error("User claims not found in context")
				} else {
					if userClaims.UserID != "test-user-id" {
						t.Errorf("Context UserID = %v, want test-user-id", userClaims.UserID)
					}
					if userClaims.Email != "test@example.com" {
						t.Errorf("Context Email = %v, want test@example.com", userClaims.Email)
					}
				}

				user := GetUserFromContext(c.Request.Context())
				if user == nil {
					t.Error("User not found in context")
				}

				c.JSON(http.StatusOK, gin.H{"message": "success"})
			}

			// Create handler chain
			authHandler := middleware.RequireAuth()
			authHandler(c)

			if !c.IsAborted() {
				testHandler(c)
			}

			// Verify response
			if w.Code != tt.wantStatus {
				t.Errorf("RequireAuth() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK && !handlerCalled {
				t.Error("RequireAuth() should have called next handler")
			}
		})
	}
}

func TestAuthMiddleware_RequireAuth_InvalidRequest(t *testing.T) {
	middleware, mockAuthService := createTestAuthMiddleware(t)

	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		authHeader   string
		mockError    bool
		mockErrorMsg string
		userError    bool
		userErrorMsg string
		wantStatus   int
		wantBody     string
	}{
		{
			name:       "missing Authorization header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Authorization token required",
		},
		{
			name:         "invalid token",
			authHeader:   "Bearer invalid-token",
			mockError:    true,
			mockErrorMsg: "invalid token",
			wantStatus:   http.StatusUnauthorized,
			wantBody:     "Invalid authorization token",
		},
		{
			name:         "user not found",
			authHeader:   "Bearer valid-token",
			userError:    true,
			userErrorMsg: "user not found",
			wantStatus:   http.StatusUnauthorized,
			wantBody:     "User not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockAuthService.SetError(tt.mockError, tt.mockErrorMsg)
			mockAuthService.SetUserError(tt.userError, tt.userErrorMsg)

			// Create test request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}

			// Add a test handler that should NOT be called
			handlerCalled := false
			testHandler := func(c *gin.Context) {
				handlerCalled = true
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			}

			// Create handler chain
			authHandler := middleware.RequireAuth()
			authHandler(c)

			if !c.IsAborted() {
				testHandler(c)
			}

			// Verify response
			if w.Code != tt.wantStatus {
				t.Errorf("RequireAuth() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if handlerCalled {
				t.Error("RequireAuth() should NOT have called next handler")
			}

			if tt.wantBody != "" {
				responseBody := w.Body.String()
				if !contains(responseBody, tt.wantBody) {
					t.Errorf("RequireAuth() body = %v, want containing %v", responseBody, tt.wantBody)
				}
			}

			// Verify request was aborted
			if !c.IsAborted() {
				t.Error("RequireAuth() should have aborted the request")
			}
		})
	}
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	middleware, mockAuthService := createTestAuthMiddleware(t)

	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	t.Run("valid token provided", func(t *testing.T) {
		mockAuthService.SetError(false, "")
		mockAuthService.SetUserError(false, "")
		mockAuthService.SetClaims(&auth.UserClaims{
			UserID:    "test-user-id",
			Email:     "test@example.com",
			Roles:     []string{"user"},
			TokenType: auth.AccessTokenType,
		})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/public", nil)
		c.Request.Header.Set("Authorization", "Bearer valid-token")

		handlerCalled := false
		testHandler := func(c *gin.Context) {
			handlerCalled = true

			// Verify user context is set
			userClaims := GetUserClaimsFromContext(c.Request.Context())
			if userClaims == nil {
				t.Error("User claims should be set when valid token provided")
			} else if userClaims.UserID != "test-user-id" {
				t.Errorf("Context UserID = %v, want test-user-id", userClaims.UserID)
			}

			user := GetUserFromContext(c.Request.Context())
			if user == nil {
				t.Error("User should be set when valid token provided")
			}

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		}

		authHandler := middleware.OptionalAuth()
		authHandler(c)
		testHandler(c)

		if w.Code != http.StatusOK {
			t.Errorf("OptionalAuth() status = %v, want %v", w.Code, http.StatusOK)
		}

		if !handlerCalled {
			t.Error("OptionalAuth() should have called next handler")
		}

		if c.IsAborted() {
			t.Error("OptionalAuth() should not have aborted the request")
		}
	})

	t.Run("no token provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/public", nil)

		handlerCalled := false
		testHandler := func(c *gin.Context) {
			handlerCalled = true

			// Verify no user context is set
			userClaims := GetUserClaimsFromContext(c.Request.Context())
			if userClaims != nil {
				t.Error("User claims should not be set when no token provided")
			}

			user := GetUserFromContext(c.Request.Context())
			if user != nil {
				t.Error("User should not be set when no token provided")
			}

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		}

		authHandler := middleware.OptionalAuth()
		authHandler(c)
		testHandler(c)

		if w.Code != http.StatusOK {
			t.Errorf("OptionalAuth() status = %v, want %v", w.Code, http.StatusOK)
		}

		if !handlerCalled {
			t.Error("OptionalAuth() should have called next handler")
		}

		if c.IsAborted() {
			t.Error("OptionalAuth() should not have aborted the request")
		}
	})

	t.Run("invalid token provided", func(t *testing.T) {
		mockAuthService.SetError(true, "invalid token")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/public", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		handlerCalled := false
		testHandler := func(c *gin.Context) {
			handlerCalled = true

			// Verify no user context is set for invalid token
			userClaims := GetUserClaimsFromContext(c.Request.Context())
			if userClaims != nil {
				t.Error("User claims should not be set when invalid token provided")
			}

			user := GetUserFromContext(c.Request.Context())
			if user != nil {
				t.Error("User should not be set when invalid token provided")
			}

			c.JSON(http.StatusOK, gin.H{"message": "success"})
		}

		authHandler := middleware.OptionalAuth()
		authHandler(c)
		testHandler(c)

		if w.Code != http.StatusOK {
			t.Errorf("OptionalAuth() status = %v, want %v", w.Code, http.StatusOK)
		}

		if !handlerCalled {
			t.Error("OptionalAuth() should have called next handler even with invalid token")
		}

		if c.IsAborted() {
			t.Error("OptionalAuth() should not have aborted the request")
		}
	})
}

func TestAuthMiddleware_ExtractToken(t *testing.T) {
	middleware, _ := createTestAuthMiddleware(t)

	tests := []struct {
		name       string
		authHeader string
		expected   string
	}{
		{
			name:       "Bearer token",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:       "Bearer with extra spaces",
			authHeader: "Bearer   eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9   ",
			expected:   "  eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9   ",
		},
		{
			name:       "Empty header",
			authHeader: "",
			expected:   "",
		},
		{
			name:       "Whitespace only",
			authHeader: "   ",
			expected:   "",
		},
		{
			name:       "Bearer only",
			authHeader: "Bearer",
			expected:   "",
		},
		{
			name:       "Bearer with empty token",
			authHeader: "Bearer  ",
			expected:   " ",
		},
		{
			name:       "Token without Bearer prefix",
			authHeader: "raw-token",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				c.Request.Header.Set("Authorization", tt.authHeader)
			}

			result := middleware.extractToken(c)
			if result != tt.expected {
				t.Errorf("extractToken() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAuthMiddleware_RequireRoles(t *testing.T) {
	middleware, _ := createTestAuthMiddleware(t)

	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	t.Run("user has required role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/admin", nil)

		// Set user claims in context
		claims := &auth.UserClaims{
			UserID: "test-user-id",
			Email:  "test@example.com",
			Roles:  []string{"user", "admin"},
		}
		ctx := context.WithValue(c.Request.Context(), UserClaimsContextKey, claims)
		c.Request = c.Request.WithContext(ctx)

		handlerCalled := false
		testHandler := func(c *gin.Context) {
			handlerCalled = true
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		}

		roleHandler := middleware.RequireRoles("admin")
		roleHandler(c)

		if !c.IsAborted() {
			testHandler(c)
		}

		if w.Code != http.StatusOK {
			t.Errorf("RequireRoles() status = %v, want %v", w.Code, http.StatusOK)
		}

		if !handlerCalled {
			t.Error("RequireRoles() should have called next handler")
		}
	})

	t.Run("user does not have required role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/admin", nil)

		// Set user claims in context without admin role
		claims := &auth.UserClaims{
			UserID: "test-user-id",
			Email:  "test@example.com",
			Roles:  []string{"user"},
		}
		ctx := context.WithValue(c.Request.Context(), UserClaimsContextKey, claims)
		c.Request = c.Request.WithContext(ctx)

		handlerCalled := false
		testHandler := func(c *gin.Context) {
			handlerCalled = true
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		}

		roleHandler := middleware.RequireRoles("admin")
		roleHandler(c)

		if !c.IsAborted() {
			testHandler(c)
		}

		if w.Code != http.StatusForbidden {
			t.Errorf("RequireRoles() status = %v, want %v", w.Code, http.StatusForbidden)
		}

		if handlerCalled {
			t.Error("RequireRoles() should NOT have called next handler")
		}
	})

	t.Run("no authentication", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/admin", nil)

		handlerCalled := false
		testHandler := func(c *gin.Context) {
			handlerCalled = true
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		}

		roleHandler := middleware.RequireRoles("admin")
		roleHandler(c)

		if !c.IsAborted() {
			testHandler(c)
		}

		if w.Code != http.StatusUnauthorized {
			t.Errorf("RequireRoles() status = %v, want %v", w.Code, http.StatusUnauthorized)
		}

		if handlerCalled {
			t.Error("RequireRoles() should NOT have called next handler")
		}
	})
}

func TestContextHelperFunctions(t *testing.T) {
	claims := &auth.UserClaims{
		UserID: "test-user-id",
		Email:  "test@example.com",
		Roles:  []string{"user", "admin"},
	}

	user := &auth.User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Name:  "Test User",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, UserClaimsContextKey, claims)
	ctx = context.WithValue(ctx, UserContextKey, user)

	t.Run("GetUserClaimsFromContext", func(t *testing.T) {
		result := GetUserClaimsFromContext(ctx)
		if result == nil {
			t.Error("GetUserClaimsFromContext() returned nil")
		} else if result.UserID != claims.UserID {
			t.Errorf("GetUserClaimsFromContext() UserID = %v, want %v", result.UserID, claims.UserID)
		}
	})

	t.Run("GetUserFromContext", func(t *testing.T) {
		result := GetUserFromContext(ctx)
		if result == nil {
			t.Error("GetUserFromContext() returned nil")
		} else if result.ID != user.ID {
			t.Errorf("GetUserFromContext() ID = %v, want %v", result.ID, user.ID)
		}
	})

	t.Run("GetUserIDFromContext", func(t *testing.T) {
		result := GetUserIDFromContext(ctx)
		if result != claims.UserID {
			t.Errorf("GetUserIDFromContext() = %v, want %v", result, claims.UserID)
		}
	})

	t.Run("GetUserEmailFromContext", func(t *testing.T) {
		result := GetUserEmailFromContext(ctx)
		if result != claims.Email {
			t.Errorf("GetUserEmailFromContext() = %v, want %v", result, claims.Email)
		}
	})

	t.Run("IsAuthenticated", func(t *testing.T) {
		if !IsAuthenticated(ctx) {
			t.Error("IsAuthenticated() should return true when user is in context")
		}

		emptyCtx := context.Background()
		if IsAuthenticated(emptyCtx) {
			t.Error("IsAuthenticated() should return false when no user in context")
		}
	})

	t.Run("HasRole", func(t *testing.T) {
		if !HasRole(ctx, "admin") {
			t.Error("HasRole() should return true for admin role")
		}

		if HasRole(ctx, "superuser") {
			t.Error("HasRole() should return false for superuser role")
		}
	})

	t.Run("HasAnyRole", func(t *testing.T) {
		if !HasAnyRole(ctx, "admin", "superuser") {
			t.Error("HasAnyRole() should return true when user has admin role")
		}

		if HasAnyRole(ctx, "superuser", "moderator") {
			t.Error("HasAnyRole() should return false when user has none of the roles")
		}
	})
}

// Helper functions

func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 ||
		(len(str) > len(substr) && contains(str[1:], substr)) ||
		str[:len(substr)] == substr)
}

// Benchmark tests
func BenchmarkAuthMiddleware_RequireAuth(b *testing.B) {
	middleware, mockAuthService := createTestAuthMiddleware(&testing.T{})
	mockAuthService.SetError(false, "")
	mockAuthService.SetUserError(false, "")

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/protected", nil)
	c.Request.Header.Set("Authorization", "Bearer valid-token")

	authHandler := middleware.RequireAuth()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset context for each iteration
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/protected", nil)
		c.Request.Header.Set("Authorization", "Bearer valid-token")
		authHandler(c)
	}
}
