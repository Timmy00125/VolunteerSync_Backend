package registration

import "context"

// RegistrationStore defines the interface for interacting with the registration data layer.

type Repository interface {
	// Registration methods
	CreateRegistration(ctx context.Context, arg *Registration) (*Registration, error)
	GetRegistrationByID(ctx context.Context, id string) (*Registration, error)
	GetRegistrationsByEventID(ctx context.Context, eventID string) ([]*Registration, error)
	GetRegistrationsByUserID(ctx context.Context, userID string) ([]*Registration, error)
	UpdateRegistration(ctx context.Context, arg *Registration) error
	DeleteRegistration(ctx context.Context, id string) error

	// Waitlist methods
	AddWaitlistEntry(ctx context.Context, arg *WaitlistEntry) (*WaitlistEntry, error)
	GetWaitlistEntryByRegistrationID(ctx context.Context, registrationID string) (*WaitlistEntry, error)
	GetWaitlistEntriesByEventID(ctx context.Context, eventID string) ([]*WaitlistEntry, error)
	UpdateWaitlistEntry(ctx context.Context, arg *WaitlistEntry) error
	RemoveWaitlistEntry(ctx context.Context, id string) error

	// Conflict methods
	CreateRegistrationConflict(ctx context.Context, arg *RegistrationConflict) (*RegistrationConflict, error)
	GetRegistrationConflictsByUserID(ctx context.Context, userID string) ([]*RegistrationConflict, error)
	UpdateRegistrationConflict(ctx context.Context, arg *RegistrationConflict) error

	// Attendance methods
	CreateAttendanceRecord(ctx context.Context, arg *AttendanceRecord) (*AttendanceRecord, error)
	GetAttendanceRecordsByRegistrationID(ctx context.Context, registrationID string) ([]*AttendanceRecord, error)
	UpdateAttendanceRecord(ctx context.Context, arg *AttendanceRecord) error
}
