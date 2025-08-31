# @registration-agent
- **Scope**: This directory (`internal/core/registration/`)
- **Purpose**: This agent is responsible for the core business logic related to event registration, waitlisting, and attendance tracking.

## Core Responsibilities
- **Event Registration**: Handle a volunteer's registration for an event, including validation against event requirements and capacity.
- **Waitlist Management**: Automatically manage a waitlist for full events, including prioritizing and promoting volunteers when spots become available.
- **Approval Workflow**: Facilitate the process for organizers to approve or decline registrations for events that require it.
- **Status Management**: Track the full lifecycle of a registration, from pending to confirmed, waitlisted, cancelled, or completed.
- **Attendance Tracking**: Manage volunteer check-in and track attendance status (e.g., attended, no-show).
- **Conflict Detection**: Identify and alert users to potential scheduling conflicts when they register for multiple events.

## Key Files & Services
- **`RegistrationService`**: The main service that orchestrates the entire registration process. It will likely depend on other services.
- **`WaitlistService`**: A dedicated service to handle all logic related to waitlists, including adding, removing, and promoting volunteers.
- **`ApprovalService`**: Manages the logic for events requiring organizer approval for registrations.
- **`ConflictService`**: Implements the logic to detect scheduling conflicts for volunteers.
- **Repository Interface**: Defines the contract for all registration-related database operations, such as creating registrations, updating statuses, and managing waitlist entries.
- **Models**: Contains the data structures for this domain, including `Registration`, `WaitlistEntry`, `AttendanceRecord`, and `RegistrationConflict`.

## Technical Guidelines
- **Transactional Integrity**: Registration is a critical workflow. Ensure that operations that modify multiple tables (e.g., creating a registration and updating event capacity) are performed within a database transaction to maintain data consistency.
- **State Machine**: The `RegistrationStatus` and `AttendanceStatus` fields act as a state machine. Ensure all status transitions are valid and that business logic correctly handles each state.
- **Asynchronous Operations**: For tasks like sending notifications or promoting from a waitlist, consider using background jobs or a queue to avoid blocking the main request thread and to improve system resilience.
- **Intelligent Algorithms**: The waitlist promotion logic should be designed to be fair and effective, potentially considering factors beyond just FIFO, such as skill matching or volunteer reliability.
- **Testing**: Rigorously test all registration scenarios, including race conditions (e.g., two users trying to register for the last spot simultaneously), waitlist promotions, and all possible status transitions.

## Development Context
- This module corresponds to **Phase 5** of the project plan. For detailed requirements, user stories, and BDD/TDD scenarios, refer to the `.github/prompts/phase-5-registration-system.md` file.
