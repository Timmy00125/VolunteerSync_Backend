package event

import (
	"time"
)

// EventStatus represents the current state of an event
type EventStatus string

const (
	EventStatusDraft     EventStatus = "DRAFT"
	EventStatusPublished EventStatus = "PUBLISHED"
	EventStatusCancelled EventStatus = "CANCELLED"
	EventStatusCompleted EventStatus = "COMPLETED"
	EventStatusArchived  EventStatus = "ARCHIVED"
)

// EventCategory represents the type/category of volunteer event
type EventCategory string

const (
	EventCategoryEnvironment      EventCategory = "ENVIRONMENT"
	EventCategoryEducation        EventCategory = "EDUCATION"
	EventCategoryHealth           EventCategory = "HEALTH"
	EventCategoryCommunityService EventCategory = "COMMUNITY_SERVICE"
	EventCategoryDisasterRelief   EventCategory = "DISASTER_RELIEF"
	EventCategoryAnimalWelfare    EventCategory = "ANIMAL_WELFARE"
	EventCategoryArtsCulture      EventCategory = "ARTS_CULTURE"
	EventCategoryTechnology       EventCategory = "TECHNOLOGY"
	EventCategorySportsRecreation EventCategory = "SPORTS_RECREATION"
	EventCategorySeniorCare       EventCategory = "SENIOR_CARE"
	EventCategoryYouthMentoring   EventCategory = "YOUTH_MENTORING"
	EventCategoryFoodSecurity     EventCategory = "FOOD_SECURITY"
)

// TimeCommitmentType represents the duration commitment for an event
type TimeCommitmentType string

const (
	TimeCommitmentOneTime    TimeCommitmentType = "ONE_TIME"
	TimeCommitmentShortTerm  TimeCommitmentType = "SHORT_TERM"  // < 1 month
	TimeCommitmentMediumTerm TimeCommitmentType = "MEDIUM_TERM" // 1-6 months
	TimeCommitmentLongTerm   TimeCommitmentType = "LONG_TERM"   // > 6 months
	TimeCommitmentOngoing    TimeCommitmentType = "ONGOING"
)

// SkillProficiency represents the required skill level
type SkillProficiency string

const (
	SkillProficiencyBeginner     SkillProficiency = "BEGINNER"
	SkillProficiencyIntermediate SkillProficiency = "INTERMEDIATE"
	SkillProficiencyAdvanced     SkillProficiency = "ADVANCED"
	SkillProficiencyExpert       SkillProficiency = "EXPERT"
)

// RecurrenceFrequency represents how often an event recurs
type RecurrenceFrequency string

const (
	RecurrenceFrequencyDaily   RecurrenceFrequency = "DAILY"
	RecurrenceFrequencyWeekly  RecurrenceFrequency = "WEEKLY"
	RecurrenceFrequencyMonthly RecurrenceFrequency = "MONTHLY"
	RecurrenceFrequencyYearly  RecurrenceFrequency = "YEARLY"
)

// DayOfWeek represents days of the week for recurrence
type DayOfWeek string

const (
	DayOfWeekMonday    DayOfWeek = "MONDAY"
	DayOfWeekTuesday   DayOfWeek = "TUESDAY"
	DayOfWeekWednesday DayOfWeek = "WEDNESDAY"
	DayOfWeekThursday  DayOfWeek = "THURSDAY"
	DayOfWeekFriday    DayOfWeek = "FRIDAY"
	DayOfWeekSaturday  DayOfWeek = "SATURDAY"
	DayOfWeekSunday    DayOfWeek = "SUNDAY"
)

// UpdateType represents the type of update made to an event
type UpdateType string

const (
	UpdateTypeMinor        UpdateType = "MINOR"
	UpdateTypeMajor        UpdateType = "MAJOR"
	UpdateTypeStatusChange UpdateType = "STATUS_CHANGE"
)

// Event represents a volunteer event
type Event struct {
	ID                   string               `json:"id" db:"id"`
	Title                string               `json:"title" db:"title"`
	Description          string               `json:"description" db:"description"`
	ShortDescription     *string              `json:"shortDescription" db:"short_description"`
	OrganizerID          string               `json:"organizerId" db:"organizer_id"`
	Status               EventStatus          `json:"status" db:"status"`
	StartTime            time.Time            `json:"startTime" db:"start_time"`
	EndTime              time.Time            `json:"endTime" db:"end_time"`
	Location             EventLocation        `json:"location"`
	Capacity             EventCapacity        `json:"capacity"`
	Requirements         EventRequirements    `json:"requirements"`
	Images               []EventImage         `json:"images"`
	Tags                 []string             `json:"tags" db:"tags"`
	Category             EventCategory        `json:"category" db:"category"`
	TimeCommitment       TimeCommitmentType   `json:"timeCommitment" db:"time_commitment"`
	RecurrenceRule       *RecurrenceRule      `json:"recurrenceRule,omitempty"`
	ParentEventID        *string              `json:"parentEventId,omitempty" db:"parent_event_id"`
	RegistrationSettings RegistrationSettings `json:"registrationSettings"`
	Slug                 *string              `json:"slug,omitempty" db:"slug"`
	ShareURL             *string              `json:"shareUrl,omitempty" db:"share_url"`
	CreatedAt            time.Time            `json:"createdAt" db:"created_at"`
	UpdatedAt            time.Time            `json:"updatedAt" db:"updated_at"`
	PublishedAt          *time.Time           `json:"publishedAt,omitempty" db:"published_at"`
}

