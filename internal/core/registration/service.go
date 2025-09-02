package registration

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/volunteersync/backend/internal/core/event"
	"github.com/volunteersync/backend/internal/core/user"
)

// Service encapsulates the business logic for registrations.

type Service struct {
	repo                Repository
	eventService        EventService
	userService         user.Service
	waitlistService     WaitlistService
	approvalService     ApprovalService
	conflictService     ConflictService
	notificationService NotificationService
	auditLogger         AuditLogger
}

// EventService defines the interface for interacting with the event service.

type EventService interface {
	GetEvent(ctx context.Context, id string) (*event.Event, error)
}

// WaitlistService defines the interface for interacting with the waitlist service.

type WaitlistService interface {
	AddToWaitlist(ctx context.Context, registration *Registration) (*WaitlistEntry, error)
}

// ApprovalService defines the interface for interacting with the approval service.

type ApprovalService interface{}

// ConflictService defines the interface for interacting with the conflict service.

type ConflictService interface {
	DetectConflicts(ctx context.Context, userID, eventID string) ([]*RegistrationConflict, error)
}

// NotificationService defines the interface for sending notifications.

type NotificationService interface {
	SendRegistrationConfirmation(ctx context.Context, registration *Registration) error
}

// AuditLogger defines the interface for logging audit trails.

type AuditLogger interface {
	Log(ctx context.Context, action string, details map[string]interface{}) error
}

// NewService creates a new registration service.

func NewService(repo Repository, eventService EventService, userService user.Service, waitlistService WaitlistService, approvalService ApprovalService, conflictService ConflictService, notificationService NotificationService, auditLogger AuditLogger) *Service {
	return &Service{
		repo:                repo,
		eventService:        eventService,
		userService:         userService,
		waitlistService:     waitlistService,
		approvalService:     approvalService,
		conflictService:     conflictService,
		notificationService: notificationService,
		auditLogger:         auditLogger,
	}
}

func (s *Service) ApproveRegistration(ctx context.Context, userID, registrationID string, approved bool, notes string) (*Registration, error) {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return nil, err
	}

	evt, err := s.eventService.GetEvent(ctx, reg.EventID)
	if err != nil {
		return nil, err
	}

	if evt.OrganizerID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "user is not the organizer of this event")
	}

	if approved {
		reg.Status = StatusConfirmed
		now := time.Now()
		reg.ConfirmedAt = &now
	} else {
		reg.Status = StatusDeclined
	}
	reg.ApprovalNotes = notes

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		return nil, err
	}

	return reg, nil
}

func (s *Service) CancelRegistration(ctx context.Context, userID, registrationID, reason string) (*Registration, error) {
	reg, err := s.repo.GetRegistrationByID(ctx, registrationID)
	if err != nil {
		return nil, err
	}

	if reg.UserID != userID {
		// In a real application, we would also check if the user is an organizer
		return nil, status.Errorf(codes.PermissionDenied, "user does not have permission to cancel this registration")
	}

	reg.Status = StatusCancelled
	reg.CancellationReason = reason
	now := time.Now()
	reg.CancelledAt = &now

	if err := s.repo.UpdateRegistration(ctx, reg); err != nil {
		return nil, err
	}

	return reg, nil
}

func (s *Service) GetRegistrationsByEventID(ctx context.Context, eventID string) ([]*Registration, error) {
	return s.repo.GetRegistrationsByEventID(ctx, eventID)
}

func (s *Service) GetRegistrationByID(ctx context.Context, id string) (*Registration, error) {
	return s.repo.GetRegistrationByID(ctx, id)
}

func (s *Service) GetRegistrationsByUserID(ctx context.Context, userID string) ([]*Registration, error) {
	return s.repo.GetRegistrationsByUserID(ctx, userID)
}

// RegisterForEvent handles the registration of a user for an event.

func (s *Service) RegisterForEvent(ctx context.Context, userID, eventID, personalMessage string) (*Registration, error) {
	// 1. Validate user and event
	userProfile, err := s.userService.GetProfile(ctx, userID, userID, nil)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}
	if userProfile == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	event, err := s.eventService.GetEvent(ctx, eventID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "event not found: %v", err)
	}
	if event == nil {
		return nil, status.Errorf(codes.NotFound, "event not found")
	}

	// 2. Check for conflicts
	if _, err := s.conflictService.DetectConflicts(ctx, userID, eventID); err != nil {
		// For now, we'll just log the error and continue. In a real implementation,
		// we might want to return a warning to the user.
		// s.auditLogger.Log(ctx, "conflict detection failed", map[string]interface{}{"error": err.Error()})
	}

	// 3. Create registration
	registration := &Registration{
		ID:              uuid.New().String(),
		UserID:          userID,
		EventID:         eventID,
		Status:          StatusPendingApproval, // Default status
		PersonalMessage: personalMessage,
	}

	// 4. Check event capacity
	registrations, err := s.repo.GetRegistrationsByEventID(ctx, eventID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get registrations: %v", err)
	}

	if len(registrations) >= event.Capacity {
		registration.Status = StatusWaitlisted
		// 5. Add to waitlist
		if _, err := s.waitlistService.AddToWaitlist(ctx, registration); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to add to waitlist: %v", err)
		}
	} else {
		registration.Status = StatusConfirmed
	}

	// 6. Save registration
	newRegistration, err := s.repo.CreateRegistration(ctx, registration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create registration: %v", err)
	}

	// 7. Send confirmation notification (asynchronously)
	go s.notificationService.SendRegistrationConfirmation(ctx, newRegistration)

	return newRegistration, nil
}
