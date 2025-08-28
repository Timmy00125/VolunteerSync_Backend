package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type mockUserStore struct {
	mock.Mock
}

func (m *mockUserStore) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
	args := m.Called(ctx, userID)
	if profile := args.Get(0); profile != nil {
		return profile.(*UserProfile), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserStore) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*UserProfile, error) {
	args := m.Called(ctx, userID, input)
	if profile := args.Get(0); profile != nil {
		return profile.(*UserProfile), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserStore) SetProfilePicture(ctx context.Context, userID, url string) error {
	args := m.Called(ctx, userID, url)
	return args.Error(0)
}

func (m *mockUserStore) ReplaceInterests(ctx context.Context, userID string, interestIDs []string) ([]Interest, error) {
	args := m.Called(ctx, userID, interestIDs)
	if interests := args.Get(0); interests != nil {
		return interests.([]Interest), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserStore) ListInterests(ctx context.Context) ([]Interest, error) {
	args := m.Called(ctx)
	if interests := args.Get(0); interests != nil {
		return interests.([]Interest), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserStore) ListUserInterests(ctx context.Context, userID string) ([]Interest, error) {
	args := m.Called(ctx, userID)
	if interests := args.Get(0); interests != nil {
		return interests.([]Interest), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserStore) AddSkill(ctx context.Context, userID string, in SkillInput) (*Skill, error) {
	args := m.Called(ctx, userID, in)
	if skill := args.Get(0); skill != nil {
		return skill.(*Skill), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserStore) RemoveSkill(ctx context.Context, userID, skillID string) error {
	args := m.Called(ctx, userID, skillID)
	return args.Error(0)
}

func (m *mockUserStore) ListSkills(ctx context.Context, userID string) ([]Skill, error) {
	args := m.Called(ctx, userID)
	if skills := args.Get(0); skills != nil {
		return skills.([]Skill), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserStore) UpdatePrivacy(ctx context.Context, userID string, in PrivacySettings) (PrivacySettings, error) {
	args := m.Called(ctx, userID, in)
	return args.Get(0).(PrivacySettings), args.Error(1)
}

func (m *mockUserStore) UpdateNotifications(ctx context.Context, userID string, in NotificationPreferences) (NotificationPreferences, error) {
	args := m.Called(ctx, userID, in)
	return args.Get(0).(NotificationPreferences), args.Error(1)
}

func (m *mockUserStore) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if roles := args.Get(0); roles != nil {
		return roles.([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserStore) SetUserRoles(ctx context.Context, userID string, roles []string, assignedBy string) error {
	args := m.Called(ctx, userID, roles, assignedBy)
	return args.Error(0)
}

func (m *mockUserStore) SearchUsers(ctx context.Context, filter UserSearchFilter, limit, offset int) ([]UserProfile, error) {
	args := m.Called(ctx, filter, limit, offset)
	if users := args.Get(0); users != nil {
		return users.([]UserProfile), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserStore) LogActivity(ctx context.Context, log ActivityLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockUserStore) ListActivityLogs(ctx context.Context, userID string, limit, offset int) ([]ActivityLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	if logs := args.Get(0); logs != nil {
		return logs.([]ActivityLog), args.Error(1)
	}
	return nil, args.Error(1)
}

type mockFileService struct {
	mock.Mock
}

func (m *mockFileService) SaveProfileImage(ctx context.Context, userID string, data []byte, mimeType string) (url, storagePath string, err error) {
	args := m.Called(ctx, userID, data, mimeType)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *mockFileService) Delete(ctx context.Context, storagePath string) error {
	args := m.Called(ctx, storagePath)
	return args.Error(0)
}

type mockNotificationService struct {
	mock.Mock
}

func (m *mockNotificationService) NotifyProfileUpdated(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type mockAuditLogger struct {
	mock.Mock
}

func (m *mockAuditLogger) Info(ctx context.Context, action string, details map[string]any) {
	m.Called(ctx, action, details)
}

func (m *mockAuditLogger) Warn(ctx context.Context, action string, details map[string]any) {
	m.Called(ctx, action, details)
}

// Test helper function to create a service with mocks
func createTestService() (*Service, *mockUserStore, *mockFileService, *mockNotificationService, *mockAuditLogger) {
	store := &mockUserStore{}
	files := &mockFileService{}
	notifier := &mockNotificationService{}
	audit := &mockAuditLogger{}
	service := NewService(store, files, notifier, audit)
	return service, store, files, notifier, audit
}

func TestNewService(t *testing.T) {
	service, _, _, _, _ := createTestService()
	assert.NotNil(t, service)
}

func TestService_GetProfile(t *testing.T) {
	service, store, _, _, _ := createTestService()
	ctx := context.Background()

	profile := &UserProfile{
		ID:    "user1",
		Name:  "John Doe",
		Email: "john@example.com",
		Privacy: PrivacySettings{
			ProfileVisibility: "PUBLIC",
			ShowEmail:         true,
			ShowLocation:      true,
		},
	}

	t.Run("successful profile retrieval", func(t *testing.T) {
		store.On("GetProfile", ctx, "user1").Return(profile, nil).Once()

		result, err := service.GetProfile(ctx, "user1", "user1", []string{})
		
		require.NoError(t, err)
		assert.Equal(t, profile.ID, result.ID)
		assert.Equal(t, profile.Name, result.Name)
		store.AssertExpectations(t)
	})

	t.Run("profile not found", func(t *testing.T) {
		store.On("GetProfile", ctx, "nonexistent").Return(nil, ErrUserNotFound).Once()

		result, err := service.GetProfile(ctx, "nonexistent", "user1", []string{})
		
		assert.Error(t, err)
		assert.Nil(t, result)
		store.AssertExpectations(t)
	})
}

func TestService_UpdateProfile(t *testing.T) {
	service, store, _, notifier, audit := createTestService()
	ctx := context.Background()

	input := UpdateProfileInput{
		Name: stringPtr("Jane Doe"),
		Bio:  stringPtr("Updated bio"),
	}

	updatedProfile := &UserProfile{
		ID:   "user1",
		Name: "Jane Doe",
		Bio:  stringPtr("Updated bio"),
	}

	t.Run("successful profile update", func(t *testing.T) {
		store.On("UpdateProfile", ctx, "user1", input).Return(updatedProfile, nil).Once()
		notifier.On("NotifyProfileUpdated", ctx, "user1").Return(nil).Once()
		audit.On("Info", ctx, "user.profile.update", map[string]any{"user_id": "user1"}).Once()

		result, err := service.UpdateProfile(ctx, "user1", input)
		
		require.NoError(t, err)
		assert.Equal(t, updatedProfile.Name, result.Name)
		store.AssertExpectations(t)
		notifier.AssertExpectations(t)
		audit.AssertExpectations(t)
	})
}

func TestService_UpdateInterests(t *testing.T) {
	service, store, _, _, audit := createTestService()
	ctx := context.Background()

	interestIDs := []string{"int1", "int2"}
	interests := []Interest{
		{ID: "int1", Name: "Environment", Category: "causes"},
		{ID: "int2", Name: "Education", Category: "causes"},
	}
	profile := &UserProfile{
		ID:        "user1",
		Name:      "John Doe",
		Interests: interests,
	}

	t.Run("successful interests update", func(t *testing.T) {
		store.On("ReplaceInterests", ctx, "user1", interestIDs).Return(interests, nil).Once()
		store.On("GetProfile", ctx, "user1").Return(profile, nil).Once()
		audit.On("Info", ctx, "user.interests.update", map[string]any{"user_id": "user1", "count": 2}).Once()

		result, err := service.UpdateInterests(ctx, "user1", interestIDs)
		
		require.NoError(t, err)
		assert.Len(t, result.Interests, 2)
		store.AssertExpectations(t)
		audit.AssertExpectations(t)
	})
}

func TestService_AddSkill(t *testing.T) {
	service, store, _, _, audit := createTestService()
	ctx := context.Background()

	skillInput := SkillInput{
		Name:        "JavaScript",
		Proficiency: "INTERMEDIATE",
	}

	skill := &Skill{
		ID:          "skill1",
		Name:        "JavaScript",
		Proficiency: "INTERMEDIATE",
	}

	profile := &UserProfile{
		ID:     "user1",
		Name:   "John Doe",
		Skills: []Skill{*skill},
	}

	t.Run("successful skill addition", func(t *testing.T) {
		store.On("AddSkill", ctx, "user1", skillInput).Return(skill, nil).Once()
		store.On("GetProfile", ctx, "user1").Return(profile, nil).Once()
		store.On("ListSkills", ctx, "user1").Return([]Skill{*skill}, nil).Once()
		audit.On("Info", ctx, "user.skill.add", map[string]any{"user_id": "user1", "name": "JavaScript"}).Once()

		result, err := service.AddSkill(ctx, "user1", skillInput)
		
		require.NoError(t, err)
		assert.Len(t, result.Skills, 1)
		assert.Equal(t, "JavaScript", result.Skills[0].Name)
		store.AssertExpectations(t)
		audit.AssertExpectations(t)
	})
}

func TestService_RemoveSkill(t *testing.T) {
	service, store, _, _, audit := createTestService()
	ctx := context.Background()

	profile := &UserProfile{
		ID:     "user1",
		Name:   "John Doe",
		Skills: []Skill{},
	}

	t.Run("successful skill removal", func(t *testing.T) {
		store.On("RemoveSkill", ctx, "user1", "skill1").Return(nil).Once()
		store.On("GetProfile", ctx, "user1").Return(profile, nil).Once()
		store.On("ListSkills", ctx, "user1").Return([]Skill{}, nil).Once()
		audit.On("Info", ctx, "user.skill.remove", map[string]any{"user_id": "user1", "skill_id": "skill1"}).Once()

		result, err := service.RemoveSkill(ctx, "user1", "skill1")
		
		require.NoError(t, err)
		assert.Len(t, result.Skills, 0)
		store.AssertExpectations(t)
		audit.AssertExpectations(t)
	})
}

func TestService_UploadProfilePicture(t *testing.T) {
	service, store, files, _, audit := createTestService()
	ctx := context.Background()

	imageData := []byte("fake image data")
	mimeType := "image/jpeg"
	expectedURL := "https://example.com/profile.jpg"
	storagePath := "profiles/user1/profile.jpg"

	t.Run("successful profile picture upload", func(t *testing.T) {
		files.On("SaveProfileImage", ctx, "user1", imageData, mimeType).Return(expectedURL, storagePath, nil).Once()
		store.On("SetProfilePicture", ctx, "user1", expectedURL).Return(nil).Once()
		audit.On("Info", ctx, "user.profile.picture.update", map[string]any{"user_id": "user1"}).Once()

		url, err := service.UploadProfilePicture(ctx, "user1", imageData, mimeType)
		
		require.NoError(t, err)
		assert.Equal(t, expectedURL, url)
		files.AssertExpectations(t)
		store.AssertExpectations(t)
		audit.AssertExpectations(t)
	})

	t.Run("file service not configured", func(t *testing.T) {
		serviceWithoutFiles := NewService(store, nil, nil, audit)
		
		url, err := serviceWithoutFiles.UploadProfilePicture(ctx, "user1", imageData, mimeType)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file service not configured")
		assert.Empty(t, url)
	})
}

func TestFilterProfileByPrivacy(t *testing.T) {
	baseProfile := UserProfile{
		ID:       "user1",
		Name:     "John Doe",
		Email:    "john@example.com",
		Bio:      stringPtr("My bio"),
		Location: &Location{City: stringPtr("NYC")},
		Privacy: PrivacySettings{
			ProfileVisibility: "PUBLIC",
			ShowEmail:         false,
			ShowLocation:      true,
		},
	}

	t.Run("owner can see all fields", func(t *testing.T) {
		result := filterProfileByPrivacy(baseProfile, "user1", []string{})
		
		assert.Equal(t, "john@example.com", result.Email)
		assert.NotNil(t, result.Location)
		assert.NotNil(t, result.Bio)
	})

	t.Run("public profile with email hidden", func(t *testing.T) {
		result := filterProfileByPrivacy(baseProfile, "user2", []string{})
		
		assert.Empty(t, result.Email)
		assert.NotNil(t, result.Location)
		assert.NotNil(t, result.Bio)
	})

	t.Run("private profile hides sensitive data", func(t *testing.T) {
		privateProfile := baseProfile
		privateProfile.Privacy.ProfileVisibility = "PRIVATE"
		
		result := filterProfileByPrivacy(privateProfile, "user2", []string{})
		
		assert.Empty(t, result.Email)
		assert.Nil(t, result.Location)
		assert.Nil(t, result.Bio)
	})

	t.Run("volunteers only profile with location hidden", func(t *testing.T) {
		volProfile := baseProfile
		volProfile.Privacy.ProfileVisibility = "VOLUNTEERS_ONLY"
		volProfile.Privacy.ShowLocation = false
		
		result := filterProfileByPrivacy(volProfile, "user2", []string{})
		
		assert.Empty(t, result.Email)
		assert.Nil(t, result.Location)
		assert.NotNil(t, result.Bio)
	})
}

func TestService_SearchUsers(t *testing.T) {
	service, store, _, _, _ := createTestService()
	ctx := context.Background()

	filter := UserSearchFilter{
		Skills: []string{"JavaScript"},
	}

	users := []UserProfile{
		{
			ID:    "user1",
			Name:  "John Doe",
			Email: "john@example.com",
			Privacy: PrivacySettings{
				ProfileVisibility: "PUBLIC",
				ShowEmail:         false,
				ShowLocation:      true,
			},
		},
		{
			ID:    "user2",
			Name:  "Jane Smith",
			Email: "jane@example.com",
			Privacy: PrivacySettings{
				ProfileVisibility: "PRIVATE",
				ShowEmail:         false,
				ShowLocation:      false,
			},
		},
	}

	t.Run("successful user search with privacy filtering", func(t *testing.T) {
		store.On("SearchUsers", ctx, filter, 10, 0).Return(users, nil).Once()

		result, err := service.SearchUsers(ctx, filter, 10, 0)
		
		require.NoError(t, err)
		assert.Len(t, result, 2)
		
		// Check that privacy filtering is applied
		assert.Empty(t, result[0].Email) // Email should be hidden
		assert.Empty(t, result[1].Email) // Email should be hidden
		
		store.AssertExpectations(t)
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}