// EventLocation represents the location information for an event
type EventLocation struct {
	Name         string       `json:"name" db:"location_name"`
	Address      string       `json:"address" db:"location_address"`
	City         string       `json:"city" db:"location_city"`
	State        *string      `json:"state,omitempty" db:"location_state"`
	Country      string       `json:"country" db:"location_country"`
	ZipCode      *string      `json:"zipCode,omitempty" db:"location_zip_code"`
	Coordinates  *Coordinates `json:"coordinates,omitempty"`
	Instructions *string      `json:"instructions,omitempty" db:"location_instructions"`
	IsRemote     bool         `json:"isRemote" db:"is_remote"`
}

// Coordinates represents geographic coordinates
type Coordinates struct {
	Latitude  float64 `json:"latitude" db:"location_latitude"`
	Longitude float64 `json:"longitude" db:"location_longitude"`
}

// EventCapacity represents capacity and registration limits
type EventCapacity struct {
	Minimum         int  `json:"minimum" db:"min_capacity"`
	Maximum         int  `json:"maximum" db:"max_capacity"`
	Current         int  `json:"current"`
	WaitlistEnabled bool `json:"waitlistEnabled" db:"waitlist_enabled"`
	WaitlistSize    int  `json:"waitlistSize"`
}

// EventRequirements represents volunteer requirements for an event
type EventRequirements struct {
	MinimumAge           *int                  `json:"minimumAge,omitempty" db:"minimum_age"`
	Skills               []SkillRequirement    `json:"skills"`
	Interests            []string              `json:"interests"` // Interest IDs
	BackgroundCheck      bool                  `json:"backgroundCheck" db:"background_check_required"`
	Training             []TrainingRequirement `json:"training"`
	Equipment            []string              `json:"equipment"`
	PhysicalRequirements *string               `json:"physicalRequirements,omitempty" db:"physical_requirements"`
}

// SkillRequirement represents a required skill for an event
type SkillRequirement struct {
	ID          string           `json:"id,omitempty" db:"id"`
	EventID     string           `json:"eventId,omitempty" db:"event_id"`
	Skill       string           `json:"skill" db:"skill_name"`
	Proficiency SkillProficiency `json:"proficiency" db:"proficiency"`
	Required    bool             `json:"required" db:"required"`
	CreatedAt   time.Time        `json:"createdAt,omitempty" db:"created_at"`
}

// TrainingRequirement represents training needed for an event
type TrainingRequirement struct {
	ID                  string    `json:"id,omitempty" db:"id"`
	EventID             string    `json:"eventId,omitempty" db:"event_id"`
	Name                string    `json:"name" db:"name"`
	Description         *string   `json:"description,omitempty" db:"description"`
	Required            bool      `json:"required" db:"required"`
	ProvidedByOrganizer bool      `json:"providedByOrganizer" db:"provided_by_organizer"`
	CreatedAt           time.Time `json:"createdAt,omitempty" db:"created_at"`
}

