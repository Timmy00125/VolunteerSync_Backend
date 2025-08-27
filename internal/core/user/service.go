package user

import (
	"context"
	"fmt"
)

// UserStore abstracts persistence for user domain.
type UserStore interface {
	GetProfile(ctx context.Context, userID string) (*UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*UserProfile, error)
	SetProfilePicture(ctx context.Context, userID, url string) error

	ReplaceInterests(ctx context.Context, userID string, interestIDs []string) ([]Interest, error)
	ListInterests(ctx context.Context) ([]Interest, error)
	ListUserInterests(ctx context.Context, userID string) ([]Interest, error)

	AddSkill(ctx context.Context, userID string, in SkillInput) (*Skill, error)
	RemoveSkill(ctx context.Context, userID, skillID string) error
	ListSkills(ctx context.Context, userID string) ([]Skill, error)

	UpdatePrivacy(ctx context.Context, userID string, in PrivacySettings) (PrivacySettings, error)
	UpdateNotifications(ctx context.Context, userID string, in NotificationPreferences) (NotificationPreferences, error)

	GetUserRoles(ctx context.Context, userID string) ([]string, error)
	SetUserRoles(ctx context.Context, userID string, roles []string, assignedBy string) error

	SearchUsers(ctx context.Context, filter UserSearchFilter, limit, offset int) ([]UserProfile, error)

	LogActivity(ctx context.Context, log ActivityLog) error
	ListActivityLogs(ctx context.Context, userID string, limit, offset int) ([]ActivityLog, error)
}

// FileService abstracts file storage.
type FileService interface {
	// SaveProfileImage stores image bytes and returns public URL and storage path key.
	SaveProfileImage(ctx context.Context, userID string, data []byte, mime string) (url, storagePath string, err error)
	// Delete removes a previously stored file by storage path key.
	Delete(ctx context.Context, storagePath string) error
}

// NotificationService placeholder for cross-system notifications.
type NotificationService interface {
	NotifyProfileUpdated(ctx context.Context, userID string) error
}

// AuditLogger records important security-relevant actions.
type AuditLogger interface {
	Info(ctx context.Context, action string, details map[string]any)
	Warn(ctx context.Context, action string, details map[string]any)
}

// UserSearchFilter mirrors GraphQL input for service layer.
type UserSearchFilter struct {
	Skills       []string
	InterestIDs  []string
	Location     *Location
	Availability *string
	Experience   *string
}

// Service coordinates user domain operations.
type Service struct {
	store    UserStore
	files    FileService
	notifier NotificationService
	audit    AuditLogger
}

// NewService constructs a user Service.
func NewService(store UserStore, files FileService, notifier NotificationService, audit AuditLogger) *Service {
	return &Service{store: store, files: files, notifier: notifier, audit: audit}
}

// GetProfile returns a profile filtered per privacy for requester.
func (s *Service) GetProfile(ctx context.Context, userID, requesterID string, requesterRoles []string) (*UserProfile, error) {
	prof, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	filtered := filterProfileByPrivacy(*prof, requesterID, requesterRoles)
	return &filtered, nil
}

// GetProfileWithDetails returns profile and fills interests/skills for presentation.
func (s *Service) GetProfileWithDetails(ctx context.Context, userID, requesterID string, requesterRoles []string) (*UserProfile, error) {
	prof, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Load interests and skills
	if ints, err := s.store.ListUserInterests(ctx, userID); err == nil {
		prof.Interests = ints
	}
	if skills, err := s.store.ListSkills(ctx, userID); err == nil {
		prof.Skills = skills
	}
	filtered := filterProfileByPrivacy(*prof, requesterID, requesterRoles)
	return &filtered, nil
}

