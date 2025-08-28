package event

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EventService provides business logic for event management
type EventService struct {
	repo Repository
}

// NewEventService creates a new event service
func NewEventService(repo Repository) *EventService {
	return &EventService{
		repo: repo,
	}
}

// CreateEvent creates a new event with business validation
func (s *EventService) CreateEvent(ctx context.Context, organizerID string, input CreateEventInput) (*Event, error) {
	// Validate business rules
	if err := s.validateCreateEventInput(input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate unique ID and slug
	eventID := uuid.New().String()
	slug := generateSlug(input.Title)

	// Generate share URL
	shareURL := fmt.Sprintf("/events/%s", slug)

	// Build event object
	event := &Event{
		ID:               eventID,
		Title:            input.Title,
		Description:      input.Description,
		ShortDescription: input.ShortDescription,
		OrganizerID:      organizerID,
		Status:           EventStatusDraft,
		StartTime:        input.StartTime,
		EndTime:          input.EndTime,
		Location: EventLocation{
			Name:         input.Location.Name,
			Address:      input.Location.Address,
			City:         input.Location.City,
			State:        input.Location.State,
			Country:      input.Location.Country,
			ZipCode:      input.Location.ZipCode,
			Instructions: input.Location.Instructions,
			IsRemote:     input.Location.IsRemote,
		},
		Capacity: EventCapacity{
			Minimum:         input.Capacity.Minimum,
			Maximum:         input.Capacity.Maximum,
			WaitlistEnabled: input.Capacity.WaitlistEnabled,
			Current:         0,
		},
		Category:       input.Category,
		TimeCommitment: input.TimeCommitment,
		Tags:           input.Tags,
		Slug:           &slug,
		ShareURL:       &shareURL,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	// Handle requirements if provided
	if input.Requirements != nil {
		event.Requirements = EventRequirements{
			MinimumAge:           input.Requirements.MinimumAge,
			BackgroundCheck:      input.Requirements.BackgroundCheck,
			PhysicalRequirements: input.Requirements.PhysicalRequirements,
			Skills:               []SkillRequirement{},
			Training:             []TrainingRequirement{},
			Interests:            []string{},
		}

		// Convert skill requirements
		for _, skill := range input.Requirements.Skills {
			event.Requirements.Skills = append(event.Requirements.Skills, SkillRequirement{
				Skill:       skill.Skill,
				Proficiency: skill.Proficiency,
				Required:    skill.Required,
			})
		}

		// Convert training requirements
		for _, training := range input.Requirements.Training {
			event.Requirements.Training = append(event.Requirements.Training, TrainingRequirement{
				Name:                training.Name,
				Description:         training.Description,
				Required:            training.Required,
				ProvidedByOrganizer: training.ProvidedByOrganizer,
			})
		}

		event.Requirements.Interests = input.Requirements.Interests
	}

	// Handle coordinates if provided
	if input.Location.Coordinates != nil {
		event.Location.Coordinates = &Coordinates{
			Latitude:  input.Location.Coordinates.Latitude,
			Longitude: input.Location.Coordinates.Longitude,
		}
	}

	// Handle registration settings
	event.RegistrationSettings = RegistrationSettings{
		OpensAt:              input.RegistrationSettings.OpensAt,
		ClosesAt:             input.RegistrationSettings.ClosesAt,
		RequiresApproval:     input.RegistrationSettings.RequiresApproval,
		AllowWaitlist:        input.RegistrationSettings.AllowWaitlist,
		ConfirmationRequired: input.RegistrationSettings.ConfirmationRequired,
		CancellationDeadline: input.RegistrationSettings.CancellationDeadline,
	}

	// Handle recurrence rule if provided
	if input.RecurrenceRule != nil {
		event.RecurrenceRule = &RecurrenceRule{
			Frequency:       input.RecurrenceRule.Frequency,
			Interval:        input.RecurrenceRule.Interval,
			DaysOfWeek:      input.RecurrenceRule.DaysOfWeek,
			DayOfMonth:      input.RecurrenceRule.DayOfMonth,
			EndDate:         input.RecurrenceRule.EndDate,
			OccurrenceCount: input.RecurrenceRule.OccurrenceCount,
		}
	}

	// Create event in repository
	if err := s.repo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return event, nil
}

// GetEvent retrieves an event by ID
func (s *EventService) GetEvent(ctx context.Context, eventID string) (*Event, error) {
	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return event, nil
}

// UpdateEvent updates an existing event
func (s *EventService) UpdateEvent(ctx context.Context, eventID string, userID string, input UpdateEventInput) (*Event, error) {
	// Get existing event
	existingEvent, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Check permissions
	if existingEvent.OrganizerID != userID {
		return nil, fmt.Errorf("unauthorized: user is not the organizer")
	}

	// Create updated event
	updatedEvent := *existingEvent
	updatedEvent.UpdatedAt = time.Now().UTC()

	// Update fields if provided
	if input.Title != nil {
		updatedEvent.Title = *input.Title
	}
	if input.Description != nil {
		updatedEvent.Description = *input.Description
	}
	if input.ShortDescription != nil {
		updatedEvent.ShortDescription = input.ShortDescription
	}
	if input.Category != nil {
		updatedEvent.Category = *input.Category
	}
	if len(input.Tags) > 0 {
		updatedEvent.Tags = input.Tags
	}

	// Update location if provided
	if input.Location != nil {
		updatedEvent.Location.Name = input.Location.Name
		updatedEvent.Location.Address = input.Location.Address
		updatedEvent.Location.City = input.Location.City
		updatedEvent.Location.State = input.Location.State
		updatedEvent.Location.Country = input.Location.Country
		updatedEvent.Location.ZipCode = input.Location.ZipCode
		updatedEvent.Location.Instructions = input.Location.Instructions
		updatedEvent.Location.IsRemote = input.Location.IsRemote

		if input.Location.Coordinates != nil {
			updatedEvent.Location.Coordinates = &Coordinates{
				Latitude:  input.Location.Coordinates.Latitude,
				Longitude: input.Location.Coordinates.Longitude,
			}
		}
	}

	// Update requirements if provided
	if input.Requirements != nil {
		updatedEvent.Requirements = EventRequirements{
			MinimumAge:           input.Requirements.MinimumAge,
			BackgroundCheck:      input.Requirements.BackgroundCheck,
			PhysicalRequirements: input.Requirements.PhysicalRequirements,
			Skills:               []SkillRequirement{},
			Training:             []TrainingRequirement{},
			Interests:            []string{},
		}

		// Convert skill requirements
		for _, skill := range input.Requirements.Skills {
			updatedEvent.Requirements.Skills = append(updatedEvent.Requirements.Skills, SkillRequirement{
				Skill:       skill.Skill,
				Proficiency: skill.Proficiency,
				Required:    skill.Required,
			})
		}

		// Convert training requirements
		for _, training := range input.Requirements.Training {
			updatedEvent.Requirements.Training = append(updatedEvent.Requirements.Training, TrainingRequirement{
				Name:                training.Name,
				Description:         training.Description,
				Required:            training.Required,
				ProvidedByOrganizer: training.ProvidedByOrganizer,
			})
		}

		updatedEvent.Requirements.Interests = input.Requirements.Interests
	}

	// Validate the updated event
	if err := s.validateEventUpdate(ctx, &updatedEvent, existingEvent); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update in repository
	if err := s.repo.Update(ctx, &updatedEvent); err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return &updatedEvent, nil
}

// PublishEvent publishes a draft event
func (s *EventService) PublishEvent(ctx context.Context, eventID string, userID string) (*Event, error) {
	// Get existing event
	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Check permissions
	if event.OrganizerID != userID {
		return nil, fmt.Errorf("unauthorized: user is not the organizer")
	}

	// Check if event is in draft status
	if event.Status != EventStatusDraft {
		return nil, fmt.Errorf("event is not in draft status")
	}

	// Validate event for publishing
	if err := s.validateEventForPublishing(ctx, event); err != nil {
		return nil, fmt.Errorf("event validation failed: %w", err)
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, eventID, EventStatusPublished); err != nil {
		return nil, fmt.Errorf("failed to publish event: %w", err)
	}

	// Get updated event
	publishedEvent, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get published event: %w", err)
	}

	return publishedEvent, nil
}

// CancelEvent cancels an event
func (s *EventService) CancelEvent(ctx context.Context, eventID string, userID string, reason string) (*Event, error) {
	// Get existing event
	event, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Check permissions
	if event.OrganizerID != userID {
		return nil, fmt.Errorf("unauthorized: user is not the organizer")
	}

	// Check if event can be cancelled
	if event.Status == EventStatusCancelled || event.Status == EventStatusCompleted {
		return nil, fmt.Errorf("event cannot be cancelled in current status: %s", event.Status)
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, eventID, EventStatusCancelled); err != nil {
		return nil, fmt.Errorf("failed to cancel event: %w", err)
	}

	// Get updated event
	cancelledEvent, err := s.repo.GetByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cancelled event: %w", err)
	}

	return cancelledEvent, nil
}

