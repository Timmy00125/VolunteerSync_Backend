package registration

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/volunteersync/backend/internal/core/event"
	"github.com/volunteersync/backend/internal/core/user"
)

// Service encapsulates the business logic for registrations.
type Service struct {
	repo         Repository
	eventService *event.EventService
	userService  *user.Service
	logger       *slog.Logger
}

// NewService creates a new registration service.
func NewService(repo Repository, eventService *event.EventService, userService *user.Service, logger *slog.Logger) *Service {
	if repo == nil {
		panic("registration repository is required")
	}
	if eventService == nil {
		panic("event service is required")
	}
	if userService == nil {
		panic("user service is required")
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Service{
		repo:         repo,
		eventService: eventService,
		userService:  userService,
		logger:       logger,
	}
}

// ApproveRegistration handles approval/decline of registration requests
func (s *Service) ApproveRegistration(ctx context.Context, organizerID, registrationID string, approved bool, notes string) (*Registration, error) {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return nil, fmt.Errorf("registration not found: %w", err)
	}

	// Validate organizer permission
	evt, err := s.eventService.GetEvent(ctx, reg.EventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if evt.OrganizerID != organizerID {
		return nil, fmt.Errorf("user is not the organizer of this event")
	}

	// Update registration status
	if approved {
		if err := s.approveRegistration(ctx, reg, evt, notes); err != nil {
			return nil, err
		}
	} else {
		reg.Status = StatusDeclined
		reg.ApprovalNotes = notes
		reg.UpdatedAt = time.Now()
	}

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		return nil, fmt.Errorf("failed to update registration: %w", err)
	}

	return reg, nil
}

// approveRegistration handles the approval logic including capacity checks
func (s *Service) approveRegistration(ctx context.Context, reg *Registration, evt *event.Event, notes string) error {
	// Check if there's still capacity
	confirmedRegs, err := s.getConfirmedRegistrations(ctx, evt.ID)
	if err != nil {
		return fmt.Errorf("failed to check capacity: %w", err)
	}

	if len(confirmedRegs) >= evt.Capacity.Maximum {
		// No capacity, add to waitlist
		reg.Status = StatusWaitlisted
		waitlistPos, err := s.getNextWaitlistPosition(ctx, evt.ID)
		if err != nil {
			return fmt.Errorf("failed to get waitlist position: %w", err)
		}
		reg.WaitlistPosition = &waitlistPos
	} else {
		// Approve and confirm
		reg.Status = StatusConfirmed
		now := time.Now()
		reg.ConfirmedAt = &now
	}

	reg.ApprovalNotes = notes
	reg.UpdatedAt = time.Now()
	return nil
}

// CancelRegistration handles cancellation of user registrations
func (s *Service) CancelRegistration(ctx context.Context, userID, registrationID, reason string) (*Registration, error) {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return nil, fmt.Errorf("registration not found: %w", err)
	}

	if reg.UserID != userID {
		return nil, fmt.Errorf("user does not have permission to cancel this registration")
	}

	// Update registration status
	reg.Status = StatusCancelled
	reg.CancellationReason = reason
	now := time.Now()
	reg.CancelledAt = &now
	reg.UpdatedAt = now

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		return nil, fmt.Errorf("failed to cancel registration: %w", err)
	}

	// Try to promote someone from waitlist if this was a confirmed registration
	if reg.Status == StatusConfirmed {
		go s.promoteFromWaitlist(context.Background(), reg.EventID)
	}

	return reg, nil
}

// promoteFromWaitlist promotes the next person from waitlist when a spot opens
func (s *Service) promoteFromWaitlist(ctx context.Context, eventID string) {
	waitlistEntries, err := s.repo.GetWaitlistEntriesByEventID(ctx, eventID)
	if err != nil {
		s.logger.Error("failed to get waitlist entries", "error", err)
		return
	}

	if len(waitlistEntries) == 0 {
		return
	}

	// Find the next person to promote (lowest position)
	var nextEntry *WaitlistEntry
	for _, entry := range waitlistEntries {
		if nextEntry == nil || entry.Position < nextEntry.Position {
			nextEntry = entry
		}
	}

	if nextEntry == nil {
		return
	}

	// Get the registration and promote
	reg, err := s.repo.GetRegistrationByID(ctx, nextEntry.RegistrationID)
	if err != nil {
		s.logger.Error("failed to get registration for promotion", "error", err)
		return
	}

	reg.Status = StatusConfirmed
	now := time.Now()
	reg.ConfirmedAt = &now
	reg.WaitlistPromotedAt = &now
	reg.WaitlistPosition = nil
	reg.UpdatedAt = now

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		s.logger.Error("failed to promote registration", "error", err)
		return
	}

	// Remove from waitlist
	if err := s.repo.RemoveWaitlistEntry(ctx, nextEntry.ID); err != nil {
		s.logger.Error("failed to remove waitlist entry", "error", err)
	}
}