// UpdateProfile updates editable fields of the current user.
func (s *Service) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*UserProfile, error) {
	prof, err := s.store.UpdateProfile(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	if s.notifier != nil {
		_ = s.notifier.NotifyProfileUpdated(ctx, userID)
	}
	if s.audit != nil {
		s.audit.Info(ctx, "user.profile.update", map[string]any{"user_id": userID})
	}
	return prof, nil
}

// UpdateInterests replaces the interests set for a user.
func (s *Service) UpdateInterests(ctx context.Context, userID string, interestIDs []string) (*UserProfile, error) {
	ints, err := s.store.ReplaceInterests(ctx, userID, interestIDs)
	if err != nil {
		return nil, err
	}
	prof, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	prof.Interests = ints
	if s.audit != nil {
		s.audit.Info(ctx, "user.interests.update", map[string]any{"user_id": userID, "count": len(interestIDs)})
	}
	return prof, nil
}

// AddSkill adds a new skill.
func (s *Service) AddSkill(ctx context.Context, userID string, in SkillInput) (*UserProfile, error) {
	if _, err := s.store.AddSkill(ctx, userID, in); err != nil {
		return nil, err
	}
	prof, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	prof.Skills, _ = s.store.ListSkills(ctx, userID)
	if s.audit != nil {
		s.audit.Info(ctx, "user.skill.add", map[string]any{"user_id": userID, "name": in.Name})
	}
	return prof, nil
}

// RemoveSkill removes an existing skill by ID.
func (s *Service) RemoveSkill(ctx context.Context, userID, skillID string) (*UserProfile, error) {
	if err := s.store.RemoveSkill(ctx, userID, skillID); err != nil {
		return nil, err
	}
	prof, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	prof.Skills, _ = s.store.ListSkills(ctx, userID)
	if s.audit != nil {
		s.audit.Info(ctx, "user.skill.remove", map[string]any{"user_id": userID, "skill_id": skillID})
	}
	return prof, nil
}

// UpdatePrivacySettings updates privacy settings.
func (s *Service) UpdatePrivacySettings(ctx context.Context, userID string, in PrivacySettings) (*UserProfile, error) {
	_, err := s.store.UpdatePrivacy(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	prof, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		s.audit.Info(ctx, "user.privacy.update", map[string]any{"user_id": userID})
	}
	return prof, nil
}

// UpdateNotificationPreferences updates notification preferences.
func (s *Service) UpdateNotificationPreferences(ctx context.Context, userID string, in NotificationPreferences) (*UserProfile, error) {
	_, err := s.store.UpdateNotifications(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	prof, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		s.audit.Info(ctx, "user.notifications.update", map[string]any{"user_id": userID})
	}
	return prof, nil
}

// UploadProfilePicture processes and stores a profile image, updating the user's profile picture URL.
func (s *Service) UploadProfilePicture(ctx context.Context, userID string, data []byte, mime string) (string, error) {
	if s.files == nil {
		return "", fmt.Errorf("file service not configured")
	}
	url, _, err := s.files.SaveProfileImage(ctx, userID, data, mime)
	if err != nil {
		return "", err
	}
	if err := s.store.SetProfilePicture(ctx, userID, url); err != nil {
		return "", err
	}
	if s.audit != nil {
		s.audit.Info(ctx, "user.profile.picture.update", map[string]any{"user_id": userID})
	}
	return url, nil
}

// ListInterests enumerates all available interests.
func (s *Service) ListInterests(ctx context.Context) ([]Interest, error) {
	return s.store.ListInterests(ctx)
}

// ListActivityLogs returns activity logs for a user.
func (s *Service) ListActivityLogs(ctx context.Context, userID string, limit, offset int) ([]ActivityLog, error) {
	return s.store.ListActivityLogs(ctx, userID, limit, offset)
}

// SearchUsers returns profiles matching filter.
func (s *Service) SearchUsers(ctx context.Context, filter UserSearchFilter, limit, offset int) ([]UserProfile, error) {
	res, err := s.store.SearchUsers(ctx, filter, limit, offset)
	if err != nil {
		return nil, err
	}
	// Apply privacy filtering for public search results (no requester context)
	out := make([]UserProfile, 0, len(res))
	for _, p := range res {
		fp := filterProfileByPrivacy(p, "", nil)
		out = append(out, fp)
	}
	return out, nil
}

// Helper: filter profile fields based on privacy and requester roles.
func filterProfileByPrivacy(p UserProfile, requesterID string, requesterRoles []string) UserProfile {
	if p.ID == requesterID {
		return p
	}
	// Non-owner filtering
	switch p.Privacy.ProfileVisibility {
	case "PRIVATE":
		// Return only minimal public info
		p.Email = ""
		p.Location = nil
		p.Interests = nil
		p.Skills = nil
		p.Bio = nil
	case "VOLUNTEERS_ONLY":
		// Limited fields; hide email unless permitted
		if !p.Privacy.ShowEmail {
			p.Email = ""
		}
		if !p.Privacy.ShowLocation {
			p.Location = nil
		}
	default: // PUBLIC
		if !p.Privacy.ShowEmail {
			p.Email = ""
		}
		if !p.Privacy.ShowLocation {
			p.Location = nil
		}
	}
	return p
}
