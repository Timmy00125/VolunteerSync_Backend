package graph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/volunteersync/backend/internal/core/auth"
	"github.com/volunteersync/backend/internal/core/event"
	mw "github.com/volunteersync/backend/internal/middleware"
)

// fakeEventRepo is a minimal in-memory implementation of event.Repository for tests
type fakeEventRepo struct {
	events map[string]*event.Event
}

func newFakeEventRepo() *fakeEventRepo { return &fakeEventRepo{events: map[string]*event.Event{}} }

// CRUD
func (f *fakeEventRepo) Create(ctx context.Context, e *event.Event) error {
	f.events[e.ID] = e
	return nil
}
func (f *fakeEventRepo) GetByID(ctx context.Context, id string) (*event.Event, error) {
	if e, ok := f.events[id]; ok {
		return e, nil
	}
	return nil, assert.AnError
}
func (f *fakeEventRepo) GetBySlug(ctx context.Context, slug string) (*event.Event, error) {
	return nil, assert.AnError
}
func (f *fakeEventRepo) Update(ctx context.Context, e *event.Event) error {
	f.events[e.ID] = e
	return nil
}
func (f *fakeEventRepo) Delete(ctx context.Context, id string) error {
	if _, ok := f.events[id]; ok {
		delete(f.events, id)
		return nil
	}
	return assert.AnError
}

// Listing/search
func (f *fakeEventRepo) List(ctx context.Context, filter event.EventSearchFilter, sort *event.EventSortInput, limit, offset int) (*event.EventConnection, error) {
	return &event.EventConnection{Edges: []event.EventEdge{}, PageInfo: event.PageInfo{}, TotalCount: 0}, nil
}
func (f *fakeEventRepo) GetByOrganizer(ctx context.Context, organizerID string) ([]*event.Event, error) {
	var out []*event.Event
	for _, e := range f.events {
		if e.OrganizerID == organizerID {
			out = append(out, e)
		}
	}
	return out, nil
}
func (f *fakeEventRepo) GetFeatured(ctx context.Context, limit int) ([]*event.Event, error) {
	return nil, nil
}
func (f *fakeEventRepo) GetNearby(ctx context.Context, lat, lng, radius float64, limit int) ([]*event.Event, error) {
	return nil, nil
}

// Status
func (f *fakeEventRepo) UpdateStatus(ctx context.Context, eventID string, status event.EventStatus) error {
	if e, ok := f.events[eventID]; ok {
		e.Status = status
		return nil
	}
	return assert.AnError
}
func (f *fakeEventRepo) GetByStatus(ctx context.Context, status event.EventStatus, limit, offset int) ([]*event.Event, error) {
	return nil, nil
}

// Skill reqs
func (f *fakeEventRepo) CreateSkillRequirement(ctx context.Context, req *event.SkillRequirement) error {
	return nil
}
func (f *fakeEventRepo) GetSkillRequirements(ctx context.Context, eventID string) ([]*event.SkillRequirement, error) {
	return nil, nil
}
func (f *fakeEventRepo) UpdateSkillRequirements(ctx context.Context, eventID string, requirements []*event.SkillRequirement) error {
	return nil
}
func (f *fakeEventRepo) DeleteSkillRequirements(ctx context.Context, eventID string) error {
	return nil
}

// Training reqs
func (f *fakeEventRepo) CreateTrainingRequirement(ctx context.Context, req *event.TrainingRequirement) error {
	return nil
}
func (f *fakeEventRepo) GetTrainingRequirements(ctx context.Context, eventID string) ([]*event.TrainingRequirement, error) {
	return nil, nil
}
func (f *fakeEventRepo) UpdateTrainingRequirements(ctx context.Context, eventID string, requirements []*event.TrainingRequirement) error {
	return nil
}
func (f *fakeEventRepo) DeleteTrainingRequirements(ctx context.Context, eventID string) error {
	return nil
}

// Interest reqs
func (f *fakeEventRepo) AddInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error {
	return nil
}
func (f *fakeEventRepo) GetInterestRequirements(ctx context.Context, eventID string) ([]string, error) {
	return nil, nil
}
func (f *fakeEventRepo) UpdateInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error {
	return nil
}
func (f *fakeEventRepo) RemoveInterestRequirements(ctx context.Context, eventID string) error {
	return nil
}

