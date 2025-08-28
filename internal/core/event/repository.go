package event

import (
	"context"
)

// Repository defines the interface for event data operations
type Repository interface {
	// Event CRUD operations
	Create(ctx context.Context, event *Event) error
	GetByID(ctx context.Context, id string) (*Event, error)
	GetBySlug(ctx context.Context, slug string) (*Event, error)
	Update(ctx context.Context, event *Event) error
	Delete(ctx context.Context, id string) error

	// Event listing and searching
	List(ctx context.Context, filter EventSearchFilter, sort *EventSortInput, limit, offset int) (*EventConnection, error)
	GetByOrganizer(ctx context.Context, organizerID string) ([]*Event, error)
	GetFeatured(ctx context.Context, limit int) ([]*Event, error)
	GetNearby(ctx context.Context, lat, lng, radius float64, limit int) ([]*Event, error)

	// Event status management
	UpdateStatus(ctx context.Context, eventID string, status EventStatus) error
	GetByStatus(ctx context.Context, status EventStatus, limit, offset int) ([]*Event, error)

	// Skill requirements
	CreateSkillRequirement(ctx context.Context, req *SkillRequirement) error
	GetSkillRequirements(ctx context.Context, eventID string) ([]*SkillRequirement, error)
	UpdateSkillRequirements(ctx context.Context, eventID string, requirements []*SkillRequirement) error
	DeleteSkillRequirements(ctx context.Context, eventID string) error

	// Training requirements
	CreateTrainingRequirement(ctx context.Context, req *TrainingRequirement) error
	GetTrainingRequirements(ctx context.Context, eventID string) ([]*TrainingRequirement, error)
	UpdateTrainingRequirements(ctx context.Context, eventID string, requirements []*TrainingRequirement) error
	DeleteTrainingRequirements(ctx context.Context, eventID string) error

	// Interest requirements
	AddInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error
	GetInterestRequirements(ctx context.Context, eventID string) ([]string, error)
	UpdateInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error
	RemoveInterestRequirements(ctx context.Context, eventID string) error

	// Event images
	CreateEventImage(ctx context.Context, image *EventImage) error
	GetEventImages(ctx context.Context, eventID string) ([]*EventImage, error)
	UpdateEventImage(ctx context.Context, image *EventImage) error
	DeleteEventImage(ctx context.Context, imageID string) error
	SetPrimaryImage(ctx context.Context, eventID, imageID string) error

	// Event announcements
	CreateAnnouncement(ctx context.Context, announcement *EventAnnouncement) error
	GetAnnouncements(ctx context.Context, eventID string) ([]*EventAnnouncement, error)
	UpdateAnnouncement(ctx context.Context, announcement *EventAnnouncement) error
	DeleteAnnouncement(ctx context.Context, announcementID string) error

	// Event updates/audit log
	LogUpdate(ctx context.Context, update *EventUpdate) error
	GetUpdateHistory(ctx context.Context, eventID string, limit, offset int) ([]*EventUpdate, error)

	// Recurring events
	GetEventInstances(ctx context.Context, parentEventID string) ([]*Event, error)
	GetUpcomingInstances(ctx context.Context, parentEventID string, limit int) ([]*Event, error)

	// Capacity management
	GetCurrentCapacity(ctx context.Context, eventID string) (int, error)
	IsAtCapacity(ctx context.Context, eventID string) (bool, error)

	// Utility functions
	EventExists(ctx context.Context, id string) (bool, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	GenerateUniqueSlug(ctx context.Context, title string) (string, error)
}
