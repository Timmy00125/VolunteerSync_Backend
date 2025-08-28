package event

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock repository for testing
type mockEventRepository struct {
	mock.Mock
}

func (m *mockEventRepository) Create(ctx context.Context, event *Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventRepository) GetByID(ctx context.Context, id string) (*Event, error) {
	args := m.Called(ctx, id)
	if event := args.Get(0); event != nil {
		return event.(*Event), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) GetBySlug(ctx context.Context, slug string) (*Event, error) {
	args := m.Called(ctx, slug)
	if event := args.Get(0); event != nil {
		return event.(*Event), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) Update(ctx context.Context, event *Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockEventRepository) List(ctx context.Context, filter EventSearchFilter, sort *EventSortInput, limit, offset int) (*EventConnection, error) {
	args := m.Called(ctx, filter, sort, limit, offset)
	if conn := args.Get(0); conn != nil {
		return conn.(*EventConnection), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) GetByOrganizer(ctx context.Context, organizerID string) ([]*Event, error) {
	args := m.Called(ctx, organizerID)
	if events := args.Get(0); events != nil {
		return events.([]*Event), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) GetFeatured(ctx context.Context, limit int) ([]*Event, error) {
	args := m.Called(ctx, limit)
	if events := args.Get(0); events != nil {
		return events.([]*Event), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) GetNearby(ctx context.Context, lat, lng, radius float64, limit int) ([]*Event, error) {
	args := m.Called(ctx, lat, lng, radius, limit)
	if events := args.Get(0); events != nil {
		return events.([]*Event), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) UpdateStatus(ctx context.Context, eventID string, status EventStatus) error {
	args := m.Called(ctx, eventID, status)
	return args.Error(0)
}

func (m *mockEventRepository) GetByStatus(ctx context.Context, status EventStatus, limit, offset int) ([]*Event, error) {
	args := m.Called(ctx, status, limit, offset)
	if events := args.Get(0); events != nil {
		return events.([]*Event), args.Error(1)
	}
	return nil, args.Error(1)
}

// Implement remaining interface methods as needed
func (m *mockEventRepository) GetSkillRequirements(ctx context.Context, eventID string) ([]*SkillRequirement, error) {
	args := m.Called(ctx, eventID)
	if skills := args.Get(0); skills != nil {
		return skills.([]*SkillRequirement), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) UpdateSkillRequirements(ctx context.Context, eventID string, requirements []*SkillRequirement) error {
	args := m.Called(ctx, eventID, requirements)
	return args.Error(0)
}

func (m *mockEventRepository) GetRequiredSkills(ctx context.Context, eventID string) ([]*SkillRequirement, error) {
	args := m.Called(ctx, eventID)
	if skills := args.Get(0); skills != nil {
		return skills.([]*SkillRequirement), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) UpdateRequiredSkills(ctx context.Context, eventID string, skills []*SkillRequirement) error {
	args := m.Called(ctx, eventID, skills)
	return args.Error(0)
}

func (m *mockEventRepository) GetTrainingRequirements(ctx context.Context, eventID string) ([]*TrainingRequirement, error) {
	args := m.Called(ctx, eventID)
	if reqs := args.Get(0); reqs != nil {
		return reqs.([]*TrainingRequirement), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) UpdateTrainingRequirements(ctx context.Context, eventID string, requirements []*TrainingRequirement) error {
	args := m.Called(ctx, eventID, requirements)
	return args.Error(0)
}

func (m *mockEventRepository) CreateEventImage(ctx context.Context, image *EventImage) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *mockEventRepository) GetEventImages(ctx context.Context, eventID string) ([]*EventImage, error) {
	args := m.Called(ctx, eventID)
	if images := args.Get(0); images != nil {
		return images.([]*EventImage), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) UpdateEventImage(ctx context.Context, image *EventImage) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *mockEventRepository) DeleteEventImage(ctx context.Context, imageID string) error {
	args := m.Called(ctx, imageID)
	return args.Error(0)
}

func (m *mockEventRepository) SetPrimaryImage(ctx context.Context, eventID, imageID string) error {
	args := m.Called(ctx, eventID, imageID)
	return args.Error(0)
}

func (m *mockEventRepository) CreateAnnouncement(ctx context.Context, announcement *EventAnnouncement) error {
	args := m.Called(ctx, announcement)
	return args.Error(0)
}

func (m *mockEventRepository) GetAnnouncements(ctx context.Context, eventID string) ([]*EventAnnouncement, error) {
	args := m.Called(ctx, eventID)
	if announcements := args.Get(0); announcements != nil {
		return announcements.([]*EventAnnouncement), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) UpdateAnnouncement(ctx context.Context, announcement *EventAnnouncement) error {
	args := m.Called(ctx, announcement)
	return args.Error(0)
}

func (m *mockEventRepository) DeleteAnnouncement(ctx context.Context, announcementID string) error {
	args := m.Called(ctx, announcementID)
	return args.Error(0)
}

func (m *mockEventRepository) LogUpdate(ctx context.Context, update *EventUpdate) error {
	args := m.Called(ctx, update)
	return args.Error(0)
}

func (m *mockEventRepository) GetUpdateHistory(ctx context.Context, eventID string, limit, offset int) ([]*EventUpdate, error) {
	args := m.Called(ctx, eventID, limit, offset)
	if updates := args.Get(0); updates != nil {
		return updates.([]*EventUpdate), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) GetEventInstances(ctx context.Context, parentEventID string) ([]*Event, error) {
	args := m.Called(ctx, parentEventID)
	if events := args.Get(0); events != nil {
		return events.([]*Event), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) GetUpcomingInstances(ctx context.Context, parentEventID string, limit int) ([]*Event, error) {
	args := m.Called(ctx, parentEventID, limit)
	if events := args.Get(0); events != nil {
		return events.([]*Event), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) GetCurrentCapacity(ctx context.Context, eventID string) (int, error) {
	args := m.Called(ctx, eventID)
	return args.Int(0), args.Error(1)
}

func (m *mockEventRepository) IsAtCapacity(ctx context.Context, eventID string) (bool, error) {
	args := m.Called(ctx, eventID)
	return args.Bool(0), args.Error(1)
}

func (m *mockEventRepository) EventExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockEventRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

func (m *mockEventRepository) GenerateUniqueSlug(ctx context.Context, title string) (string, error) {
	args := m.Called(ctx, title)
	return args.String(0), args.Error(1)
}

// Add missing interface methods
func (m *mockEventRepository) CreateSkillRequirement(ctx context.Context, req *SkillRequirement) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockEventRepository) DeleteSkillRequirements(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func (m *mockEventRepository) CreateTrainingRequirement(ctx context.Context, req *TrainingRequirement) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockEventRepository) DeleteTrainingRequirements(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func (m *mockEventRepository) AddInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error {
	args := m.Called(ctx, eventID, interestIDs)
	return args.Error(0)
}

func (m *mockEventRepository) GetInterestRequirements(ctx context.Context, eventID string) ([]string, error) {
	args := m.Called(ctx, eventID)
	if interests := args.Get(0); interests != nil {
		return interests.([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockEventRepository) UpdateInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error {
	args := m.Called(ctx, eventID, interestIDs)
	return args.Error(0)
}

func (m *mockEventRepository) RemoveInterestRequirements(ctx context.Context, eventID string) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

func createTestEventService() (*EventService, *mockEventRepository) {
	repo := &mockEventRepository{}
	service := NewEventService(repo)
	return service, repo
}

func createValidEventInput() CreateEventInput {
	now := time.Now().UTC()
	startTime := now.Add(24 * time.Hour)
	endTime := startTime.Add(2 * time.Hour)
	closesAt := startTime.Add(-1 * time.Hour)

	return CreateEventInput{
		Title:            "Test Event",
		Description:      "A test event for unit testing",
		ShortDescription: stringPtr("Short description"),
		Category:         EventCategoryEnvironment,
		StartTime:        startTime,
		EndTime:          endTime,
		Location: EventLocationInput{
			Name:         "Test Location",
			Address:      "123 Test St",
			City:         "Test City",
			State:        stringPtr("TS"),
			Country:      "US",
			ZipCode:      stringPtr("12345"),
			Instructions: stringPtr("Test instructions"),
			IsRemote:     false,
		},
		Capacity: EventCapacityInput{
			Minimum:         1,
			Maximum:         10,
			WaitlistEnabled: true,
		},
		RegistrationSettings: RegistrationSettingsInput{
			OpensAt:              nil,
			ClosesAt:             closesAt,
			RequiresApproval:     false,
			AllowWaitlist:        true,
			ConfirmationRequired: false,
			CancellationDeadline: nil,
		},
		TimeCommitment: TimeCommitmentOneTime,
		Tags:           []string{"test", "environment"},
	}
}

func TestNewEventService(t *testing.T) {
	service, _ := createTestEventService()
	assert.NotNil(t, service)
}

func TestEventService_CreateEvent(t *testing.T) {
	service, repo := createTestEventService()
	ctx := context.Background()

	t.Run("successful event creation", func(t *testing.T) {
		input := createValidEventInput()
		organizerID := "organizer123"

		repo.On("Create", ctx, mock.MatchedBy(func(event *Event) bool {
			return event.Title == input.Title &&
				event.OrganizerID == organizerID &&
				event.Status == EventStatusDraft
		})).Return(nil).Once()

		event, err := service.CreateEvent(ctx, organizerID, input)

		require.NoError(t, err)
		assert.Equal(t, input.Title, event.Title)
		assert.Equal(t, organizerID, event.OrganizerID)
		assert.Equal(t, EventStatusDraft, event.Status)
		assert.NotEmpty(t, event.ID)
		assert.NotNil(t, event.Slug)
		assert.NotNil(t, event.ShareURL)
		repo.AssertExpectations(t)
	})

	t.Run("validation failure - start time in past", func(t *testing.T) {
		input := createValidEventInput()
		input.StartTime = time.Now().UTC().Add(-1 * time.Hour) // Past time
		organizerID := "organizer123"

		event, err := service.CreateEvent(ctx, organizerID, input)

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("validation failure - invalid capacity", func(t *testing.T) {
		input := createValidEventInput()
		input.Capacity.Maximum = 0 // Invalid capacity
		organizerID := "organizer123"

		event, err := service.CreateEvent(ctx, organizerID, input)

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("repository error", func(t *testing.T) {
		input := createValidEventInput()
		organizerID := "organizer123"

		repo.On("Create", ctx, mock.AnythingOfType("*event.Event")).Return(assert.AnError).Once()

		event, err := service.CreateEvent(ctx, organizerID, input)

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "failed to create event")
		repo.AssertExpectations(t)
	})
}

func TestEventService_GetEvent(t *testing.T) {
	service, repo := createTestEventService()
	ctx := context.Background()

	expectedEvent := &Event{
		ID:          "event123",
		Title:       "Test Event",
		Description: "Test description",
		Status:      EventStatusPublished,
	}

	t.Run("successful event retrieval", func(t *testing.T) {
		repo.On("GetByID", ctx, "event123").Return(expectedEvent, nil).Once()

		event, err := service.GetEvent(ctx, "event123")

		require.NoError(t, err)
		assert.Equal(t, expectedEvent.ID, event.ID)
		assert.Equal(t, expectedEvent.Title, event.Title)
		repo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		repo.On("GetByID", ctx, "nonexistent").Return(nil, assert.AnError).Once()

		event, err := service.GetEvent(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, event)
		repo.AssertExpectations(t)
	})
}

func TestEventService_UpdateEvent(t *testing.T) {
	service, repo := createTestEventService()
	ctx := context.Background()

	existingEvent := &Event{
		ID:          "event123",
		Title:       "Original Title",
		Description: "Original description",
		OrganizerID: "organizer123",
		Status:      EventStatusDraft,
		StartTime:   time.Now().UTC().Add(24 * time.Hour),
		EndTime:     time.Now().UTC().Add(26 * time.Hour),
		Location: EventLocation{
			Name:    "Original Location",
			Address: "Original Address",
			City:    "Original City",
			Country: "US",
		},
		Capacity: EventCapacity{
			Minimum: 1,
			Maximum: 10,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	t.Run("successful event update", func(t *testing.T) {
		input := UpdateEventInput{
			Title:       stringPtr("Updated Title"),
			Description: stringPtr("Updated description"),
		}

		repo.On("GetByID", ctx, "event123").Return(existingEvent, nil).Once()
		repo.On("Update", ctx, mock.MatchedBy(func(event *Event) bool {
			return event.Title == "Updated Title" &&
				event.Description == "Updated description" &&
				event.OrganizerID == "organizer123"
		})).Return(nil).Once()

		event, err := service.UpdateEvent(ctx, "event123", "organizer123", input)

		require.NoError(t, err)
		assert.Equal(t, "Updated Title", event.Title)
		assert.Equal(t, "Updated description", event.Description)
		repo.AssertExpectations(t)
	})

	t.Run("unauthorized update - wrong organizer", func(t *testing.T) {
		input := UpdateEventInput{
			Title: stringPtr("Updated Title"),
		}

		repo.On("GetByID", ctx, "event123").Return(existingEvent, nil).Once()

		event, err := service.UpdateEvent(ctx, "event123", "wronguser", input)

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "unauthorized")
		repo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		input := UpdateEventInput{
			Title: stringPtr("Updated Title"),
		}

		repo.On("GetByID", ctx, "nonexistent").Return(nil, assert.AnError).Once()

		event, err := service.UpdateEvent(ctx, "nonexistent", "organizer123", input)

		assert.Error(t, err)
		assert.Nil(t, event)
		repo.AssertExpectations(t)
	})
}

func TestEventService_PublishEvent(t *testing.T) {
	service, repo := createTestEventService()
	ctx := context.Background()

	draftEvent := &Event{
		ID:          "event123",
		Title:       "Complete Event",
		Description: "Complete description",
		OrganizerID: "organizer123",
		Status:      EventStatusDraft,
		StartTime:   time.Now().UTC().Add(24 * time.Hour),
		EndTime:     time.Now().UTC().Add(26 * time.Hour),
		Location: EventLocation{
			Name: "Valid Location",
		},
		Capacity: EventCapacity{
			Maximum: 10,
		},
	}

	publishedEvent := *draftEvent
	publishedEvent.Status = EventStatusPublished

	t.Run("successful event publishing", func(t *testing.T) {
		repo.On("GetByID", ctx, "event123").Return(draftEvent, nil).Once()
		repo.On("UpdateStatus", ctx, "event123", EventStatusPublished).Return(nil).Once()
		repo.On("GetByID", ctx, "event123").Return(&publishedEvent, nil).Once()

		event, err := service.PublishEvent(ctx, "event123", "organizer123")

		require.NoError(t, err)
		assert.Equal(t, EventStatusPublished, event.Status)
		repo.AssertExpectations(t)
	})

	t.Run("unauthorized publish - wrong organizer", func(t *testing.T) {
		repo.On("GetByID", ctx, "event123").Return(draftEvent, nil).Once()

		event, err := service.PublishEvent(ctx, "event123", "wronguser")

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "unauthorized")
		repo.AssertExpectations(t)
	})

	t.Run("cannot publish non-draft event", func(t *testing.T) {
		publishedEventAlready := *draftEvent
		publishedEventAlready.Status = EventStatusPublished

		repo.On("GetByID", ctx, "event123").Return(&publishedEventAlready, nil).Once()

		event, err := service.PublishEvent(ctx, "event123", "organizer123")

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "not in draft status")
		repo.AssertExpectations(t)
	})
}

func TestEventService_CancelEvent(t *testing.T) {
	service, repo := createTestEventService()
	ctx := context.Background()

	publishedEvent := &Event{
		ID:          "event123",
		Title:       "Event to Cancel",
		OrganizerID: "organizer123",
		Status:      EventStatusPublished,
	}

	cancelledEvent := *publishedEvent
	cancelledEvent.Status = EventStatusCancelled

	t.Run("successful event cancellation", func(t *testing.T) {
		repo.On("GetByID", ctx, "event123").Return(publishedEvent, nil).Once()
		repo.On("UpdateStatus", ctx, "event123", EventStatusCancelled).Return(nil).Once()
		repo.On("GetByID", ctx, "event123").Return(&cancelledEvent, nil).Once()

		event, err := service.CancelEvent(ctx, "event123", "organizer123", "Event cancelled due to weather")

		require.NoError(t, err)
		assert.Equal(t, EventStatusCancelled, event.Status)
		repo.AssertExpectations(t)
	})

	t.Run("unauthorized cancellation", func(t *testing.T) {
		repo.On("GetByID", ctx, "event123").Return(publishedEvent, nil).Once()

		event, err := service.CancelEvent(ctx, "event123", "wronguser", "reason")

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "unauthorized")
		repo.AssertExpectations(t)
	})

	t.Run("cannot cancel already cancelled event", func(t *testing.T) {
		alreadyCancelled := *publishedEvent
		alreadyCancelled.Status = EventStatusCancelled

		repo.On("GetByID", ctx, "event123").Return(&alreadyCancelled, nil).Once()

		event, err := service.CancelEvent(ctx, "event123", "organizer123", "reason")

		assert.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "cannot be cancelled")
		repo.AssertExpectations(t)
	})
}

func TestEventService_DeleteEvent(t *testing.T) {
	service, repo := createTestEventService()
	ctx := context.Background()

	event := &Event{
		ID:          "event123",
		OrganizerID: "organizer123",
		Status:      EventStatusDraft,
	}

	t.Run("successful event deletion", func(t *testing.T) {
		repo.On("GetByID", ctx, "event123").Return(event, nil).Once()
		repo.On("Delete", ctx, "event123").Return(nil).Once()

		err := service.DeleteEvent(ctx, "event123", "organizer123")

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("unauthorized deletion", func(t *testing.T) {
		repo.On("GetByID", ctx, "event123").Return(event, nil).Once()

		err := service.DeleteEvent(ctx, "event123", "wronguser")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
		repo.AssertExpectations(t)
	})

	t.Run("event not found", func(t *testing.T) {
		repo.On("GetByID", ctx, "nonexistent").Return(nil, assert.AnError).Once()

		err := service.DeleteEvent(ctx, "nonexistent", "organizer123")

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

func TestValidateEventTimes(t *testing.T) {
	now := time.Now().UTC()

	testCases := []struct {
		name      string
		startTime time.Time
		endTime   time.Time
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid times",
			startTime: now.Add(1 * time.Hour),
			endTime:   now.Add(3 * time.Hour),
			expectErr: false,
		},
		{
			name:      "start time in past",
			startTime: now.Add(-1 * time.Hour),
			endTime:   now.Add(1 * time.Hour),
			expectErr: true,
			errMsg:    "start time cannot be in the past",
		},
		{
			name:      "end time before start time",
			startTime: now.Add(2 * time.Hour),
			endTime:   now.Add(1 * time.Hour),
			expectErr: true,
			errMsg:    "end time cannot be before start time",
		},
		{
			name:      "duration too short",
			startTime: now.Add(1 * time.Hour),
			endTime:   now.Add(1*time.Hour + 15*time.Minute),
			expectErr: true,
			errMsg:    "event duration must be at least 30 minutes",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEventTimes(tc.startTime, tc.endTime)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateSlug(t *testing.T) {
	testCases := []struct {
		title    string
		expected string
	}{
		{
			title:    "Hello World",
			expected: "hello-world",
		},
		{
			title:    "Multiple   Spaces",
			expected: "multiple-spaces",
		},
		{
			title:    "Special!@#$%Characters",
			expected: "specialcharacters",
		},
		{
			title:    "Numbers123AndText",
			expected: "numbers123andtext",
		},
		{
			title:    "---Multiple---Hyphens---",
			expected: "multiple-hyphens",
		},
		{
			title:    "Very Long Title That Should Be Truncated To Fifty Characters Maximum Length",
			expected: "very-long-title-that-should-be-truncated-to-fifty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			result := generateSlug(tc.title)
			assert.Equal(t, tc.expected, result)
			assert.LessOrEqual(t, len(result), 50)
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}