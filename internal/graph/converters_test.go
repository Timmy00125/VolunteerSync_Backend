package graph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	usercore "github.com/volunteersync/backend/internal/core/user"
	"github.com/volunteersync/backend/internal/graph/model"
)

func TestToDomainUpdateProfile(t *testing.T) {
	tests := []struct {
		name     string
		input    model.UpdateProfileInput
		expected usercore.UpdateProfileInput
	}{
		{
			name: "complete profile update",
			input: model.UpdateProfileInput{
				Name: stringPtr("John Doe"),
				Bio:  stringPtr("Software Engineer"),
				Location: &model.LocationInput{
					City:    stringPtr("San Francisco"),
					State:   stringPtr("CA"),
					Country: stringPtr("USA"),
					Lat:     float64Ptr(37.7749),
					Lng:     float64Ptr(-122.4194),
				},
			},
			expected: usercore.UpdateProfileInput{
				Name: stringPtr("John Doe"),
				Bio:  stringPtr("Software Engineer"),
				Location: &usercore.Location{
					City:    stringPtr("San Francisco"),
					State:   stringPtr("CA"),
					Country: stringPtr("USA"),
					Lat:     float64Ptr(37.7749),
					Lng:     float64Ptr(-122.4194),
				},
			},
		},
		{
			name: "minimal profile update",
			input: model.UpdateProfileInput{
				Name: stringPtr("Jane Smith"),
			},
			expected: usercore.UpdateProfileInput{
				Name: stringPtr("Jane Smith"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDomainUpdateProfile(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToDomainSkillInput(t *testing.T) {
	tests := []struct {
		name     string
		input    model.SkillInput
		expected usercore.SkillInput
	}{
		{
			name: "advanced Go skill",
			input: model.SkillInput{
				Name:        "Go",
				Proficiency: model.SkillProficiencyAdvanced,
			},
			expected: usercore.SkillInput{
				Name:        "Go",
				Proficiency: "ADVANCED",
			},
		},
		{
			name: "beginner JavaScript skill",
			input: model.SkillInput{
				Name:        "JavaScript",
				Proficiency: model.SkillProficiencyBeginner,
			},
			expected: usercore.SkillInput{
				Name:        "JavaScript",
				Proficiency: "BEGINNER",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDomainSkillInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToDomainPrivacyInput(t *testing.T) {
	tests := []struct {
		name     string
		input    model.PrivacySettingsInput
		expected usercore.PrivacySettings
	}{
		{
			name: "public profile settings",
			input: model.PrivacySettingsInput{
				ProfileVisibility: &[]model.ProfileVisibility{model.ProfileVisibilityPublic}[0],
				ShowEmail:         &[]bool{true}[0],
				ShowLocation:      &[]bool{true}[0],
				AllowMessaging:    &[]bool{true}[0],
			},
			expected: usercore.PrivacySettings{
				ProfileVisibility: "PUBLIC",
				ShowEmail:         true,
				ShowLocation:      true,
				AllowMessaging:    true,
			},
		},
		{
			name: "private profile settings",
			input: model.PrivacySettingsInput{
				ProfileVisibility: &[]model.ProfileVisibility{model.ProfileVisibilityPrivate}[0],
				ShowEmail:         &[]bool{false}[0],
				ShowLocation:      &[]bool{false}[0],
				AllowMessaging:    &[]bool{false}[0],
			},
			expected: usercore.PrivacySettings{
				ProfileVisibility: "PRIVATE",
				ShowEmail:         false,
				ShowLocation:      false,
				AllowMessaging:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDomainPrivacyInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToGraphUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    *usercore.UserProfile
		expected *model.User
	}{
		{
			name: "complete user profile",
			input: &usercore.UserProfile{
				ID:                "user-123",
				Name:              "John Doe",
				Email:             "john@example.com",
				Bio:               stringPtr("Software Engineer"),
				ProfilePictureURL: stringPtr("https://example.com/pic.jpg"),
				Location: &usercore.Location{
					City:    stringPtr("San Francisco"),
					State:   stringPtr("CA"),
					Country: stringPtr("USA"),
					Lat:     float64Ptr(37.7749),
					Lng:     float64Ptr(-122.4194),
				},
				Interests: []usercore.Interest{
					{ID: "1", Name: "Technology", Category: "TECHNOLOGY"},
				},
				Skills: []usercore.Skill{
					{ID: "1", Name: "Go", Proficiency: "ADVANCED", Verified: true},
				},
				Privacy: usercore.PrivacySettings{
					ProfileVisibility: "PUBLIC",
					ShowLocation:      true,
				},
				Roles:        []string{"user"},
				IsVerified:   true,
				CreatedAt:    now,
				UpdatedAt:    now,
				LastActiveAt: &now,
			},
			expected: &model.User{
				ID:             "user-123",
				Name:           "John Doe",
				Email:          "john@example.com",
				Bio:            stringPtr("Software Engineer"),
				ProfilePicture: stringPtr("https://example.com/pic.jpg"),
				Location: &model.Location{
					City:    stringPtr("San Francisco"),
					State:   stringPtr("CA"),
					Country: stringPtr("USA"),
					Coordinates: &model.Coordinates{
						Lat: 37.7749,
						Lng: -122.4194,
					},
				},
				Interests: []*model.Interest{
					{ID: "1", Name: "Technology", Category: model.InterestCategoryTechnology},
				},
				Skills: []*model.Skill{
					{ID: "1", Name: "Go", Proficiency: model.SkillProficiencyAdvanced, Verified: true},
				},
				Roles:         []string{"user"},
				IsVerified:    true,
				EmailVerified: true,
				CreatedAt:     now,
				UpdatedAt:     now,
				JoinedAt:      now,
				LastActiveAt:  &now,
				PublicProfile: &model.PublicProfile{
					ID:             "user-123",
					Name:           "John Doe",
					Bio:            stringPtr("Software Engineer"),
					ProfilePicture: stringPtr("https://example.com/pic.jpg"),
					Location: &model.Location{
						City:    stringPtr("San Francisco"),
						State:   stringPtr("CA"),
						Country: stringPtr("USA"),
						Coordinates: &model.Coordinates{
							Lat: 37.7749,
							Lng: -122.4194,
						},
					},
					Interests: []*model.Interest{
						{ID: "1", Name: "Technology", Category: model.InterestCategoryTechnology},
					},
					Skills: []*model.Skill{
						{ID: "1", Name: "Go", Proficiency: model.SkillProficiencyAdvanced, Verified: true},
					},
					VolunteerStats: &model.VolunteerStats{
						Hours:              0,
						EventsParticipated: 0,
					},
				},
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toGraphUser(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToGraphPublicProfile(t *testing.T) {
	tests := []struct {
		name     string
		input    *usercore.UserProfile
		expected *model.PublicProfile
	}{
		{
			name: "public profile with location shown",
			input: &usercore.UserProfile{
				ID:                "user-123",
				Name:              "John Doe",
				Bio:               stringPtr("Software Engineer"),
				ProfilePictureURL: stringPtr("https://example.com/pic.jpg"),
				Location: &usercore.Location{
					City:    stringPtr("San Francisco"),
					State:   stringPtr("CA"),
					Country: stringPtr("USA"),
					Lat:     float64Ptr(37.7749),
					Lng:     float64Ptr(-122.4194),
				},
				Privacy: usercore.PrivacySettings{
					ShowLocation: true,
				},
				Interests: []usercore.Interest{
					{ID: "1", Name: "Technology", Category: "TECHNOLOGY"},
				},
				Skills: []usercore.Skill{
					{ID: "1", Name: "Go", Proficiency: "ADVANCED", Verified: true},
				},
			},
			expected: &model.PublicProfile{
				ID:             "user-123",
				Name:           "John Doe",
				Bio:            stringPtr("Software Engineer"),
				ProfilePicture: stringPtr("https://example.com/pic.jpg"),
				Location: &model.Location{
					City:    stringPtr("San Francisco"),
					State:   stringPtr("CA"),
					Country: stringPtr("USA"),
					Coordinates: &model.Coordinates{
						Lat: 37.7749,
						Lng: -122.4194,
					},
				},
				Interests: []*model.Interest{
					{ID: "1", Name: "Technology", Category: model.InterestCategoryTechnology},
				},
				Skills: []*model.Skill{
					{ID: "1", Name: "Go", Proficiency: model.SkillProficiencyAdvanced, Verified: true},
				},
				VolunteerStats: &model.VolunteerStats{
					Hours:              0,
					EventsParticipated: 0,
				},
			},
		},
		{
			name: "profile with location hidden",
			input: &usercore.UserProfile{
				ID:   "user-456",
				Name: "Jane Smith",
				Location: &usercore.Location{
					City: stringPtr("New York"),
				},
				Privacy: usercore.PrivacySettings{
					ShowLocation: false,
				},
			},
			expected: &model.PublicProfile{
				ID:        "user-456",
				Name:      "Jane Smith",
				Location:  nil, // Should be nil when ShowLocation is false
				Interests: []*model.Interest{},
				Skills:    []*model.Skill{},
				VolunteerStats: &model.VolunteerStats{
					Hours:              0,
					EventsParticipated: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toGraphPublicProfile(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
