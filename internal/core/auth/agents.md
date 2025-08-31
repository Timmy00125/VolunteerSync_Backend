# @auth-agent
- **Scope**: This directory (`internal/core/auth/`)
- **Purpose**: This agent is responsible for the core logic of user authentication and authorization in the VolunteerSync platform.

## Core Responsibilities
- **User Registration**: Handle new user sign-ups, including password hashing and initial token generation.
- **User Login**: Authenticate users with email/password and Google OAuth.
- **Token Management**: Generate, validate, and refresh JWT access and refresh tokens.
- **Password Security**: Securely hash and verify passwords using `bcrypt`.
- **Authorization**: Provide mechanisms to check user roles and permissions.

## Key Files & Services
- **`auth_service.go`**: The primary service orchestrating all authentication logic. It coordinates user repository interactions, password verification, and token generation.
- **`jwt_service.go`**: A dedicated service for all JWT-related operations. It handles the creation of token pairs (`AccessToken`, `RefreshToken`) and the validation of incoming tokens.
- **`password_service.go`**: A service focused on password security. It uses `bcrypt` to hash new passwords and verify existing ones. It also includes password strength validation.
- **`oauth_service.go`**: Handles the Google OAuth2 flow, from generating the auth URL to processing the callback and creating/linking user accounts.
- **`repository.go`**: Defines the interfaces (`UserRepository`, `RefreshTokenRepository`) for database interactions, abstracting the data layer from the service logic.
- **`models.go`**: Contains the data structures for this domain, such as `User`, `RefreshToken`, and request/response models.

## Technical Guidelines
- **Security First**: Always prioritize security. Sanitize inputs, use constant-time comparison for password verification, and handle tokens with care.
- **Statelessness**: The authentication mechanism is JWT-based and should remain stateless. All necessary user information for a request should be encoded in the access token claims.
- **Error Handling**: Return specific, meaningful errors (e.g., `ErrInvalidCredentials`, `ErrUserNotFound`, `ErrAccountLocked`) to be handled by the GraphQL layer.
- **Testing**: Write thorough unit tests for all services. Mock the repository interfaces to test the service logic in isolation. Pay special attention to testing edge cases in authentication flows (e.g., token expiration, account lockout, invalid inputs).

## Development Context
- This module corresponds to **Phase 2** of the project plan. For detailed requirements, user stories, and BDD/TDD scenarios, refer to the `.github/prompts/phase-2-authentication.md` file.
