# @event-agent
- **Scope**: This directory (`internal/core/event/`)
- **Purpose**: This agent is responsible for the core business logic related to event management.

## Core Responsibilities
- **Event Lifecycle Management**: Handle the creation, updating, publishing, cancellation, and archival of events.
- **Event Data Management**: Manage all data associated with an event, including location, capacity, requirements, tags, and images.
- **Search and Discovery**: Provide logic for searching, filtering, and discovering events based on various criteria like location, date, skills, and interests.
- **Recurring Events**: Implement the logic for creating and managing recurring event series.

## Key Files & Services
- **`service.go`**: The `EventService` contains the primary business logic for all event-related operations. It validates inputs, orchestrates repository calls, and enforces business rules (e.g., an organizer can't change the date of a published event).
- **`repository.go`**: Defines the `Repository` interface for all event-related database operations. This abstracts the data persistence layer.
- **`models.go`**: Contains the comprehensive data structures for the event domain, including `Event`, `EventLocation`, `EventCapacity`, `RecurrenceRule`, and various input/filter/sort models.

## Technical Guidelines
- **Rich Data Model**: The `Event` model is complex. When working with it, ensure all related entities (like `Requirements`, `Location`, `Capacity`) are handled correctly.
- **Business Logic Validation**: The `EventService` is the gatekeeper for all business rules. Perform thorough validation here before passing data to the repository. For example, validate that `EndTime` is after `StartTime`.
- **Slug Generation**: Implement a consistent and robust method for generating URL-friendly slugs from event titles. Ensure uniqueness.
- **Geospatial Queries**: For location-based search, ensure the repository interface supports geospatial queries (e.g., finding events within a certain radius).
- **Testing**: Write unit tests for the `EventService`, mocking the repository to test business logic in isolation. Cover all aspects of the event lifecycle and validation rules.

## Development Context
- This module corresponds to **Phase 4** of the project plan. For detailed requirements, user stories, and BDD/TDD scenarios, refer to the `.github/prompts/phase-4-event-management.md` file.