// Images
func (f *fakeEventRepo) CreateEventImage(ctx context.Context, image *event.EventImage) error {
	return nil
}
func (f *fakeEventRepo) GetEventImages(ctx context.Context, eventID string) ([]*event.EventImage, error) {
	return nil, nil
}
func (f *fakeEventRepo) UpdateEventImage(ctx context.Context, image *event.EventImage) error {
	return nil
}
func (f *fakeEventRepo) DeleteEventImage(ctx context.Context, imageID string) error { return nil }
func (f *fakeEventRepo) SetPrimaryImage(ctx context.Context, eventID, imageID string) error {
	return nil
}

// Announcements
func (f *fakeEventRepo) CreateAnnouncement(ctx context.Context, announcement *event.EventAnnouncement) error {
	return nil
}
func (f *fakeEventRepo) GetAnnouncements(ctx context.Context, eventID string) ([]*event.EventAnnouncement, error) {
	return nil, nil
}
func (f *fakeEventRepo) UpdateAnnouncement(ctx context.Context, announcement *event.EventAnnouncement) error {
	return nil
}
func (f *fakeEventRepo) DeleteAnnouncement(ctx context.Context, announcementID string) error {
	return nil
}

// Updates
func (f *fakeEventRepo) LogUpdate(ctx context.Context, update *event.EventUpdate) error { return nil }
func (f *fakeEventRepo) GetUpdateHistory(ctx context.Context, eventID string, limit, offset int) ([]*event.EventUpdate, error) {
	return nil, nil
}

// Recurring
func (f *fakeEventRepo) GetEventInstances(ctx context.Context, parentEventID string) ([]*event.Event, error) {
	return nil, nil
}
func (f *fakeEventRepo) GetUpcomingInstances(ctx context.Context, parentEventID string, limit int) ([]*event.Event, error) {
	return nil, nil
}

// Capacity
func (f *fakeEventRepo) GetCurrentCapacity(ctx context.Context, eventID string) (int, error) {
	return 0, nil
}
func (f *fakeEventRepo) IsAtCapacity(ctx context.Context, eventID string) (bool, error) {
	return false, nil
}

// Utils
func (f *fakeEventRepo) EventExists(ctx context.Context, id string) (bool, error) {
	_, ok := f.events[id]
	return ok, nil
}
func (f *fakeEventRepo) SlugExists(ctx context.Context, slug string) (bool, error) { return false, nil }
func (f *fakeEventRepo) GenerateUniqueSlug(ctx context.Context, title string) (string, error) {
	return title, nil
}

func TestDeleteEventMutation(t *testing.T) {
	repo := newFakeEventRepo()
	svc := event.NewEventService(repo)
	r := &Resolver{EventService: svc}
	m := &mutationResolver{r}

	// Seed an event
	ev := &event.Event{
		ID:          "evt-1",
		Title:       "T",
		Description: "D",
		OrganizerID: "user-1",
		Status:      event.EventStatusDraft,
		StartTime:   time.Now().Add(1 * time.Hour).UTC(),
		EndTime:     time.Now().Add(2 * time.Hour).UTC(),
		Location:    event.EventLocation{Name: "loc", Address: "a", City: "c", Country: "US", IsRemote: true},
		Capacity:    event.EventCapacity{Minimum: 0, Maximum: 10},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, repo.Create(context.Background(), ev))

	t.Run("requires auth", func(t *testing.T) {
		ok, err := m.DeleteEvent(context.Background(), "evt-1")
		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("deletes when organizer", func(t *testing.T) {
		claims := &auth.UserClaims{UserID: "user-1"}
		ctx := context.WithValue(context.Background(), mw.UserClaimsContextKey, claims)

		ok, err := m.DeleteEvent(ctx, "evt-1")
		require.NoError(t, err)
		assert.True(t, ok)

		// Ensure it's gone
		_, err = repo.GetByID(context.Background(), "evt-1")
		assert.Error(t, err)
	})
}