// Validation functions

func (s *EventService) validateCreateEventInput(input CreateEventInput) error {
	// Validate event times
	if err := validateEventTimes(input.StartTime, input.EndTime); err != nil {
		return err
	}

	// Validate capacity
	if err := s.validateCapacity(input.Capacity); err != nil {
		return err
	}

	// Validate registration settings
	if err := s.validateRegistrationSettings(input.RegistrationSettings, input.StartTime); err != nil {
		return err
	}

	return nil
}

func (s *EventService) validateEventUpdate(ctx context.Context, updatedEvent *Event, originalEvent *Event) error {
	// Check if critical changes are allowed
	if updatedEvent.Status == EventStatusPublished {
		// Only allow certain fields to be updated for published events
		if updatedEvent.StartTime != originalEvent.StartTime {
			return fmt.Errorf("cannot change start time for published event")
		}
		if updatedEvent.EndTime != originalEvent.EndTime {
			return fmt.Errorf("cannot change end time for published event")
		}
	}

	return nil
}

func (s *EventService) validateEventForPublishing(ctx context.Context, event *Event) error {
	// Check if event has all required fields
	if event.Title == "" {
		return fmt.Errorf("title is required")
	}
	if event.Description == "" {
		return fmt.Errorf("description is required")
	}
	if event.Location.Name == "" {
		return fmt.Errorf("location name is required")
	}
	if event.Capacity.Maximum <= 0 {
		return fmt.Errorf("maximum capacity must be greater than 0")
	}

	// Validate times
	if err := validateEventTimes(event.StartTime, event.EndTime); err != nil {
		return err
	}

	return nil
}

