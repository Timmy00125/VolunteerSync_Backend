package graph

import (
	"database/sql"

	"github.com/volunteersync/backend/internal/core/auth"
	"github.com/volunteersync/backend/internal/core/event"
	usercore "github.com/volunteersync/backend/internal/core/user"
)

// Resolver serves as dependency injection for your app, add any dependencies you need here.
type Resolver struct {
	DB           *sql.DB
	AuthService  *auth.AuthService
	OAuthService *auth.OAuthService
	UserService  *usercore.Service
	EventService *event.EventService
}
