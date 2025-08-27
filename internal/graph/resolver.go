package graph

import (
	"database/sql"

	"github.com/volunteersync/backend/internal/core/auth"
	usercore "github.com/volunteersync/backend/internal/core/user"
)

// Resolver serves as dependency injection for your app, add any dependencies you need here.
type Resolver struct {
	DB           *sql.DB
	AuthService  *auth.AuthService
	OAuthService *auth.OAuthService
	UserService  *usercore.Service
}