func validateEventTimes(startTime, endTime time.Time) error {
	now := time.Now().UTC()

	if startTime.Before(now) {
		return fmt.Errorf("start time cannot be in the past")
	}

	if endTime.Before(startTime) {
		return fmt.Errorf("end time cannot be before start time")
	}

	if endTime.Sub(startTime) < 30*time.Minute {
		return fmt.Errorf("event duration must be at least 30 minutes")
	}

	return nil
}

func (s *EventService) validateCapacity(capacity EventCapacityInput) error {
	if capacity.Maximum <= 0 {
		return fmt.Errorf("maximum capacity must be greater than 0")
	}

	if capacity.Minimum < 0 {
		return fmt.Errorf("minimum capacity cannot be negative")
	}

	if capacity.Minimum > capacity.Maximum {
		return fmt.Errorf("minimum capacity cannot be greater than maximum capacity")
	}

	return nil
}

func (s *EventService) validateRegistrationSettings(settings RegistrationSettingsInput, eventStartTime time.Time) error {
	now := time.Now().UTC()

	if settings.OpensAt != nil && settings.OpensAt.Before(now) {
		return fmt.Errorf("registration open time cannot be in the past")
	}

	if settings.ClosesAt.Before(now) {
		return fmt.Errorf("registration close time cannot be in the past")
	}

	if settings.OpensAt != nil && settings.ClosesAt.Before(*settings.OpensAt) {
		return fmt.Errorf("registration close time cannot be before open time")
	}

	if settings.ClosesAt.After(eventStartTime) {
		return fmt.Errorf("registration must close before event starts")
	}

	if settings.CancellationDeadline != nil {
		if settings.CancellationDeadline.Before(now) {
			return fmt.Errorf("cancellation deadline cannot be in the past")
		}
		if settings.CancellationDeadline.After(eventStartTime) {
			return fmt.Errorf("cancellation deadline must be before event starts")
		}
	}

	return nil
}

