package graph

import (
	usercore "github.com/volunteersync/backend/internal/core/user"
	"github.com/volunteersync/backend/internal/graph/model"
)

// toDomainUpdateProfile converts GraphQL UpdateProfileInput to domain UpdateProfileInput
func toDomainUpdateProfile(input model.UpdateProfileInput) usercore.UpdateProfileInput {
	result := usercore.UpdateProfileInput{
		Name: input.Name,
		Bio:  input.Bio,
	}

	if input.Location != nil {
		result.Location = &usercore.Location{
			City:    input.Location.City,
			State:   input.Location.State,
			Country: input.Location.Country,
			Lat:     input.Location.Lat,
			Lng:     input.Location.Lng,
		}
	}

	return result
}

// toDomainSkillInput converts GraphQL SkillInput to domain SkillInput
func toDomainSkillInput(input model.SkillInput) usercore.SkillInput {
	return usercore.SkillInput{
		Name:        input.Name,
		Proficiency: string(input.Proficiency),
	}
}

// toDomainPrivacyInput converts GraphQL PrivacySettingsInput to domain PrivacySettings
func toDomainPrivacyInput(input model.PrivacySettingsInput) usercore.PrivacySettings {
	result := usercore.PrivacySettings{}

	if input.ProfileVisibility != nil {
		result.ProfileVisibility = string(*input.ProfileVisibility)
	}
	if input.ShowEmail != nil {
		result.ShowEmail = *input.ShowEmail
	}
	if input.ShowLocation != nil {
		result.ShowLocation = *input.ShowLocation
	}
	if input.AllowMessaging != nil {
		result.AllowMessaging = *input.AllowMessaging
	}

	return result
}

// toDomainNotifInput converts GraphQL NotificationPreferencesInput to domain NotificationPreferences
func toDomainNotifInput(input model.NotificationPreferencesInput) usercore.NotificationPreferences {
	result := usercore.NotificationPreferences{}

	if input.EmailNotifications != nil {
		result.EmailNotifications = *input.EmailNotifications
	}
	if input.PushNotifications != nil {
		result.PushNotifications = *input.PushNotifications
	}
	if input.SmsNotifications != nil {
		result.SMSNotifications = *input.SmsNotifications
	}
	if input.EventReminders != nil {
		result.EventReminders = *input.EventReminders
	}
	if input.NewOpportunities != nil {
		result.NewOpportunities = *input.NewOpportunities
	}
	if input.NewsletterSubscription != nil {
		result.NewsletterSubscription = *input.NewsletterSubscription
	}

	return result
}

// toDomainSearchFilter converts GraphQL UserSearchFilter to domain UserSearchFilter
func toDomainSearchFilter(filter model.UserSearchFilter) usercore.UserSearchFilter {
	result := usercore.UserSearchFilter{
		Skills:      filter.Skills,
		InterestIDs: filter.Interests,
	}

	if filter.Location != nil {
		result.Location = &usercore.Location{
			City:    filter.Location.City,
			State:   filter.Location.State,
			Country: filter.Location.Country,
			Lat:     filter.Location.Lat,
			Lng:     filter.Location.Lng,
		}
	}

	if filter.Availability != nil {
		avail := string(*filter.Availability)
		result.Availability = &avail
	}

	if filter.Experience != nil {
		exp := string(*filter.Experience)
		result.Experience = &exp
	}

	return result
}

// toGraphUser converts domain UserProfile to GraphQL User
func toGraphUser(profile *usercore.UserProfile) *model.User {
	if profile == nil {
		return nil
	}

	user := &model.User{
		ID:             profile.ID,
		Email:          profile.Email,
		Name:           profile.Name,
		Bio:            profile.Bio,
		ProfilePicture: profile.ProfilePictureURL,
		Roles:          profile.Roles,
		IsVerified:     profile.IsVerified,
		CreatedAt:      profile.CreatedAt,
		UpdatedAt:      profile.UpdatedAt,
		JoinedAt:       profile.CreatedAt, // Using CreatedAt as JoinedAt
		LastActiveAt:   profile.LastActiveAt,
		EmailVerified:  profile.IsVerified, // Assuming email verification aligns with general verification
	}

	// Convert location
	if profile.Location != nil {
		user.Location = &model.Location{
			City:    profile.Location.City,
			State:   profile.Location.State,
			Country: profile.Location.Country,
		}
		if profile.Location.Lat != nil && profile.Location.Lng != nil {
			user.Location.Coordinates = &model.Coordinates{
				Lat: *profile.Location.Lat,
				Lng: *profile.Location.Lng,
			}
		}
	}

	// Convert interests
	user.Interests = make([]*model.Interest, len(profile.Interests))
	for i, interest := range profile.Interests {
		user.Interests[i] = &model.Interest{
			ID:       interest.ID,
			Name:     interest.Name,
			Category: model.InterestCategory(interest.Category),
		}
	}

	// Convert skills
	user.Skills = make([]*model.Skill, len(profile.Skills))
	for i, skill := range profile.Skills {
		user.Skills[i] = &model.Skill{
			ID:          skill.ID,
			Name:        skill.Name,
			Proficiency: model.SkillProficiency(skill.Proficiency),
			Verified:    skill.Verified,
		}
	}

	// Create public profile
	user.PublicProfile = toGraphPublicProfile(profile)

	return user
}

// toGraphPublicProfile converts domain UserProfile to GraphQL PublicProfile
func toGraphPublicProfile(profile *usercore.UserProfile) *model.PublicProfile {
	if profile == nil {
		return nil
	}

	publicProfile := &model.PublicProfile{
		ID:             profile.ID,
		Name:           profile.Name,
		Bio:            profile.Bio,
		ProfilePicture: profile.ProfilePictureURL,
	}

	// Only include location if privacy allows
	if profile.Privacy.ShowLocation && profile.Location != nil {
		publicProfile.Location = &model.Location{
			City:    profile.Location.City,
			State:   profile.Location.State,
			Country: profile.Location.Country,
		}
		if profile.Location.Lat != nil && profile.Location.Lng != nil {
			publicProfile.Location.Coordinates = &model.Coordinates{
				Lat: *profile.Location.Lat,
				Lng: *profile.Location.Lng,
			}
		}
	}

	// Convert interests
	publicProfile.Interests = make([]*model.Interest, len(profile.Interests))
	for i, interest := range profile.Interests {
		publicProfile.Interests[i] = &model.Interest{
			ID:       interest.ID,
			Name:     interest.Name,
			Category: model.InterestCategory(interest.Category),
		}
	}

	// Convert skills
	publicProfile.Skills = make([]*model.Skill, len(profile.Skills))
	for i, skill := range profile.Skills {
		publicProfile.Skills[i] = &model.Skill{
			ID:          skill.ID,
			Name:        skill.Name,
			Proficiency: model.SkillProficiency(skill.Proficiency),
			Verified:    skill.Verified,
		}
	}

	// Add placeholder volunteer stats
	publicProfile.VolunteerStats = &model.VolunteerStats{
		Hours:              0, // These would come from a separate service in a real implementation
		EventsParticipated: 0,
	}

	return publicProfile
}
