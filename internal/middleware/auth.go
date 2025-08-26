package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/volunteersync/backend/internal/core/auth"
)

// AuthService interface defines the authentication methods needed by middleware
type AuthService interface {
	ValidateAccessToken(token string) (*auth.UserClaims, error)
	GetUserByID(ctx context.Context, userID string) (*auth.User, error)
}

// ContextKey type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the key for user information in context
	UserContextKey ContextKey = "user"
	// UserClaimsContextKey is the key for user claims in context
	UserClaimsContextKey ContextKey = "user_claims"
)

// AuthMiddleware provides authentication middleware functionality
type AuthMiddleware struct {
	authService AuthService
	logger      *slog.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService AuthService, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth is middleware that requires valid authentication
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			am.logger.Warn("missing authorization token", "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		claims, err := am.authService.ValidateAccessToken(token)
		if err != nil {
			am.logger.Warn("invalid authorization token", "error", err, "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization token"})
			c.Abort()
			return
		}

		// Get full user information
		user, err := am.authService.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil {
			am.logger.Error("failed to get user", "user_id", claims.UserID, "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Check if user account is locked
		if user.IsLocked() {
			am.logger.Warn("access attempt with locked account", "user_id", user.ID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is temporarily locked"})
			c.Abort()
			return
		}

		// Add user and claims to context
		ctx := context.WithValue(c.Request.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, UserClaimsContextKey, claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	})
}

// OptionalAuth is middleware that extracts user information if token is present
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		token := am.extractToken(c)
		if token == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		claims, err := am.authService.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, but we don't abort for optional auth
			am.logger.Debug("invalid optional auth token", "error", err, "path", c.Request.URL.Path)
			c.Next()
			return
		}

		// Get full user information
		user, err := am.authService.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil {
			am.logger.Debug("failed to get user for optional auth", "user_id", claims.UserID, "error", err)
			c.Next()
			return
		}

		// Check if user account is locked
		if user.IsLocked() {
			am.logger.Debug("optional auth with locked account", "user_id", user.ID)
			c.Next()
			return
		}

		// Add user and claims to context
		ctx := context.WithValue(c.Request.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, UserClaimsContextKey, claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	})
}

// RequireRoles is middleware that requires specific roles
func (am *AuthMiddleware) RequireRoles(requiredRoles ...string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		claims := GetUserClaimsFromContext(c.Request.Context())
		if claims == nil {
			am.logger.Warn("role check without authentication", "path", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		if !am.hasAnyRole(claims.Roles, requiredRoles) {
			am.logger.Warn("insufficient permissions", "user_id", claims.UserID, "required_roles", requiredRoles, "user_roles", claims.Roles)
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	})
}

// extractToken extracts the JWT token from the Authorization header
func (am *AuthMiddleware) extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token format
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// hasAnyRole checks if user has any of the required roles
func (am *AuthMiddleware) hasAnyRole(userRoles, requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true
	}

	roleMap := make(map[string]bool)
	for _, role := range userRoles {
		roleMap[role] = true
	}

	for _, requiredRole := range requiredRoles {
		if roleMap[requiredRole] {
			return true
		}
	}

	return false
}

// Helper functions for extracting information from context

// GetUserFromContext extracts user information from context
func GetUserFromContext(ctx context.Context) *auth.User {
	user, ok := ctx.Value(UserContextKey).(*auth.User)
	if !ok {
		return nil
	}
	return user
}

// GetUserClaimsFromContext extracts user claims from context
func GetUserClaimsFromContext(ctx context.Context) *auth.UserClaims {
	claims, ok := ctx.Value(UserClaimsContextKey).(*auth.UserClaims)
	if !ok {
		return nil
	}
	return claims
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) string {
	claims := GetUserClaimsFromContext(ctx)
	if claims == nil {
		return ""
	}
	return claims.UserID
}

// GetUserEmailFromContext extracts user email from context
func GetUserEmailFromContext(ctx context.Context) string {
	claims := GetUserClaimsFromContext(ctx)
	if claims == nil {
		return ""
	}
	return claims.Email
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(ctx context.Context) bool {
	return GetUserFromContext(ctx) != nil
}

// HasRole checks if the authenticated user has a specific role
func HasRole(ctx context.Context, role string) bool {
	claims := GetUserClaimsFromContext(ctx)
	if claims == nil {
		return false
	}

	for _, userRole := range claims.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the authenticated user has any of the specified roles
func HasAnyRole(ctx context.Context, roles ...string) bool {
	claims := GetUserClaimsFromContext(ctx)
	if claims == nil {
		return false
	}

	userRoleMap := make(map[string]bool)
	for _, userRole := range claims.Roles {
		userRoleMap[userRole] = true
	}

	for _, role := range roles {
		if userRoleMap[role] {
			return true
		}
	}
	return false
}