// GetEventByID retrieves an event by its ID
func (s *EventService) GetEventByID(ctx context.Context, eventID string) (*Event, error) {
	return s.repo.GetByID(ctx, eventID)
}

// GetEventBySlug retrieves an event by its slug
func (s *EventService) GetEventBySlug(ctx context.Context, slug string) (*Event, error) {
	return s.repo.GetBySlug(ctx, slug)
}

// SearchEvents searches for events with filters, sorting, and pagination
func (s *EventService) SearchEvents(ctx context.Context, filter EventSearchFilter, sort *EventSortInput, limit, offset int) (*EventConnection, error) {
	return s.repo.List(ctx, filter, sort, limit, offset)
}

// GetUserEvents retrieves events for a specific user
func (s *EventService) GetUserEvents(ctx context.Context, userID string, statuses []EventStatus, limit, offset int) (*EventConnection, error) {
	events, err := s.repo.GetByOrganizer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user events: %w", err)
	}

	// Filter by status if provided
	var filteredEvents []*Event
	if len(statuses) > 0 {
		statusMap := make(map[EventStatus]bool)
		for _, status := range statuses {
			statusMap[status] = true
		}

		for _, event := range events {
			if statusMap[event.Status] {
				filteredEvents = append(filteredEvents, event)
			}
		}
	} else {
		filteredEvents = events
	}

	// Apply pagination
	start := offset
	end := offset + limit
	if start >= len(filteredEvents) {
		filteredEvents = []*Event{}
	} else {
		if end > len(filteredEvents) {
			end = len(filteredEvents)
		}
		filteredEvents = filteredEvents[start:end]
	}

	edges := make([]EventEdge, len(filteredEvents))
	for i, event := range filteredEvents {
		edges[i] = EventEdge{
			Node:   *event,
			Cursor: fmt.Sprintf("%d", offset+i),
		}
	}

	return &EventConnection{
		Edges:      edges,
		PageInfo:   PageInfo{HasNextPage: end < len(events), HasPreviousPage: start > 0},
		TotalCount: len(events),
	}, nil
}

// GetNearbyEvents retrieves events near a specific location
func (s *EventService) GetNearbyEvents(ctx context.Context, lat, lng, radius float64, filter EventSearchFilter, limit, offset int) (*EventConnection, error) {
	events, err := s.repo.GetNearby(ctx, lat, lng, radius, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get nearby events: %w", err)
	}

	// Apply offset
	start := offset
	if start >= len(events) {
		events = []*Event{}
	} else {
		end := offset + limit
		if end > len(events) {
			end = len(events)
		}
		events = events[start:end]
	}

	edges := make([]EventEdge, len(events))
	for i, event := range events {
		edges[i] = EventEdge{
			Node:   *event,
			Cursor: fmt.Sprintf("%d", offset+i),
		}
	}

	return &EventConnection{
		Edges:      edges,
		PageInfo:   PageInfo{HasNextPage: false, HasPreviousPage: offset > 0},
		TotalCount: len(events),
	}, nil
} // Helper functions

func generateSlug(title string) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters except hyphens
	result := ""
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result += string(char)
		}
	}
	// Remove consecutive hyphens and trim
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	result = strings.Trim(result, "-")

	if len(result) > 50 {
		result = result[:50]
		result = strings.Trim(result, "-")
	}

	return result
}