// GetRegistrationsByEventID returns all registrations for an event
func (s *Service) GetRegistrationsByEventID(ctx context.Context, eventID string) ([]*Registration, error) {
	return s.repo.GetRegistrationsByEventID(ctx, eventID)
}

// GetRegistrationByID returns a specific registration by ID
func (s *Service) GetRegistrationByID(ctx context.Context, id string) (*Registration, error) {
	return s.repo.GetRegistrationByID(ctx, id)
}

// GetRegistrationsByUserID returns all registrations for a user
func (s *Service) GetRegistrationsByUserID(ctx context.Context, userID string) ([]*Registration, error) {
	return s.repo.GetRegistrationsByUserID(ctx, userID)
}

// CheckInVolunteer handles volunteer check-in for an event
func (s *Service) CheckInVolunteer(ctx context.Context, registrationID, checkedInBy string) (*Registration, error) {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return nil, fmt.Errorf("registration not found: %w", err)
	}

	if reg.Status != StatusConfirmed {
		return nil, fmt.Errorf("registration is not confirmed")
	}

	now := time.Now()
	reg.CheckedInAt = &now
	reg.CheckedInBy = &checkedInBy
	reg.AttendanceStatus = AttendanceCheckedIn
	reg.UpdatedAt = now

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		return nil, fmt.Errorf("failed to check in volunteer: %w", err)
	}

	return reg, nil
}

// MarkEventCompleted marks a registration as completed after the event
func (s *Service) MarkEventCompleted(ctx context.Context, registrationID string) (*Registration, error) {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return nil, fmt.Errorf("registration not found: %w", err)
	}

	if reg.AttendanceStatus != AttendanceCheckedIn {
		reg.AttendanceStatus = AttendanceNoShow
	} else {
		reg.AttendanceStatus = AttendanceCompleted
	}

	reg.Status = StatusCompleted
	now := time.Now()
	reg.CompletedAt = &now
	reg.UpdatedAt = now

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		return nil, fmt.Errorf("failed to mark registration completed: %w", err)
	}

	return reg, nil
}

// GetWaitlistByEventID returns waitlist entries for an event
func (s *Service) GetWaitlistByEventID(ctx context.Context, eventID string) ([]*WaitlistEntry, error) {
	return s.repo.GetWaitlistEntriesByEventID(ctx, eventID)
}

// BulkRegister handles registration for multiple events
func (s *Service) BulkRegister(ctx context.Context, userID string, eventIDs []string, personalMessage string, skipConflicts bool) ([]*Registration, error) {
	var registrations []*Registration
	var errors []error

	for _, eventID := range eventIDs {
		registration, err := s.RegisterForEvent(ctx, userID, eventID, personalMessage)
		if err != nil {
			if skipConflicts {
				// Log error but continue with other registrations
				s.logger.Warn("failed to register for event", "eventID", eventID, "error", err)
				errors = append(errors, err)
				continue
			}
			return nil, fmt.Errorf("failed to register for event %s: %w", eventID, err)
		}
		registrations = append(registrations, registration)
	}

	if len(errors) > 0 && len(registrations) == 0 {
		return nil, fmt.Errorf("all registrations failed")
	}

	return registrations, nil
}

