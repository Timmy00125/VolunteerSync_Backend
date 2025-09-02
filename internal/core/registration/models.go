package registration

import (
	"time"
)

type RegistrationStatus string

const (
	StatusPendingApproval RegistrationStatus = "PENDING_APPROVAL"
	StatusConfirmed       RegistrationStatus = "CONFIRMED"
	StatusWaitlisted      RegistrationStatus = "WAITLISTED"
	StatusCancelled       RegistrationStatus = "CANCELLED"
	StatusDeclined        RegistrationStatus = "DECLINED"
	StatusNoShow          RegistrationStatus = "NO_SHOW"
	StatusCompleted       RegistrationStatus = "COMPLETED"
)

type AttendanceStatus string

const (
	AttendanceRegistered AttendanceStatus = "REGISTERED"
	AttendanceCheckedIn  AttendanceStatus = "CHECKED_IN"
	AttendanceCompleted  AttendanceStatus = "COMPLETED"
	AttendanceNoShow     AttendanceStatus = "NO_SHOW"
	AttendanceCancelled  AttendanceStatus = "CANCELLED"
)

type ConflictType string

const (
	ConflictTimeOverlap         ConflictType = "TIME_OVERLAP"
	ConflictLocation            ConflictType = "LOCATION_CONFLICT"
	ConflictTravelTime          ConflictType = "TRAVEL_TIME_CONFLICT"
	ConflictSkillOvercommitment ConflictType = "SKILL_OVERCOMMITMENT"
)

type ConflictSeverity string

const (
	SeverityLow      ConflictSeverity = "LOW"
	SeverityMedium   ConflictSeverity = "MEDIUM"
	SeverityHigh     ConflictSeverity = "HIGH"
	SeverityCritical ConflictSeverity = "CRITICAL"
)

type Registration struct {
	ID                    string             `json:"id"`
	UserID                string             `json:"userId"`
	EventID               string             `json:"eventId"`
	Status                RegistrationStatus `json:"status"`
	PersonalMessage       string             `json:"personalMessage"`
	ApprovalNotes         string             `json:"approvalNotes"`
	CancellationReason    string             `json:"cancellationReason"`
	AttendanceStatus      AttendanceStatus   `json:"attendanceStatus"`
	AppliedAt             time.Time          `json:"appliedAt"`
	ConfirmedAt           *time.Time         `json:"confirmedAt,omitempty"`
	CancelledAt           *time.Time         `json:"cancelledAt,omitempty"`
	CheckedInAt           *time.Time         `json:"checkedInAt,omitempty"`
	CompletedAt           *time.Time         `json:"completedAt,omitempty"`
	WaitlistPosition      *int               `json:"waitlistPosition,omitempty"`
	WaitlistPromotedAt    *time.Time         `json:"waitlistPromotedAt,omitempty"`
	PromotionOfferedAt    *time.Time         `json:"promotionOfferedAt,omitempty"`
	PromotionExpiresAt    *time.Time         `json:"promotionExpiresAt,omitempty"`
	AutoPromote           bool               `json:"autoPromote"`
	EmergencyContactName  string             `json:"emergencyContactName"`
	EmergencyContactPhone string             `json:"emergencyContactPhone"`
	DietaryRestrictions   string             `json:"dietaryRestrictions"`
	AccessibilityNeeds    string             `json:"accessibilityNeeds"`
	CheckedInBy           *string            `json:"checkedInBy,omitempty"`
	ApprovedBy            *string            `json:"approvedBy,omitempty"`
	CreatedAt             time.Time          `json:"createdAt"`
	UpdatedAt             time.Time          `json:"updatedAt"`
}

type RegistrationSkill struct {
	RegistrationID string    `json:"registrationId"`
	SkillName      string    `json:"skillName"`
	Proficiency    string    `json:"proficiency"`
	CreatedAt      time.Time `json:"createdAt"`
}

type RegistrationInterest struct {
	RegistrationID string    `json:"registrationId"`
	InterestID     string    `json:"interestId"`
	CreatedAt      time.Time `json:"createdAt"`
}

type AttendanceRecord struct {
	ID               string     `json:"id"`
	RegistrationID   string     `json:"registrationId"`
	Status           string     `json:"status"`
	CheckedInAt      *time.Time `json:"checkedInAt,omitempty"`
	CheckedOutAt     *time.Time `json:"checkedOutAt,omitempty"`
	CheckedInBy      *string    `json:"checkedInBy,omitempty"`
	LocationVerified bool       `json:"locationVerified"`
	Notes            string     `json:"notes"`
	CreatedAt        time.Time  `json:"createdAt"`
}

type RegistrationStatusChange struct {
	ID             string    `json:"id"`
	RegistrationID string    `json:"registrationId"`
	OldStatus      *string   `json:"oldStatus,omitempty"`
	NewStatus      string    `json:"newStatus"`
	ChangedBy      *string   `json:"changedBy,omitempty"`
	Reason         string    `json:"reason"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"createdAt"`
}

type WaitlistEntry struct {
	ID                 string     `json:"id"`
	RegistrationID     string     `json:"registrationId"`
	Position           int        `json:"position"`
	PriorityScore      float64    `json:"priorityScore"`
	AutoPromote        bool       `json:"autoPromote"`
	PromotionOfferedAt *time.Time `json:"promotionOfferedAt,omitempty"`
	PromotionExpiresAt *time.Time `json:"promotionExpiresAt,omitempty"`
	DeclinedPromotion  bool       `json:"declinedPromotion"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type RegistrationConflict struct {
	ID                 string           `json:"id"`
	UserID             string           `json:"userId"`
	PrimaryEventID     string           `json:"primaryEventId"`
	ConflictingEventID string           `json:"conflictingEventId"`
	ConflictType       ConflictType     `json:"conflictType"`
	Severity           ConflictSeverity `json:"severity"`
	Resolved           bool             `json:"resolved"`
	ResolutionNotes    string           `json:"resolutionNotes"`
	CreatedAt          time.Time        `json:"createdAt"`
}
