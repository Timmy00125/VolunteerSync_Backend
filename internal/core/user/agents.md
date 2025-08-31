# @user-agent
- **Scope**: This directory (`internal/core/user/`)
- **Purpose**: This agent is responsible for managing user profiles, settings, and related data.

## Core Responsibilities
- **Profile Management**: Handle CRUD operations for user profiles, including bio, location, and contact information.
- **Interest and Skill Management**: Allow users to add, update, and remove their skills and interests.
- **Settings Management**: Manage user-specific settings, including privacy controls and notification preferences.
- **File Uploads**: Handle the logic for uploading and managing user profile pictures, coordinating with a file storage service.

## Key Files & Services
- **`service.go`**: The `UserService` contains the primary business logic for managing user profiles. It orchestrates calls to the `UserStore` and `FileService`.
- **`files_local.go`**: An implementation of the `FileService` interface that stores files on the local filesystem. This is used for profile picture uploads.
- **`models.go`**: Contains the data structures for the user domain, such as `UserProfile`, `Skill`, `Interest`, and various settings models.
- **The `UserStore` interface (defined in `service.go`)**: This interface defines the contract for all user-related database operations, abstracting the data persistence layer.

## Technical Guidelines
- **Privacy by Design**: Always filter user profile data based on the user's privacy settings and the requester's relationship to the user. The `filterProfileByPrivacy` helper function is critical for this.
- **Separation of Concerns**: The `UserService` should not contain file storage or database query logic directly. It should delegate these tasks to the `FileService` and `UserStore` interfaces, respectively.
- **Input Validation**: Thoroughly validate all input for updating profiles, skills, and settings to ensure data integrity.
- **Audit Logging**: Important actions like updating a profile or changing settings should be logged for security and auditing purposes via the `AuditLogger` interface.
- **Testing**: Write unit tests for the `UserService`, mocking the `UserStore`, `FileService`, and other dependencies to test the business logic in isolation.

## Development Context
- This module corresponds to **Phase 3** of the project plan. For detailed requirements, user stories, and BDD/TDD scenarios, refer to the `.github/prompts/phase-3-user-management.md` file.