// PromoteFromWaitlist manually promotes a specific registration from waitlist
func (s *Service) PromoteFromWaitlist(ctx context.Context, registrationID string) (*Registration, error) {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return nil, fmt.Errorf("registration not found: %w", err)
	}

	if reg.Status != StatusWaitlisted {
		return nil, fmt.Errorf("registration is not on waitlist")
	}

	// Check if there's capacity
	evt, err := s.eventService.GetEvent(ctx, reg.EventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	confirmedCount, err := s.getConfirmedRegistrationCount(ctx, evt.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check capacity: %w", err)
	}

	if confirmedCount >= evt.Capacity.Maximum {
		return nil, fmt.Errorf("event is at maximum capacity")
	}

	// Promote the registration
	reg.Status = StatusConfirmed
	now := time.Now()
	reg.ConfirmedAt = &now
	reg.WaitlistPromotedAt = &now
	reg.WaitlistPosition = nil
	reg.UpdatedAt = now

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		return nil, fmt.Errorf("failed to promote registration: %w", err)
	}

	// Remove waitlist entry if it exists
	waitlistEntry, err := s.repo.GetWaitlistEntryByRegistrationID(ctx, registrationID)
	if err == nil && waitlistEntry != nil {
		if err := s.repo.RemoveWaitlistEntry(ctx, waitlistEntry.ID); err != nil {
			s.logger.Warn("failed to remove waitlist entry", "error", err)
		}
	}

	return reg, nil
}

// GetRegistrationStats returns statistics for an event's registrations
func (s *Service) GetRegistrationStats(ctx context.Context, eventID string) (*RegistrationStats, error) {
	registrations, err := s.repo.GetRegistrationsByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get registrations: %w", err)
	}

	stats := &RegistrationStats{
		TotalRegistrations: len(registrations),
	}

	var confirmed, waitlisted, cancelled, completed, checkedIn int
	for _, reg := range registrations {
		switch reg.Status {
		case StatusConfirmed:
			confirmed++
		case StatusWaitlisted:
			waitlisted++
		case StatusCancelled:
			cancelled++
		case StatusCompleted:
			completed++
		}

		if reg.AttendanceStatus == AttendanceCheckedIn || reg.AttendanceStatus == AttendanceCompleted {
			checkedIn++
		}
	}

	stats.ConfirmedRegistrations = confirmed
	stats.WaitlistCount = waitlisted

	if stats.TotalRegistrations > 0 {
		stats.CancellationRate = float64(cancelled) / float64(stats.TotalRegistrations)
		if confirmed > 0 {
			stats.AttendanceRate = float64(checkedIn) / float64(confirmed)
			stats.NoShowRate = float64(confirmed-checkedIn) / float64(confirmed)
		}
	}

	return stats, nil
}

// RegistrationStats represents registration statistics for an event
type RegistrationStats struct {
	TotalRegistrations     int     `json:"totalRegistrations"`
	ConfirmedRegistrations int     `json:"confirmedRegistrations"`
	WaitlistCount          int     `json:"waitlistCount"`
	AttendanceRate         float64 `json:"attendanceRate"`
	NoShowRate             float64 `json:"noShowRate"`
	CancellationRate       float64 `json:"cancellationRate"`
}

// RegisterForEvent handles the registration of a user for an event.
func (s *Service) RegisterForEvent(ctx context.Context, userID, eventID, personalMessage string) (*Registration, error) {
	// Validate inputs
	if err := s.validateRegistrationInputs(ctx, userID, eventID); err != nil {
		return nil, err
	}

	// Create base registration
	registration := &Registration{
		ID:               uuid.New().String(),
		UserID:           userID,
		EventID:          eventID,
		PersonalMessage:  personalMessage,
		AttendanceStatus: AttendanceRegistered,
		AppliedAt:        time.Now(),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Set registration status and save
	return s.processRegistration(ctx, registration)
}

// validateRegistrationInputs performs all validation checks
func (s *Service) validateRegistrationInputs(ctx context.Context, userID, eventID string) error {
	if err := s.validateUser(ctx, userID); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	if _, err := s.validateEvent(ctx, eventID); err != nil {
		return fmt.Errorf("event validation failed: %w", err)
	}

	if err := s.checkDuplicateRegistration(ctx, userID, eventID); err != nil {
		return err
	}

	return nil
}

// processRegistration determines status and saves the registration
func (s *Service) processRegistration(ctx context.Context, registration *Registration) (*Registration, error) {
	evt, err := s.eventService.GetEvent(ctx, registration.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if err := s.setRegistrationStatus(ctx, registration, evt); err != nil {
		return nil, fmt.Errorf("failed to set registration status: %w", err)
	}

	return s.repo.CreateRegistration(ctx, registration)
}

// validateUser checks if the user exists and is active
func (s *Service) validateUser(ctx context.Context, userID string) error {
	profile, err := s.userService.GetProfile(ctx, userID, userID, nil)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if profile == nil {
		return fmt.Errorf("user not found")
	}
	return nil
}

// validateEvent checks if the event exists and is available for registration
func (s *Service) validateEvent(ctx context.Context, eventID string) (*event.Event, error) {
	evt, err := s.eventService.GetEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}
	if evt == nil {
		return nil, fmt.Errorf("event not found")
	}

	// Check if event is available for registration
	if err := s.checkEventAvailability(evt); err != nil {
		return nil, err
	}

	return evt, nil
}

// checkEventAvailability validates event status and deadlines
func (s *Service) checkEventAvailability(evt *event.Event) error {
	if evt.Status != event.EventStatusPublished {
		return fmt.Errorf("event is not available for registration")
	}

	if time.Now().After(evt.RegistrationSettings.ClosesAt) {
		return fmt.Errorf("registration deadline has passed")
	}

	return nil
}

// checkDuplicateRegistration ensures user hasn't already registered for this event
func (s *Service) checkDuplicateRegistration(ctx context.Context, userID, eventID string) error {
	registrations, err := s.repo.GetRegistrationsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to check existing registrations: %w", err)
	}

	for _, reg := range registrations {
		if reg.EventID == eventID && reg.Status != StatusCancelled && reg.Status != StatusDeclined {
			return fmt.Errorf("user is already registered for this event")
		}
	}

	return nil
}

