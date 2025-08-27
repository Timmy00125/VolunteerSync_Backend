package user

import "time"

// Location represents user's location data.
type Location struct {
	City    *string
	State   *string
	Country *string
	Lat     *float64
	Lng     *float64
}

// PrivacySettings controls visibility and messaging preferences.
type PrivacySettings struct {
	ProfileVisibility string // PUBLIC, VOLUNTEERS_ONLY, PRIVATE
	ShowEmail         bool
	ShowLocation      bool
	AllowMessaging    bool
}

// NotificationPreferences stores notification toggles.
type NotificationPreferences struct {
	EmailNotifications     bool
	PushNotifications      bool
	SMSNotifications       bool
	EventReminders         bool
	NewOpportunities       bool
	NewsletterSubscription bool
}

// Interest represents an interest with category.
type Interest struct {
	ID       string
	Name     string
	Category string
}

// Skill represents a skill with proficiency.
type Skill struct {
	ID          string
	Name        string
	Proficiency string // BEGINNER|INTERMEDIATE|ADVANCED|EXPERT
	Verified    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserProfile aggregates user data for profile management.
type UserProfile struct {
	ID                string
	Name              string
	Email             string
	Bio               *string
	Location          *Location
	ProfilePictureURL *string
	Interests         []Interest
	Skills            []Skill
	Privacy           PrivacySettings
	Notifications     NotificationPreferences
	Roles             []string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	LastActiveAt      *time.Time
	IsVerified        bool
}

// ActivityLog represents a user activity record.
type ActivityLog struct {
	ID        string
	UserID    string
	Action    string
	Details   map[string]any
	IPAddress *string
	UserAgent *string
	CreatedAt time.Time
}

// UpdateProfileInput represents editable fields of a profile.
type UpdateProfileInput struct {
	Name     *string
	Bio      *string
	Location *Location
}

// SkillInput represents input to add a skill.
type SkillInput struct {
	Name        string
	Proficiency string
}
