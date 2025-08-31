# VolunteerSync Backend Agents

This file defines the Copilot agents for the VolunteerSync backend project.

## @general
- **Overall Goal**: Implement a robust, scalable, and secure backend for a volunteer management platform.
- **Tech Stack**: Go, Gin (for HTTP routing), gqlgen (for GraphQL), PostgreSQL (for database).
- **Architecture**: Modular monolith. Keep a clear separation of concerns:
  - `internal/graph`: GraphQL resolvers and schema. Resolvers should be thin and delegate business logic to core services.
  - `internal/core`: Core business logic services (e.g., `AuthService`, `EventService`). This is where the main application logic resides.
  - `internal/store`: Database interaction layer (repositories). Handles all SQL queries and data persistence.
  - `cmd/api`: Application entry point.
- **Coding Style**:
  - Follow Go best practices and conventions.
  - Use meaningful variable and function names.
  - Implement comprehensive error handling. Return specific errors where possible.
  - Write unit tests for all business logic and integration tests for GraphQL resolvers. Aim for >90% test coverage.
- **Database**:
  - Use UUIDs for all primary keys.
  - Use `TIMESTAMPTZ` for all date/time fields.
  - Use the `migrate` tool for schema migrations.
- **Security**:
  - Validate all inputs.
  - Use parameterized queries to prevent SQL injection.
  - Implement role-based access control (RBAC) in services and middleware.

---

## @auth-agent
- **Scope**: `internal/core/auth/`, `internal/store/postgres/auth_repo.go`, `internal/middleware/`
- **Purpose**: Manages all aspects of user authentication and authorization.
- **Key Concepts**:
  - **JWT**: Use JWTs for stateless authentication. Generate access tokens (15-min expiry) and refresh tokens (7-day expiry).
  - **Password Hashing**: Use `bcrypt` with a cost factor of at least 12.
  - **OAuth**: Integrate with Google OAuth for social login.
  - **Services**: The `AuthService` contains the primary logic for registration, login, and token management. The `JWTService` handles token creation and validation. The `PasswordService` handles hashing and verification.
- **Instructions**: When working in this scope, focus on security best practices for authentication. Ensure token claims are minimal and necessary. Handle token revocation and refresh token rotation correctly. Refer to `phase-2-authentication.md` for detailed requirements.

---

## @user-agent
- **Scope**: `internal/core/user/`, `internal/store/postgres/user_store.go`
- **Purpose**: Manages user profiles, preferences, and related data.
- **Key Concepts**:
  - **User Profile**: Includes user's name, bio, location, interests, and skills.
  - **Privacy**: Implement privacy settings to control profile visibility.
  - **File Uploads**: Handle profile picture uploads securely, storing them in a designated file service.
  - **RBAC**: Manage user roles (VOLUNTEER, ORGANIZER, ADMIN).
- **Instructions**: When working in this scope, focus on creating a comprehensive user management system. Ensure privacy settings are respected when fetching user data. Refer to `phase-3-user-management.md` for detailed requirements.

---

## @event-agent
- **Scope**: `internal/core/event/`, `internal/store/postgres/event_store.go`
- **Purpose**: Manages the entire lifecycle of volunteer events.
- **Key Concepts**:
  - **Event Lifecycle**: Events have a status (DRAFT, PUBLISHED, CANCELLED, COMPLETED, ARCHIVED).
  - **Event Model**: Events have detailed information including location, capacity, requirements, and recurrence rules.
  - **Search**: Implement location-based search and filtering by various criteria.
- **Instructions**: When working in this scope, focus on the event data model and its lifecycle. Ensure that event creation, updates, and publishing logic are robust. Refer to `phase-4-event-management.md` for detailed requirements.

---

## @graphql-agent
- **Scope**: `internal/graph/`
- **Purpose**: Defines and resolves the GraphQL API.
- **Key Concepts**:
  - **Schema-First**: The API is defined in `schema.graphqls`. Use `gqlgen` to generate models and resolver stubs.
  - **Thin Resolvers**: Resolvers should be lightweight. They translate the GraphQL request and call the appropriate core service to handle the business logic.
  - **Dataloaders**: Use dataloaders to prevent N+1 query problems, especially for nested data (e.g., fetching the organizer for a list of events).
- **Instructions**: When modifying the GraphQL layer, always start with the schema. Ensure resolvers handle errors gracefully and delegate complex logic to the `internal/core` services.
