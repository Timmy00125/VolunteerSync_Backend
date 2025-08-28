package graph

import (
	"github.com/volunteersync/backend/internal/core/event"
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

// Event converters

// toDomainCreateEventInput converts GraphQL CreateEventInput to domain CreateEventInput
func toDomainCreateEventInput(input model.CreateEventInput) event.CreateEventInput {
	result := event.CreateEventInput{
		Title:            input.Title,
		Description:      input.Description,
		ShortDescription: input.ShortDescription,
		StartTime:        input.StartTime,
		EndTime:          input.EndTime,
		Location: event.EventLocationInput{
			Name:         input.Location.Name,
			Address:      input.Location.Address,
			City:         input.Location.City,
			State:        input.Location.State,
			Country:      input.Location.Country,
			ZipCode:      input.Location.ZipCode,
			Instructions: input.Location.Instructions,
			IsRemote:     input.Location.IsRemote,
		},
		Capacity: event.EventCapacityInput{
			Minimum:         input.Capacity.Minimum,
			Maximum:         input.Capacity.Maximum,
			WaitlistEnabled: input.Capacity.WaitlistEnabled,
		},
		Category:       convertGraphQLEventCategory(input.Category),
		TimeCommitment: convertGraphQLTimeCommitmentType(input.TimeCommitment),
		Tags:           input.Tags,
		RegistrationSettings: event.RegistrationSettingsInput{
			OpensAt:              input.RegistrationSettings.OpensAt,
			ClosesAt:             input.RegistrationSettings.ClosesAt,
			RequiresApproval:     input.RegistrationSettings.RequiresApproval,
			AllowWaitlist:        input.RegistrationSettings.AllowWaitlist,
			ConfirmationRequired: input.RegistrationSettings.ConfirmationRequired,
			CancellationDeadline: input.RegistrationSettings.CancellationDeadline,
		},
	}

	// Handle coordinates
	if input.Location.Coordinates != nil {
		result.Location.Coordinates = &event.CoordinatesInput{
			Latitude:  input.Location.Coordinates.Lat,
			Longitude: input.Location.Coordinates.Lng,
		}
	}

	// Handle requirements
	if input.Requirements != nil {
		result.Requirements = &event.EventRequirementsInput{
			MinimumAge:           input.Requirements.MinimumAge,
			BackgroundCheck:      input.Requirements.BackgroundCheck,
			PhysicalRequirements: input.Requirements.PhysicalRequirements,
			Interests:            input.Requirements.Interests,
		}

		// Convert skill requirements
		for _, skill := range input.Requirements.Skills {
			result.Requirements.Skills = append(result.Requirements.Skills, event.SkillRequirementInput{
				Skill:       skill.Skill,
				Proficiency: convertGraphQLSkillProficiency(skill.Proficiency),
				Required:    skill.Required,
			})
		}

		// Convert training requirements
		for _, training := range input.Requirements.Training {
			result.Requirements.Training = append(result.Requirements.Training, event.TrainingRequirementInput{
				Name:                training.Name,
				Description:         training.Description,
				Required:            training.Required,
				ProvidedByOrganizer: training.ProvidedByOrganizer,
			})
		}
	}

	// Handle recurrence rule
	if input.RecurrenceRule != nil {
		result.RecurrenceRule = &event.RecurrenceRuleInput{
			Frequency:       convertGraphQLRecurrenceFrequency(input.RecurrenceRule.Frequency),
			Interval:        input.RecurrenceRule.Interval,
			DaysOfWeek:      convertGraphQLDaysOfWeek(input.RecurrenceRule.DaysOfWeek),
			DayOfMonth:      input.RecurrenceRule.DayOfMonth,
			EndDate:         input.RecurrenceRule.EndDate,
			OccurrenceCount: input.RecurrenceRule.OccurrenceCount,
		}
	}

	return result
}

// toDomainUpdateEventInput converts GraphQL UpdateEventInput to domain UpdateEventInput
func toDomainUpdateEventInput(input model.UpdateEventInput) event.UpdateEventInput {
	result := event.UpdateEventInput{
		Title:            input.Title,
		Description:      input.Description,
		ShortDescription: input.ShortDescription,
		Tags:             input.Tags,
	}

	if input.Category != nil {
		category := convertGraphQLEventCategory(*input.Category)
		result.Category = &category
	}

	if input.Location != nil {
		result.Location = &event.EventLocationInput{
			Name:         input.Location.Name,
			Address:      input.Location.Address,
			City:         input.Location.City,
			State:        input.Location.State,
			Country:      input.Location.Country,
			ZipCode:      input.Location.ZipCode,
			Instructions: input.Location.Instructions,
			IsRemote:     input.Location.IsRemote,
		}

		if input.Location.Coordinates != nil {
			result.Location.Coordinates = &event.CoordinatesInput{
				Latitude:  input.Location.Coordinates.Lat,
				Longitude: input.Location.Coordinates.Lng,
			}
		}
	}

	if input.Requirements != nil {
		result.Requirements = &event.EventRequirementsInput{
			MinimumAge:           input.Requirements.MinimumAge,
			BackgroundCheck:      input.Requirements.BackgroundCheck,
			PhysicalRequirements: input.Requirements.PhysicalRequirements,
			Interests:            input.Requirements.Interests,
		}

		for _, skill := range input.Requirements.Skills {
			result.Requirements.Skills = append(result.Requirements.Skills, event.SkillRequirementInput{
				Skill:       skill.Skill,
				Proficiency: convertGraphQLSkillProficiency(skill.Proficiency),
				Required:    skill.Required,
			})
		}

		for _, training := range input.Requirements.Training {
			result.Requirements.Training = append(result.Requirements.Training, event.TrainingRequirementInput{
				Name:                training.Name,
				Description:         training.Description,
				Required:            training.Required,
				ProvidedByOrganizer: training.ProvidedByOrganizer,
			})
		}
	}

	return result
}

// toGraphQLEvent converts domain Event to GraphQL Event
func toGraphQLEvent(e *event.Event) *model.Event {
	result := &model.Event{
		ID:               e.ID,
		Title:            e.Title,
		Description:      e.Description,
		ShortDescription: e.ShortDescription,
		OrganizerID:      e.OrganizerID,
		Status:           convertDomainEventStatus(e.Status),
		StartTime:        e.StartTime,
		EndTime:          e.EndTime,
		Location: &model.EventLocation{
			Name:         e.Location.Name,
			Address:      e.Location.Address,
			City:         e.Location.City,
			State:        e.Location.State,
			Country:      e.Location.Country,
			ZipCode:      e.Location.ZipCode,
			Instructions: e.Location.Instructions,
			IsRemote:     e.Location.IsRemote,
		},
		Capacity: &model.EventCapacity{
			Minimum:         e.Capacity.Minimum,
			Maximum:         e.Capacity.Maximum,
			Current:         e.Capacity.Current,
			WaitlistEnabled: e.Capacity.WaitlistEnabled,
		},
		Requirements: &model.EventRequirements{
			MinimumAge:           e.Requirements.MinimumAge,
			BackgroundCheck:      e.Requirements.BackgroundCheck,
			PhysicalRequirements: e.Requirements.PhysicalRequirements,
			Interests:            e.Requirements.Interests,
		},
		Category:       convertDomainEventCategory(e.Category),
		TimeCommitment: convertDomainTimeCommitmentType(e.TimeCommitment),
		Tags:           e.Tags,
		Slug:           e.Slug,
		ShareURL:       e.ShareURL,
		RegistrationSettings: &model.RegistrationSettings{
			OpensAt:              e.RegistrationSettings.OpensAt,
			ClosesAt:             e.RegistrationSettings.ClosesAt,
			RequiresApproval:     e.RegistrationSettings.RequiresApproval,
			AllowWaitlist:        e.RegistrationSettings.AllowWaitlist,
			ConfirmationRequired: e.RegistrationSettings.ConfirmationRequired,
			CancellationDeadline: e.RegistrationSettings.CancellationDeadline,
		},
		Images:               []*model.EventImage{},
		Announcements:        []*model.EventAnnouncement{},
		CreatedAt:            e.CreatedAt,
		UpdatedAt:            e.UpdatedAt,
		CurrentRegistrations: e.Capacity.Current,
		AvailableSpots:       e.Capacity.Maximum - e.Capacity.Current,
		IsAtCapacity:         e.Capacity.Current >= e.Capacity.Maximum,
		CanRegister:          e.Status == event.EventStatusPublished && e.Capacity.Current < e.Capacity.Maximum,
	}

	// Handle coordinates
	if e.Location.Coordinates != nil {
		result.Location.Coordinates = &model.Coordinates{
			Lat: e.Location.Coordinates.Latitude,
			Lng: e.Location.Coordinates.Longitude,
		}
	}

	// Convert skill requirements
	for _, skill := range e.Requirements.Skills {
		result.Requirements.Skills = append(result.Requirements.Skills, &model.SkillRequirement{
			ID:          skill.ID,
			Skill:       skill.Skill,
			Proficiency: convertDomainSkillProficiency(skill.Proficiency),
			Required:    skill.Required,
		})
	}

	// Convert training requirements
	for _, training := range e.Requirements.Training {
		result.Requirements.Training = append(result.Requirements.Training, &model.TrainingRequirement{
			ID:                  training.ID,
			Name:                training.Name,
			Description:         training.Description,
			Required:            training.Required,
			ProvidedByOrganizer: training.ProvidedByOrganizer,
		})
	}

	// Handle recurrence rule
	if e.RecurrenceRule != nil {
		result.RecurrenceRule = &model.RecurrenceRule{
			Frequency:       convertDomainRecurrenceFrequency(e.RecurrenceRule.Frequency),
			Interval:        e.RecurrenceRule.Interval,
			DaysOfWeek:      convertDomainDaysOfWeek(e.RecurrenceRule.DaysOfWeek),
			DayOfMonth:      e.RecurrenceRule.DayOfMonth,
			EndDate:         e.RecurrenceRule.EndDate,
			OccurrenceCount: e.RecurrenceRule.OccurrenceCount,
		}
	}

	return result
}

// Enum converters

func convertGraphQLEventCategory(category model.EventCategory) event.EventCategory {
	switch category {
	case model.EventCategoryCommunityService:
		return event.EventCategoryCommunityService
	case model.EventCategoryEnvironmental:
		return event.EventCategoryEnvironment
	case model.EventCategoryEducation:
		return event.EventCategoryEducation
	case model.EventCategoryHealthWellness:
		return event.EventCategoryHealth
	case model.EventCategoryDisasterRelief:
		return event.EventCategoryDisasterRelief
	case model.EventCategoryAnimalWelfare:
		return event.EventCategoryAnimalWelfare
	case model.EventCategoryArtsCulture:
		return event.EventCategoryArtsCulture
	case model.EventCategoryTechnology:
		return event.EventCategoryTechnology
	case model.EventCategorySportsRecreation:
		return event.EventCategorySportsRecreation
	case model.EventCategoryFoodHunger:
		return event.EventCategoryFoodSecurity
	case model.EventCategoryYouthDevelopment:
		return event.EventCategoryYouthMentoring
	case model.EventCategorySeniorCare:
		return event.EventCategorySeniorCare
	case model.EventCategoryHomelessServices:
		return event.EventCategoryCommunityService // Map to community service
	case model.EventCategoryFundraising:
		return event.EventCategoryCommunityService // Map to community service
	case model.EventCategoryAdvocacy:
		return event.EventCategoryCommunityService // Map to community service
	default:
		return event.EventCategoryCommunityService
	}
}

func convertDomainEventCategory(category event.EventCategory) model.EventCategory {
	switch category {
	case event.EventCategoryCommunityService:
		return model.EventCategoryCommunityService
	case event.EventCategoryEnvironment:
		return model.EventCategoryEnvironmental
	case event.EventCategoryEducation:
		return model.EventCategoryEducation
	case event.EventCategoryHealth:
		return model.EventCategoryHealthWellness
	case event.EventCategoryDisasterRelief:
		return model.EventCategoryDisasterRelief
	case event.EventCategoryAnimalWelfare:
		return model.EventCategoryAnimalWelfare
	case event.EventCategoryArtsCulture:
		return model.EventCategoryArtsCulture
	case event.EventCategoryTechnology:
		return model.EventCategoryTechnology
	case event.EventCategorySportsRecreation:
		return model.EventCategorySportsRecreation
	case event.EventCategoryFoodSecurity:
		return model.EventCategoryFoodHunger
	case event.EventCategoryYouthMentoring:
		return model.EventCategoryYouthDevelopment
	case event.EventCategorySeniorCare:
		return model.EventCategorySeniorCare
	default:
		return model.EventCategoryCommunityService
	}
}

func convertGraphQLTimeCommitmentType(timeCommitment model.TimeCommitmentType) event.TimeCommitmentType {
	switch timeCommitment {
	case model.TimeCommitmentTypeOneTime:
		return event.TimeCommitmentOneTime
	case model.TimeCommitmentTypeWeekly:
		return event.TimeCommitmentShortTerm // Map weekly to short term
	case model.TimeCommitmentTypeMonthly:
		return event.TimeCommitmentMediumTerm // Map monthly to medium term
	case model.TimeCommitmentTypeSeasonal:
		return event.TimeCommitmentLongTerm // Map seasonal to long term
	case model.TimeCommitmentTypeOngoing:
		return event.TimeCommitmentOngoing
	default:
		return event.TimeCommitmentOneTime
	}
}

func convertDomainTimeCommitmentType(timeCommitment event.TimeCommitmentType) model.TimeCommitmentType {
	switch timeCommitment {
	case event.TimeCommitmentOneTime:
		return model.TimeCommitmentTypeOneTime
	case event.TimeCommitmentShortTerm:
		return model.TimeCommitmentTypeWeekly
	case event.TimeCommitmentMediumTerm:
		return model.TimeCommitmentTypeMonthly
	case event.TimeCommitmentLongTerm:
		return model.TimeCommitmentTypeSeasonal
	case event.TimeCommitmentOngoing:
		return model.TimeCommitmentTypeOngoing
	default:
		return model.TimeCommitmentTypeOneTime
	}
}

func convertGraphQLSkillProficiency(proficiency model.SkillProficiency) event.SkillProficiency {
	switch proficiency {
	case model.SkillProficiencyBeginner:
		return event.SkillProficiencyBeginner
	case model.SkillProficiencyIntermediate:
		return event.SkillProficiencyIntermediate
	case model.SkillProficiencyAdvanced:
		return event.SkillProficiencyAdvanced
	case model.SkillProficiencyExpert:
		return event.SkillProficiencyExpert
	default:
		return event.SkillProficiencyBeginner
	}
}

func convertDomainSkillProficiency(proficiency event.SkillProficiency) model.SkillProficiency {
	switch proficiency {
	case event.SkillProficiencyBeginner:
		return model.SkillProficiencyBeginner
	case event.SkillProficiencyIntermediate:
		return model.SkillProficiencyIntermediate
	case event.SkillProficiencyAdvanced:
		return model.SkillProficiencyAdvanced
	case event.SkillProficiencyExpert:
		return model.SkillProficiencyExpert
	default:
		return model.SkillProficiencyBeginner
	}
}

func convertDomainEventStatus(status event.EventStatus) model.EventStatus {
	switch status {
	case event.EventStatusDraft:
		return model.EventStatusDraft
	case event.EventStatusPublished:
		return model.EventStatusPublished
	case event.EventStatusCancelled:
		return model.EventStatusCancelled
	case event.EventStatusCompleted:
		return model.EventStatusCompleted
	case event.EventStatusArchived:
		return model.EventStatusArchived
	default:
		return model.EventStatusDraft
	}
}

func convertGraphQLRecurrenceFrequency(frequency model.RecurrenceFrequency) event.RecurrenceFrequency {
	switch frequency {
	case model.RecurrenceFrequencyDaily:
		return event.RecurrenceFrequencyDaily
	case model.RecurrenceFrequencyWeekly:
		return event.RecurrenceFrequencyWeekly
	case model.RecurrenceFrequencyMonthly:
		return event.RecurrenceFrequencyMonthly
	case model.RecurrenceFrequencyYearly:
		return event.RecurrenceFrequencyYearly
	default:
		return event.RecurrenceFrequencyWeekly
	}
}

func convertDomainRecurrenceFrequency(frequency event.RecurrenceFrequency) model.RecurrenceFrequency {
	switch frequency {
	case event.RecurrenceFrequencyDaily:
		return model.RecurrenceFrequencyDaily
	case event.RecurrenceFrequencyWeekly:
		return model.RecurrenceFrequencyWeekly
	case event.RecurrenceFrequencyMonthly:
		return model.RecurrenceFrequencyMonthly
	case event.RecurrenceFrequencyYearly:
		return model.RecurrenceFrequencyYearly
	default:
		return model.RecurrenceFrequencyWeekly
	}
}

func convertGraphQLDaysOfWeek(days []model.DayOfWeek) []event.DayOfWeek {
	result := make([]event.DayOfWeek, len(days))
	for i, day := range days {
		switch day {
		case model.DayOfWeekSunday:
			result[i] = event.DayOfWeekSunday
		case model.DayOfWeekMonday:
			result[i] = event.DayOfWeekMonday
		case model.DayOfWeekTuesday:
			result[i] = event.DayOfWeekTuesday
		case model.DayOfWeekWednesday:
			result[i] = event.DayOfWeekWednesday
		case model.DayOfWeekThursday:
			result[i] = event.DayOfWeekThursday
		case model.DayOfWeekFriday:
			result[i] = event.DayOfWeekFriday
		case model.DayOfWeekSaturday:
			result[i] = event.DayOfWeekSaturday
		}
	}
	return result
}

func convertDomainDaysOfWeek(days []event.DayOfWeek) []model.DayOfWeek {
	result := make([]model.DayOfWeek, len(days))
	for i, day := range days {
		switch day {
		case event.DayOfWeekSunday:
			result[i] = model.DayOfWeekSunday
		case event.DayOfWeekMonday:
			result[i] = model.DayOfWeekMonday
		case event.DayOfWeekTuesday:
			result[i] = model.DayOfWeekTuesday
		case event.DayOfWeekWednesday:
			result[i] = model.DayOfWeekWednesday
		case event.DayOfWeekThursday:
			result[i] = model.DayOfWeekThursday
		case event.DayOfWeekFriday:
			result[i] = model.DayOfWeekFriday
		case event.DayOfWeekSaturday:
			result[i] = model.DayOfWeekSaturday
		}
	}
	return result
}
