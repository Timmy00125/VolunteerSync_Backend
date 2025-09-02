package postgres

import (
	"context"
	"database/sql"

	"github.com/volunteersync/backend/internal/core/registration"
)

// RegistrationStorePG implements the registration.Repository interface using PostgreSQL

type RegistrationStorePG struct {
	db *sql.DB
}

// NewRegistrationStore creates a new PostgreSQL registration store

func NewRegistrationStore(db *sql.DB) *RegistrationStorePG {
	return &RegistrationStorePG{db: db}
}

func (s *RegistrationStorePG) UpdateAttendanceRecord(ctx context.Context, a *registration.AttendanceRecord) error {
	query := `
		UPDATE attendance_records
		SET
			status = $2, checked_in_at = $3, checked_out_at = $4, checked_in_by = $5, location_verified = $6, notes = $7
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query, a.ID, a.Status, a.CheckedInAt, a.CheckedOutAt, a.CheckedInBy, a.LocationVerified, a.Notes)

	return err
}

func (s *RegistrationStorePG) GetAttendanceRecordsByRegistrationID(ctx context.Context, registrationID string) ([]*registration.AttendanceRecord, error) {
	query := `
		SELECT
			id, registration_id, status, checked_in_at, checked_out_at, checked_in_by, location_verified, notes, created_at
		FROM attendance_records
		WHERE registration_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, registrationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*registration.AttendanceRecord
	for rows.Next() {
		a := &registration.AttendanceRecord{}
		if err := rows.Scan(
			&a.ID, &a.RegistrationID, &a.Status, &a.CheckedInAt, &a.CheckedOutAt, &a.CheckedInBy, &a.LocationVerified, &a.Notes, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, a)
	}

	return records, nil
}

func (s *RegistrationStorePG) CreateAttendanceRecord(ctx context.Context, a *registration.AttendanceRecord) (*registration.AttendanceRecord, error) {
	query := `
		INSERT INTO attendance_records (
			id, registration_id, status, checked_in_at, checked_out_at, checked_in_by, location_verified, notes, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, NOW()
		) RETURNING id, created_at
	`

	err := s.db.QueryRowContext(ctx, query,
		a.ID, a.RegistrationID, a.Status, a.CheckedInAt, a.CheckedOutAt, a.CheckedInBy, a.LocationVerified, a.Notes,
	).Scan(&a.ID, &a.CreatedAt)

	if err != nil {
		return nil, err
	}

	return a, nil
}

func (s *RegistrationStorePG) UpdateRegistrationConflict(ctx context.Context, c *registration.RegistrationConflict) error {
	query := `
		UPDATE registration_conflicts
		SET
			resolved = $2, resolution_notes = $3
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query, c.ID, c.Resolved, c.ResolutionNotes)

	return err
}

func (s *RegistrationStorePG) GetRegistrationConflictsByUserID(ctx context.Context, userID string) ([]*registration.RegistrationConflict, error) {
	query := `
		SELECT
			id, user_id, primary_event_id, conflicting_event_id, conflict_type, severity, resolved, resolution_notes, created_at
		FROM registration_conflicts
		WHERE user_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conflicts []*registration.RegistrationConflict
	for rows.Next() {
		c := &registration.RegistrationConflict{}
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.PrimaryEventID, &c.ConflictingEventID, &c.ConflictType, &c.Severity, &c.Resolved, &c.ResolutionNotes, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		conflicts = append(conflicts, c)
	}

	return conflicts, nil
}

func (s *RegistrationStorePG) CreateRegistrationConflict(ctx context.Context, c *registration.RegistrationConflict) (*registration.RegistrationConflict, error) {
	query := `
		INSERT INTO registration_conflicts (
			id, user_id, primary_event_id, conflicting_event_id, conflict_type, severity, resolved, resolution_notes, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, NOW()
		) RETURNING id, created_at
	`

	err := s.db.QueryRowContext(ctx, query,
		c.ID, c.UserID, c.PrimaryEventID, c.ConflictingEventID, c.ConflictType, c.Severity, c.Resolved, c.ResolutionNotes,
	).Scan(&c.ID, &c.CreatedAt)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (s *RegistrationStorePG) RemoveWaitlistEntry(ctx context.Context, id string) error {
	query := `DELETE FROM waitlist_entries WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *RegistrationStorePG) UpdateWaitlistEntry(ctx context.Context, w *registration.WaitlistEntry) error {
	query := `
		UPDATE waitlist_entries
		SET
			position = $2, priority_score = $3, auto_promote = $4, promotion_offered_at = $5, promotion_expires_at = $6, declined_promotion = $7, updated_at = NOW()
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query,
		w.ID, w.Position, w.PriorityScore, w.AutoPromote, w.PromotionOfferedAt, w.PromotionExpiresAt, w.DeclinedPromotion,
	)

	return err
}

func (s *RegistrationStorePG) GetWaitlistEntriesByEventID(ctx context.Context, eventID string) ([]*registration.WaitlistEntry, error) {
	query := `
		SELECT
			w.id, w.registration_id, w.position, w.priority_score, w.auto_promote, w.promotion_offered_at, w.promotion_expires_at, w.declined_promotion, w.created_at, w.updated_at
		FROM waitlist_entries w
		JOIN registrations r ON w.registration_id = r.id
		WHERE r.event_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var waitlistEntries []*registration.WaitlistEntry
	for rows.Next() {
		w := &registration.WaitlistEntry{}
		if err := rows.Scan(
			&w.ID, &w.RegistrationID, &w.Position, &w.PriorityScore, &w.AutoPromote, &w.PromotionOfferedAt, &w.PromotionExpiresAt, &w.DeclinedPromotion, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, err
		}
		waitlistEntries = append(waitlistEntries, w)
	}

	return waitlistEntries, nil
}

func (s *RegistrationStorePG) GetWaitlistEntryByRegistrationID(ctx context.Context, registrationID string) (*registration.WaitlistEntry, error) {
	query := `
		SELECT
			id, registration_id, position, priority_score, auto_promote, promotion_offered_at, promotion_expires_at, declined_promotion, created_at, updated_at
		FROM waitlist_entries
		WHERE registration_id = $1
	`

	w := &registration.WaitlistEntry{}

	err := s.db.QueryRowContext(ctx, query, registrationID).Scan(
		&w.ID, &w.RegistrationID, &w.Position, &w.PriorityScore, &w.AutoPromote, &w.PromotionOfferedAt, &w.PromotionExpiresAt, &w.DeclinedPromotion, &w.CreatedAt, &w.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or a specific not-found error
		}
		return nil, err
	}

	return w, nil
}

func (s *RegistrationStorePG) AddWaitlistEntry(ctx context.Context, w *registration.WaitlistEntry) (*registration.WaitlistEntry, error) {
	query := `
		INSERT INTO waitlist_entries (
			id, registration_id, position, priority_score, auto_promote, promotion_offered_at, promotion_expires_at, declined_promotion, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(ctx, query,
		w.ID, w.RegistrationID, w.Position, w.PriorityScore, w.AutoPromote, w.PromotionOfferedAt, w.PromotionExpiresAt, w.DeclinedPromotion,
	).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return w, nil
}

func (s *RegistrationStorePG) DeleteRegistration(ctx context.Context, id string) error {
	query := `DELETE FROM registrations WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

func (s *RegistrationStorePG) UpdateRegistration(ctx context.Context, r *registration.Registration) error {
	query := `
		UPDATE registrations
		SET
			status = $2, personal_message = $3, approval_notes = $4, cancellation_reason = $5, attendance_status = $6,
			confirmed_at = $7, cancelled_at = $8, checked_in_at = $9, completed_at = $10, waitlist_position = $11,
			waitlist_promoted_at = $12, promotion_offered_at = $13, promotion_expires_at = $14, auto_promote = $15,
			emergency_contact_name = $16, emergency_contact_phone = $17, dietary_restrictions = $18, accessibility_needs = $19,
			checked_in_by = $20, approved_by = $21, updated_at = NOW()
		WHERE id = $1
	`

	_, err := s.db.ExecContext(ctx, query,
		r.ID, r.Status, r.PersonalMessage, r.ApprovalNotes, r.CancellationReason, r.AttendanceStatus,
		r.ConfirmedAt, r.CancelledAt, r.CheckedInAt, r.CompletedAt, r.WaitlistPosition, r.WaitlistPromotedAt,
		r.PromotionOfferedAt, r.PromotionExpiresAt, r.AutoPromote, r.EmergencyContactName, r.EmergencyContactPhone,
		r.DietaryRestrictions, r.AccessibilityNeeds, r.CheckedInBy, r.ApprovedBy,
	)

	return err
}

func (s *RegistrationStorePG) GetRegistrationsByUserID(ctx context.Context, userID string) ([]*registration.Registration, error) {
	query := `
		SELECT
			id, user_id, event_id, status, personal_message, approval_notes, cancellation_reason, attendance_status,
			applied_at, confirmed_at, cancelled_at, checked_in_at, completed_at, waitlist_position, waitlist_promoted_at,
			promotion_offered_at, promotion_expires_at, auto_promote, emergency_contact_name, emergency_contact_phone,
			dietary_restrictions, accessibility_needs, checked_in_by, approved_by, created_at, updated_at
		FROM registrations
		WHERE user_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var registrations []*registration.Registration
	for rows.Next() {
		r := &registration.Registration{}
		if err := rows.Scan(
			&r.ID, &r.UserID, &r.EventID, &r.Status, &r.PersonalMessage, &r.ApprovalNotes, &r.CancellationReason, &r.AttendanceStatus,
			&r.AppliedAt, &r.ConfirmedAt, &r.CancelledAt, &r.CheckedInAt, &r.CompletedAt, &r.WaitlistPosition, &r.WaitlistPromotedAt,
			&r.PromotionOfferedAt, &r.PromotionExpiresAt, &r.AutoPromote, &r.EmergencyContactName, &r.EmergencyContactPhone,
			&r.DietaryRestrictions, &r.AccessibilityNeeds, &r.CheckedInBy, &r.ApprovedBy, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		registrations = append(registrations, r)
	}

	return registrations, nil
}

func (s *RegistrationStorePG) GetRegistrationsByEventID(ctx context.Context, eventID string) ([]*registration.Registration, error) {
	query := `
		SELECT
			id, user_id, event_id, status, personal_message, approval_notes, cancellation_reason, attendance_status,
			applied_at, confirmed_at, cancelled_at, checked_in_at, completed_at, waitlist_position, waitlist_promoted_at,
			promotion_offered_at, promotion_expires_at, auto_promote, emergency_contact_name, emergency_contact_phone,
			dietary_restrictions, accessibility_needs, checked_in_by, approved_by, created_at, updated_at
		FROM registrations
		WHERE event_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var registrations []*registration.Registration
	for rows.Next() {
		r := &registration.Registration{}
		if err := rows.Scan(
			&r.ID, &r.UserID, &r.EventID, &r.Status, &r.PersonalMessage, &r.ApprovalNotes, &r.CancellationReason, &r.AttendanceStatus,
			&r.AppliedAt, &r.ConfirmedAt, &r.CancelledAt, &r.CheckedInAt, &r.CompletedAt, &r.WaitlistPosition, &r.WaitlistPromotedAt,
			&r.PromotionOfferedAt, &r.PromotionExpiresAt, &r.AutoPromote, &r.EmergencyContactName, &r.EmergencyContactPhone,
			&r.DietaryRestrictions, &r.AccessibilityNeeds, &r.CheckedInBy, &r.ApprovedBy, &r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, err
		}
		registrations = append(registrations, r)
	}

	return registrations, nil
}

func (s *RegistrationStorePG) GetRegistrationByID(ctx context.Context, id string) (*registration.Registration, error) {
	query := `
		SELECT
			id, user_id, event_id, status, personal_message, approval_notes, cancellation_reason, attendance_status,
			applied_at, confirmed_at, cancelled_at, checked_in_at, completed_at, waitlist_position, waitlist_promoted_at,
			promotion_offered_at, promotion_expires_at, auto_promote, emergency_contact_name, emergency_contact_phone,
			dietary_restrictions, accessibility_needs, checked_in_by, approved_by, created_at, updated_at
		FROM registrations
		WHERE id = $1
	`

	r := &registration.Registration{}

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&r.ID, &r.UserID, &r.EventID, &r.Status, &r.PersonalMessage, &r.ApprovalNotes, &r.CancellationReason, &r.AttendanceStatus,
		&r.AppliedAt, &r.ConfirmedAt, &r.CancelledAt, &r.CheckedInAt, &r.CompletedAt, &r.WaitlistPosition, &r.WaitlistPromotedAt,
		&r.PromotionOfferedAt, &r.PromotionExpiresAt, &r.AutoPromote, &r.EmergencyContactName, &r.EmergencyContactPhone,
		&r.DietaryRestrictions, &r.AccessibilityNeeds, &r.CheckedInBy, &r.ApprovedBy, &r.CreatedAt, &r.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Or a specific not-found error
		}
		return nil, err
	}

	return r, nil
}

// CreateRegistration creates a new registration in the database

func (s *RegistrationStorePG) CreateRegistration(ctx context.Context, r *registration.Registration) (*registration.Registration, error) {
	query := `
		INSERT INTO registrations (
			id, user_id, event_id, status, personal_message, approval_notes, cancellation_reason, attendance_status,
			applied_at, confirmed_at, cancelled_at, checked_in_at, completed_at, waitlist_position, waitlist_promoted_at,
			promotion_offered_at, promotion_expires_at, auto_promote, emergency_contact_name, emergency_contact_phone,
			dietary_restrictions, accessibility_needs, checked_in_by, approved_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(ctx, query,
		r.ID, r.UserID, r.EventID, r.Status, r.PersonalMessage, r.ApprovalNotes, r.CancellationReason, r.AttendanceStatus,
		r.AppliedAt, r.ConfirmedAt, r.CancelledAt, r.CheckedInAt, r.CompletedAt, r.WaitlistPosition, r.WaitlistPromotedAt,
		r.PromotionOfferedAt, r.PromotionExpiresAt, r.AutoPromote, r.EmergencyContactName, r.EmergencyContactPhone,
		r.DietaryRestrictions, r.AccessibilityNeeds, r.CheckedInBy, r.ApprovedBy,
	).Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return r, nil
}