// setRegistrationStatus determines the appropriate status for a new registration
func (s *Service) setRegistrationStatus(ctx context.Context, registration *Registration, evt *event.Event) error {
	// Check if approval is required
	if evt.RegistrationSettings.RequiresApproval {
		registration.Status = StatusPendingApproval
		return nil
	}

	// Check capacity and set status accordingly
	return s.setStatusByCapacity(ctx, registration, evt)
}

// setStatusByCapacity sets registration status based on event capacity
func (s *Service) setStatusByCapacity(ctx context.Context, registration *Registration, evt *event.Event) error {
	confirmedCount, err := s.getConfirmedRegistrationCount(ctx, evt.ID)
	if err != nil {
		return fmt.Errorf("failed to get current registrations: %w", err)
	}

	if confirmedCount >= evt.Capacity.Maximum {
		registration.Status = StatusWaitlisted
		return s.setWaitlistPosition(ctx, registration, evt.ID)
	}

	registration.Status = StatusConfirmed
	now := time.Now()
	registration.ConfirmedAt = &now
	return nil
}

// getConfirmedRegistrationCount returns count of confirmed registrations
func (s *Service) getConfirmedRegistrationCount(ctx context.Context, eventID string) (int, error) {
	confirmedRegs, err := s.getConfirmedRegistrations(ctx, eventID)
	if err != nil {
		return 0, err
	}
	return len(confirmedRegs), nil
}

// setWaitlistPosition sets the waitlist position for a registration
func (s *Service) setWaitlistPosition(ctx context.Context, registration *Registration, eventID string) error {
	waitlistPos, err := s.getNextWaitlistPosition(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get waitlist position: %w", err)
	}
	registration.WaitlistPosition = &waitlistPos
	return nil
}

// getConfirmedRegistrations returns all confirmed registrations for an event
func (s *Service) getConfirmedRegistrations(ctx context.Context, eventID string) ([]*Registration, error) {
	allRegs, err := s.repo.GetRegistrationsByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	var confirmed []*Registration
	for _, reg := range allRegs {
		if reg.Status == StatusConfirmed {
			confirmed = append(confirmed, reg)
		}
	}

	return confirmed, nil
}

// getNextWaitlistPosition calculates the next position in the waitlist
func (s *Service) getNextWaitlistPosition(ctx context.Context, eventID string) (int, error) {
	allRegs, err := s.repo.GetRegistrationsByEventID(ctx, eventID)
	if err != nil {
		return 0, err
	}

	maxPosition := 0
	for _, reg := range allRegs {
		if reg.Status == StatusWaitlisted && reg.WaitlistPosition != nil && *reg.WaitlistPosition > maxPosition {
			maxPosition = *reg.WaitlistPosition
		}
	}

	return maxPosition + 1, nil
}