// EventImage represents an image associated with an event
type EventImage struct {
	ID           string    `json:"id" db:"id"`
	EventID      string    `json:"eventId" db:"event_id"`
	FileID       string    `json:"fileId" db:"file_id"`
	URL          string    `json:"url"`
	AltText      *string   `json:"altText,omitempty" db:"alt_text"`
	IsPrimary    bool      `json:"isPrimary" db:"is_primary"`
	DisplayOrder int       `json:"displayOrder" db:"display_order"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

// EventAnnouncement represents an announcement for an event
type EventAnnouncement struct {
	ID        string    `json:"id" db:"id"`
	EventID   string    `json:"eventId" db:"event_id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	IsUrgent  bool      `json:"isUrgent" db:"is_urgent"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// RecurrenceRule represents how an event recurs
type RecurrenceRule struct {
	Frequency       RecurrenceFrequency `json:"frequency"`
	Interval        int                 `json:"interval"`
	DaysOfWeek      []DayOfWeek         `json:"daysOfWeek,omitempty"`
	DayOfMonth      *int                `json:"dayOfMonth,omitempty"`
	EndDate         *time.Time          `json:"endDate,omitempty"`
	OccurrenceCount *int                `json:"occurrenceCount,omitempty"`
}

// RegistrationSettings represents event registration configuration
type RegistrationSettings struct {
	OpensAt              *time.Time `json:"opensAt,omitempty" db:"registration_opens_at"`
	ClosesAt             time.Time  `json:"closesAt" db:"registration_closes_at"`
	RequiresApproval     bool       `json:"requiresApproval" db:"requires_approval"`
	AllowWaitlist        bool       `json:"allowWaitlist" db:"waitlist_enabled"`
	ConfirmationRequired bool       `json:"confirmationRequired" db:"confirmation_required"`
	CancellationDeadline *time.Time `json:"cancellationDeadline,omitempty" db:"cancellation_deadline"`
}

// EventUpdate represents a change made to an event (audit log)
type EventUpdate struct {
	ID         string     `json:"id" db:"id"`
	EventID    string     `json:"eventId" db:"event_id"`
	UpdatedBy  string     `json:"updatedBy" db:"updated_by"`
	FieldName  string     `json:"fieldName" db:"field_name"`
	OldValue   *string    `json:"oldValue,omitempty" db:"old_value"`
	NewValue   *string    `json:"newValue,omitempty" db:"new_value"`
	UpdateType UpdateType `json:"updateType" db:"update_type"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
}

// CreateEventInput represents input for creating a new event
type CreateEventInput struct {
	Title                string                    `json:"title" validate:"required,min=3,max=200"`
	Description          string                    `json:"description" validate:"required,min=10,max=5000"`
	ShortDescription     *string                   `json:"shortDescription,omitempty" validate:"omitempty,max=300"`
	StartTime            time.Time                 `json:"startTime" validate:"required"`
	EndTime              time.Time                 `json:"endTime" validate:"required"`
	Location             EventLocationInput        `json:"location" validate:"required"`
	Capacity             EventCapacityInput        `json:"capacity" validate:"required"`
	Requirements         *EventRequirementsInput   `json:"requirements,omitempty"`
	Tags                 []string                  `json:"tags,omitempty" validate:"max=10,dive,max=50"`
	Category             EventCategory             `json:"category" validate:"required"`
	TimeCommitment       TimeCommitmentType        `json:"timeCommitment" validate:"required"`
	RecurrenceRule       *RecurrenceRuleInput      `json:"recurrenceRule,omitempty"`
	RegistrationSettings RegistrationSettingsInput `json:"registrationSettings" validate:"required"`
}

// UpdateEventInput represents input for updating an existing event
type UpdateEventInput struct {
	Title            *string                 `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
	Description      *string                 `json:"description,omitempty" validate:"omitempty,min=10,max=5000"`
	ShortDescription *string                 `json:"shortDescription,omitempty" validate:"omitempty,max=300"`
	Location         *EventLocationInput     `json:"location,omitempty"`
	Requirements     *EventRequirementsInput `json:"requirements,omitempty"`
	Tags             []string                `json:"tags,omitempty" validate:"max=10,dive,max=50"`
	Category         *EventCategory          `json:"category,omitempty"`
}

// EventLocationInput represents input for event location
type EventLocationInput struct {
	Name         string            `json:"name" validate:"required,min=1,max=200"`
	Address      string            `json:"address" validate:"required,min=5,max=500"`
	City         string            `json:"city" validate:"required,min=1,max=100"`
	State        *string           `json:"state,omitempty" validate:"omitempty,max=100"`
	Country      string            `json:"country" validate:"required,min=2,max=100"`
	ZipCode      *string           `json:"zipCode,omitempty" validate:"omitempty,max=20"`
	Coordinates  *CoordinatesInput `json:"coordinates,omitempty"`
	Instructions *string           `json:"instructions,omitempty" validate:"omitempty,max=1000"`
	IsRemote     bool              `json:"isRemote"`
}

// CoordinatesInput represents input for geographic coordinates
type CoordinatesInput struct {
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

// EventCapacityInput represents input for event capacity
type EventCapacityInput struct {
	Minimum         int  `json:"minimum" validate:"required,min=1"`
	Maximum         int  `json:"maximum" validate:"required,min=1"`
	WaitlistEnabled bool `json:"waitlistEnabled"`
}

// EventRequirementsInput represents input for event requirements
type EventRequirementsInput struct {
	MinimumAge           *int                       `json:"minimumAge,omitempty" validate:"omitempty,min=0,max=120"`
	Skills               []SkillRequirementInput    `json:"skills,omitempty" validate:"max=20"`
	Interests            []string                   `json:"interests,omitempty" validate:"max=20"`
	BackgroundCheck      bool                       `json:"backgroundCheck"`
	Training             []TrainingRequirementInput `json:"training,omitempty" validate:"max=10"`
	Equipment            []string                   `json:"equipment,omitempty" validate:"max=20,dive,max=100"`
	PhysicalRequirements *string                    `json:"physicalRequirements,omitempty" validate:"omitempty,max=1000"`
}

// SkillRequirementInput represents input for skill requirements
type SkillRequirementInput struct {
	Skill       string           `json:"skill" validate:"required,min=1,max=100"`
	Proficiency SkillProficiency `json:"proficiency" validate:"required"`
	Required    bool             `json:"required"`
}

// TrainingRequirementInput represents input for training requirements
type TrainingRequirementInput struct {
	Name                string  `json:"name" validate:"required,min=1,max=200"`
	Description         *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Required            bool    `json:"required"`
	ProvidedByOrganizer bool    `json:"providedByOrganizer"`
}

// RecurrenceRuleInput represents input for recurrence rules
type RecurrenceRuleInput struct {
	Frequency       RecurrenceFrequency `json:"frequency" validate:"required"`
	Interval        int                 `json:"interval" validate:"required,min=1,max=365"`
	DaysOfWeek      []DayOfWeek         `json:"daysOfWeek,omitempty" validate:"max=7"`
	DayOfMonth      *int                `json:"dayOfMonth,omitempty" validate:"omitempty,min=1,max=31"`
	EndDate         *time.Time          `json:"endDate,omitempty"`
	OccurrenceCount *int                `json:"occurrenceCount,omitempty" validate:"omitempty,min=1,max=1000"`
}

// RegistrationSettingsInput represents input for registration settings
type RegistrationSettingsInput struct {
	OpensAt              *time.Time `json:"opensAt,omitempty"`
	ClosesAt             time.Time  `json:"closesAt" validate:"required"`
	RequiresApproval     bool       `json:"requiresApproval"`
	AllowWaitlist        bool       `json:"allowWaitlist"`
	ConfirmationRequired bool       `json:"confirmationRequired"`
	CancellationDeadline *time.Time `json:"cancellationDeadline,omitempty"`
}

// EventSearchFilter represents filters for event search
type EventSearchFilter struct {
	Query             *string              `json:"query,omitempty"`
	Status            []EventStatus        `json:"status,omitempty"`
	Location          *LocationSearchInput `json:"location,omitempty"`
	DateRange         *DateRangeInput      `json:"dateRange,omitempty"`
	Skills            []string             `json:"skills,omitempty"`
	Interests         []string             `json:"interests,omitempty"`
	Categories        []EventCategory      `json:"categories,omitempty"`
	TimeCommitment    []TimeCommitmentType `json:"timeCommitment,omitempty"`
	Tags              []string             `json:"tags,omitempty"`
	HasAvailableSpots *bool                `json:"hasAvailableSpots,omitempty"`
}

// LocationSearchInput represents location-based search parameters
type LocationSearchInput struct {
	Center CoordinatesInput `json:"center" validate:"required"`
	Radius float64          `json:"radius" validate:"required,min=0.1,max=500"` // in kilometers
}

// DateRangeInput represents a date range for filtering
type DateRangeInput struct {
	StartDate time.Time `json:"startDate" validate:"required"`
	EndDate   time.Time `json:"endDate" validate:"required"`
}

// EventSortInput represents sorting options for events
type EventSortInput struct {
	Field     EventSortField `json:"field" validate:"required"`
	Direction SortDirection  `json:"direction" validate:"required"`
}

// EventSortField represents fields that can be used for sorting
type EventSortField string

const (
	EventSortFieldStartTime         EventSortField = "START_TIME"
	EventSortFieldCreatedAt         EventSortField = "CREATED_AT"
	EventSortFieldPopularity        EventSortField = "POPULARITY"
	EventSortFieldDistance          EventSortField = "DISTANCE"
	EventSortFieldCapacityRemaining EventSortField = "CAPACITY_REMAINING"
)

// SortDirection represents sorting direction
type SortDirection string

const (
	SortDirectionASC  SortDirection = "ASC"
	SortDirectionDESC SortDirection = "DESC"
)

// EventConnection represents a paginated list of events
type EventConnection struct {
	Edges      []EventEdge `json:"edges"`
	PageInfo   PageInfo    `json:"pageInfo"`
	TotalCount int         `json:"totalCount"`
}

// EventEdge represents an edge in the event connection
type EventEdge struct {
	Node   Event  `json:"node"`
	Cursor string `json:"cursor"`
}

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor,omitempty"`
	EndCursor       *string `json:"endCursor,omitempty"`
}
