package graph

import (
	"database/sql"

	"github.com/volunteersync/backend/internal/core/auth"
	"github.com/volunteersync/backend/internal/core/event"
	"github.com/volunteersync/backend/internal/core/registration"
	usercore "github.com/volunteersync/backend/internal/core/user"
	"github.com/volunteersync/backend/internal/graph/generated"
)

// Resolver serves as dependency injection for your app, add any dependencies you need here.
type Resolver struct {
	DB                  *sql.DB
	AuthService         *auth.AuthService
	OAuthService        *auth.OAuthService
	UserService         *usercore.Service
	EventService        *event.EventService
	RegistrationService *registration.Service
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Event returns generated.EventResolver implementation.
func (r *Resolver) Event() generated.EventResolver { return &eventResolver{r} }

// PublicProfile returns generated.PublicProfileResolver implementation.
func (r *Resolver) PublicProfile() generated.PublicProfileResolver { return &publicProfileResolver{r} }

// Registration returns generated.RegistrationResolver implementation.
func (r *Resolver) Registration() generated.RegistrationResolver { return &registrationResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }
