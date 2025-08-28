package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/lib/pq"
	"github.com/volunteersync/backend/internal/core/event"
)

// EventStorePG implements the event.Repository interface using PostgreSQL
type EventStorePG struct {
	db *sql.DB
}

// NewEventStore creates a new PostgreSQL event store
func NewEventStore(db *sql.DB) *EventStorePG {
	return &EventStorePG{db: db}
}

// Create creates a new event in the database
func (s *EventStorePG) Create(ctx context.Context, e *event.Event) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Insert main event record
	err = s.insertMainEvent(ctx, tx, e)
	if err != nil {
		return err
	}

	// Insert requirements
	if err := s.insertSkillRequirements(ctx, tx, e.ID, e.Requirements.Skills); err != nil {
		return err
	}

	if err := s.insertTrainingRequirements(ctx, tx, e.ID, e.Requirements.Training); err != nil {
		return err
	}

	if err := s.insertInterestRequirements(ctx, tx, e.ID, e.Requirements.Interests); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *EventStorePG) insertMainEvent(ctx context.Context, tx *sql.Tx, e *event.Event) error {
	var recurrenceJSON []byte
	if e.RecurrenceRule != nil {
		var err error
		recurrenceJSON, err = json.Marshal(e.RecurrenceRule)
		if err != nil {
			return err
		}
	}

	var lat, lng *float64
	if e.Location.Coordinates != nil {
		lat = &e.Location.Coordinates.Latitude
		lng = &e.Location.Coordinates.Longitude
	}

	query := `
		INSERT INTO events (
			id, title, description, short_description, organizer_id, status,
			start_time, end_time, location_name, location_address, location_city,
			location_state, location_country, location_zip_code, location_latitude,
			location_longitude, location_instructions, is_remote, min_capacity,
			max_capacity, waitlist_enabled, minimum_age, background_check_required,
			physical_requirements, category, time_commitment, tags,
			registration_opens_at, registration_closes_at, requires_approval,
			confirmation_required, cancellation_deadline, parent_event_id,
			recurrence_rule, slug, share_url, created_at, updated_at, published_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
			$31, $32, $33, $34, $35, $36, $37, $38, $39
		)`

	_, err := tx.ExecContext(ctx, query,
		e.ID, e.Title, e.Description, e.ShortDescription, e.OrganizerID, e.Status,
		e.StartTime, e.EndTime, e.Location.Name, e.Location.Address, e.Location.City,
		e.Location.State, e.Location.Country, e.Location.ZipCode, lat, lng,
		e.Location.Instructions, e.Location.IsRemote, e.Capacity.Minimum,
		e.Capacity.Maximum, e.Capacity.WaitlistEnabled, e.Requirements.MinimumAge,
		e.Requirements.BackgroundCheck, e.Requirements.PhysicalRequirements,
		e.Category, e.TimeCommitment, pq.Array(e.Tags),
		e.RegistrationSettings.OpensAt, e.RegistrationSettings.ClosesAt,
		e.RegistrationSettings.RequiresApproval, e.RegistrationSettings.ConfirmationRequired,
		e.RegistrationSettings.CancellationDeadline, e.ParentEventID,
		recurrenceJSON, e.Slug, e.ShareURL, e.CreatedAt, e.UpdatedAt, e.PublishedAt,
	)

	return err
}

// GetByID retrieves an event by its ID
func (s *EventStorePG) GetByID(ctx context.Context, id string) (*event.Event, error) {
	e := &event.Event{}
	query := `
		SELECT 
			id, title, description, short_description, organizer_id, status,
			start_time, end_time, location_name, location_address, location_city,
			location_state, location_country, location_zip_code, location_latitude,
			location_longitude, location_instructions, is_remote, min_capacity,
			max_capacity, waitlist_enabled, minimum_age, background_check_required,
			physical_requirements, category, time_commitment, tags,
			registration_opens_at, registration_closes_at, requires_approval,
			confirmation_required, cancellation_deadline, parent_event_id,
			recurrence_rule, slug, share_url, created_at, updated_at, published_at
		FROM events 
		WHERE id = $1`

	err := s.scanEvent(s.db.QueryRowContext(ctx, query, id), e)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found: %s", id)
		}
		return nil, err
	}

	// Load related data
	if err := s.loadEventRelations(ctx, e); err != nil {
		return nil, err
	}

	return e, nil
}

// GetBySlug retrieves an event by its slug
func (s *EventStorePG) GetBySlug(ctx context.Context, slug string) (*event.Event, error) {
	var id string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM events WHERE slug = $1", slug).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found with slug: %s", slug)
		}
		return nil, err
	}

	return s.GetByID(ctx, id)
}

// Update updates an existing event
func (s *EventStorePG) Update(ctx context.Context, e *event.Event) error {
	var recurrenceJSON []byte
	if e.RecurrenceRule != nil {
		var err error
		recurrenceJSON, err = json.Marshal(e.RecurrenceRule)
		if err != nil {
			return err
		}
	}

	var lat, lng *float64
	if e.Location.Coordinates != nil {
		lat = &e.Location.Coordinates.Latitude
		lng = &e.Location.Coordinates.Longitude
	}

	query := `
		UPDATE events SET
			title = $2, description = $3, short_description = $4,
			location_name = $5, location_address = $6, location_city = $7,
			location_state = $8, location_country = $9, location_zip_code = $10,
			location_latitude = $11, location_longitude = $12,
			location_instructions = $13, is_remote = $14,
			min_capacity = $15, max_capacity = $16, waitlist_enabled = $17,
			minimum_age = $18, background_check_required = $19,
			physical_requirements = $20, category = $21, time_commitment = $22,
			tags = $23, registration_opens_at = $24, registration_closes_at = $25,
			requires_approval = $26, confirmation_required = $27,
			cancellation_deadline = $28, recurrence_rule = $29, updated_at = NOW()
		WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, e.ID, e.Title, e.Description, e.ShortDescription,
		e.Location.Name, e.Location.Address, e.Location.City, e.Location.State,
		e.Location.Country, e.Location.ZipCode, lat, lng, e.Location.Instructions,
		e.Location.IsRemote, e.Capacity.Minimum, e.Capacity.Maximum,
		e.Capacity.WaitlistEnabled, e.Requirements.MinimumAge,
		e.Requirements.BackgroundCheck, e.Requirements.PhysicalRequirements,
		e.Category, e.TimeCommitment, pq.Array(e.Tags),
		e.RegistrationSettings.OpensAt, e.RegistrationSettings.ClosesAt,
		e.RegistrationSettings.RequiresApproval, e.RegistrationSettings.ConfirmationRequired,
		e.RegistrationSettings.CancellationDeadline, recurrenceJSON)

	return err
}

// Delete soft deletes an event
func (s *EventStorePG) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE events SET status = 'ARCHIVED', updated_at = NOW() WHERE id = $1", id)
	return err
}

// List retrieves events with filtering, sorting, and pagination
func (s *EventStorePG) List(ctx context.Context, filter event.EventSearchFilter, sort *event.EventSortInput, limit, offset int) (*event.EventConnection, error) {
	// Build basic query
	baseQuery := "FROM events e WHERE e.status != 'ARCHIVED'"
	args := []interface{}{}
	argCount := 0

	// Add text search filter
	if filter.Query != nil && *filter.Query != "" {
		argCount++
		baseQuery += fmt.Sprintf(` AND to_tsvector('english', e.title || ' ' || e.description) @@ plainto_tsquery('english', $%d)`, argCount)
		args = append(args, *filter.Query)
	}

	// Add category filter
	if len(filter.Categories) > 0 {
		argCount++
		baseQuery += fmt.Sprintf(` AND e.category = ANY($%d)`, argCount)
		categories := make([]string, len(filter.Categories))
		for i, cat := range filter.Categories {
			categories[i] = string(cat)
		}
		args = append(args, pq.Array(categories))
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, err
	}

	// Build order clause
	orderBy := "ORDER BY e.start_time ASC"
	if sort != nil {
		direction := "ASC"
		if sort.Direction == event.SortDirectionDESC {
			direction = "DESC"
		}
		if sort.Field == event.EventSortFieldStartTime {
			orderBy = fmt.Sprintf("ORDER BY e.start_time %s", direction)
		}
	}

	// Get events
	selectQuery := fmt.Sprintf(`
		SELECT 
			e.id, e.title, e.description, e.short_description, e.organizer_id, e.status,
			e.start_time, e.end_time, e.location_name, e.location_address, e.location_city,
			e.location_state, e.location_country, e.location_zip_code, e.location_latitude,
			e.location_longitude, e.location_instructions, e.is_remote, e.min_capacity,
			e.max_capacity, e.waitlist_enabled, e.minimum_age, e.background_check_required,
			e.physical_requirements, e.category, e.time_commitment, e.tags,
			e.registration_opens_at, e.registration_closes_at, e.requires_approval,
			e.confirmation_required, e.cancellation_deadline, e.parent_event_id,
			e.recurrence_rule, e.slug, e.share_url, e.created_at, e.updated_at, e.published_at
		%s %s LIMIT %d OFFSET %d`, baseQuery, orderBy, limit, offset)

	rows, err := s.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []*event.Event{}
	for rows.Next() {
		e := &event.Event{}
		if err := s.scanEventFromRows(rows, e); err != nil {
			return nil, err
		}
		if err := s.loadEventRelations(ctx, e); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return s.buildConnection(events, totalCount, limit, offset), nil
}

// scanEvent scans event from a single row
func (s *EventStorePG) scanEvent(row *sql.Row, e *event.Event) error {
	var recurrenceJSON []byte
	var tags pq.StringArray
	var lat, lng sql.NullFloat64

	err := row.Scan(
		&e.ID, &e.Title, &e.Description, &e.ShortDescription, &e.OrganizerID, &e.Status,
		&e.StartTime, &e.EndTime, &e.Location.Name, &e.Location.Address, &e.Location.City,
		&e.Location.State, &e.Location.Country, &e.Location.ZipCode, &lat, &lng,
		&e.Location.Instructions, &e.Location.IsRemote, &e.Capacity.Minimum,
		&e.Capacity.Maximum, &e.Capacity.WaitlistEnabled, &e.Requirements.MinimumAge,
		&e.Requirements.BackgroundCheck, &e.Requirements.PhysicalRequirements,
		&e.Category, &e.TimeCommitment, &tags, &e.RegistrationSettings.OpensAt,
		&e.RegistrationSettings.ClosesAt, &e.RegistrationSettings.RequiresApproval,
		&e.RegistrationSettings.ConfirmationRequired, &e.RegistrationSettings.CancellationDeadline,
		&e.ParentEventID, &recurrenceJSON, &e.Slug, &e.ShareURL,
		&e.CreatedAt, &e.UpdatedAt, &e.PublishedAt,
	)
	if err != nil {
		return err
	}

	// Set coordinates if available
	if lat.Valid && lng.Valid {
		e.Location.Coordinates = &event.Coordinates{
			Latitude:  lat.Float64,
			Longitude: lng.Float64,
		}
	}

	e.Tags = []string(tags)

	// Parse recurrence rule
	if len(recurrenceJSON) > 0 {
		var rule event.RecurrenceRule
		if err := json.Unmarshal(recurrenceJSON, &rule); err == nil {
			e.RecurrenceRule = &rule
		}
	}

	return nil
}

// scanEventFromRows scans event from rows result
func (s *EventStorePG) scanEventFromRows(rows *sql.Rows, e *event.Event) error {
	var recurrenceJSON []byte
	var tags pq.StringArray
	var lat, lng sql.NullFloat64

	err := rows.Scan(
		&e.ID, &e.Title, &e.Description, &e.ShortDescription, &e.OrganizerID, &e.Status,
		&e.StartTime, &e.EndTime, &e.Location.Name, &e.Location.Address, &e.Location.City,
		&e.Location.State, &e.Location.Country, &e.Location.ZipCode, &lat, &lng,
		&e.Location.Instructions, &e.Location.IsRemote, &e.Capacity.Minimum,
		&e.Capacity.Maximum, &e.Capacity.WaitlistEnabled, &e.Requirements.MinimumAge,
		&e.Requirements.BackgroundCheck, &e.Requirements.PhysicalRequirements,
		&e.Category, &e.TimeCommitment, &tags, &e.RegistrationSettings.OpensAt,
		&e.RegistrationSettings.ClosesAt, &e.RegistrationSettings.RequiresApproval,
		&e.RegistrationSettings.ConfirmationRequired, &e.RegistrationSettings.CancellationDeadline,
		&e.ParentEventID, &recurrenceJSON, &e.Slug, &e.ShareURL,
		&e.CreatedAt, &e.UpdatedAt, &e.PublishedAt,
	)
	if err != nil {
		return err
	}

	// Set coordinates if available
	if lat.Valid && lng.Valid {
		e.Location.Coordinates = &event.Coordinates{
			Latitude:  lat.Float64,
			Longitude: lng.Float64,
		}
	}

	e.Tags = []string(tags)

	// Parse recurrence rule
	if len(recurrenceJSON) > 0 {
		var rule event.RecurrenceRule
		if err := json.Unmarshal(recurrenceJSON, &rule); err == nil {
			e.RecurrenceRule = &rule
		}
	}

	return nil
}

// loadEventRelations loads related data for an event
func (s *EventStorePG) loadEventRelations(ctx context.Context, e *event.Event) error {
	// Load skill requirements
	skills, err := s.GetSkillRequirements(ctx, e.ID)
	if err != nil {
		return err
	}
	e.Requirements.Skills = make([]event.SkillRequirement, len(skills))
	for i, skill := range skills {
		e.Requirements.Skills[i] = *skill
	}

	// Load training requirements
	training, err := s.GetTrainingRequirements(ctx, e.ID)
	if err != nil {
		return err
	}
	e.Requirements.Training = make([]event.TrainingRequirement, len(training))
	for i, req := range training {
		e.Requirements.Training[i] = *req
	}

	// Load interest requirements
	interests, err := s.GetInterestRequirements(ctx, e.ID)
	if err != nil {
		return err
	}
	e.Requirements.Interests = interests

	// Load images
	images, err := s.GetEventImages(ctx, e.ID)
	if err != nil {
		return err
	}
	e.Images = make([]event.EventImage, len(images))
	for i, img := range images {
		e.Images[i] = *img
	}

	// Get current capacity
	currentCapacity, err := s.GetCurrentCapacity(ctx, e.ID)
	if err != nil {
		return err
	}
	e.Capacity.Current = currentCapacity

	return nil
}

// buildConnection builds EventConnection response
func (s *EventStorePG) buildConnection(events []*event.Event, totalCount, limit, offset int) *event.EventConnection {
	edges := make([]event.EventEdge, len(events))
	for i, e := range events {
		edges[i] = event.EventEdge{
			Node:   *e,
			Cursor: e.ID,
		}
	}

	hasNextPage := offset+limit < totalCount
	hasPreviousPage := offset > 0

	var startCursor, endCursor *string
	if len(events) > 0 {
		start := events[0].ID
		end := events[len(events)-1].ID
		startCursor = &start
		endCursor = &end
	}

	return &event.EventConnection{
		Edges: edges,
		PageInfo: event.PageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: hasPreviousPage,
			StartCursor:     startCursor,
			EndCursor:       endCursor,
		},
		TotalCount: totalCount,
	}
}

// Additional interface methods - placeholder implementations
func (s *EventStorePG) GetByOrganizer(ctx context.Context, organizerID string) ([]*event.Event, error) {
	return []*event.Event{}, nil // TODO: Implement
}

func (s *EventStorePG) GetFeatured(ctx context.Context, limit int) ([]*event.Event, error) {
	return []*event.Event{}, nil // TODO: Implement
}

func (s *EventStorePG) GetNearby(ctx context.Context, lat, lng, radius float64, limit int) ([]*event.Event, error) {
	// Use Haversine formula for distance calculation in standard PostgreSQL
	// Distance in kilometers
	query := `
		SELECT 
			id, title, description, short_description, organizer_id, status,
			start_time, end_time, location_name, location_address, location_city,
			location_state, location_country, location_zip_code, location_latitude,
			location_longitude, location_instructions, is_remote, min_capacity,
			max_capacity, waitlist_enabled, minimum_age, background_check_required,
			physical_requirements, category, time_commitment, tags,
			registration_opens_at, registration_closes_at, requires_approval,
			confirmation_required, cancellation_deadline, parent_event_id,
			recurrence_rule, slug, share_url, created_at, updated_at, published_at,
			(6371 * acos(cos(radians($1)) * cos(radians(location_latitude)) * 
			 cos(radians(location_longitude) - radians($2)) + 
			 sin(radians($1)) * sin(radians(location_latitude)))) AS distance
		FROM events 
		WHERE location_latitude IS NOT NULL 
			AND location_longitude IS NOT NULL
			AND status = 'PUBLISHED'
			AND (6371 * acos(cos(radians($1)) * cos(radians(location_latitude)) * 
				 cos(radians(location_longitude) - radians($2)) + 
				 sin(radians($1)) * sin(radians(location_latitude)))) <= $3
		ORDER BY distance
		LIMIT $4`

	rows, err := s.db.QueryContext(ctx, query, lat, lng, radius, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query nearby events: %w", err)
	}
	defer rows.Close()

	var events []*event.Event
	for rows.Next() {
		e := &event.Event{}
		var latNull, lngNull sql.NullFloat64
		var recurrenceJSON []byte
		var tags pq.StringArray
		var distance float64

		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.ShortDescription, &e.OrganizerID, &e.Status,
			&e.StartTime, &e.EndTime, &e.Location.Name, &e.Location.Address, &e.Location.City,
			&e.Location.State, &e.Location.Country, &e.Location.ZipCode, &latNull, &lngNull,
			&e.Location.Instructions, &e.Location.IsRemote, &e.Capacity.Minimum,
			&e.Capacity.Maximum, &e.Capacity.WaitlistEnabled, &e.Requirements.MinimumAge,
			&e.Requirements.BackgroundCheck, &e.Requirements.PhysicalRequirements,
			&e.Category, &e.TimeCommitment, &tags,
			&e.RegistrationSettings.OpensAt, &e.RegistrationSettings.ClosesAt,
			&e.RegistrationSettings.RequiresApproval, &e.RegistrationSettings.ConfirmationRequired,
			&e.RegistrationSettings.CancellationDeadline, &e.ParentEventID,
			&recurrenceJSON, &e.Slug, &e.ShareURL, &e.CreatedAt, &e.UpdatedAt, &e.PublishedAt,
			&distance,
		); err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Set coordinates if available
		if latNull.Valid && lngNull.Valid {
			e.Location.Coordinates = &event.Coordinates{
				Latitude:  latNull.Float64,
				Longitude: lngNull.Float64,
			}
		} else {
			e.Location.Coordinates = nil
		}

		e.Tags = []string(tags)

		// Parse recurrence rule
		if len(recurrenceJSON) > 0 {
			var rule event.RecurrenceRule
			if err := json.Unmarshal(recurrenceJSON, &rule); err == nil {
				e.RecurrenceRule = &rule
			}
		}

		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (s *EventStorePG) UpdateStatus(ctx context.Context, eventID string, status event.EventStatus) error {
	_, err := s.db.ExecContext(ctx, "UPDATE events SET status = $1, updated_at = NOW() WHERE id = $2", status, eventID)
	return err
}

func (s *EventStorePG) GetByStatus(ctx context.Context, status event.EventStatus, limit, offset int) ([]*event.Event, error) {
	return []*event.Event{}, nil // TODO: Implement
}

// Skill requirement methods
func (s *EventStorePG) insertSkillRequirements(ctx context.Context, tx *sql.Tx, eventID string, skills []event.SkillRequirement) error {
	for _, skill := range skills {
		query := `
			INSERT INTO event_skill_requirements (id, event_id, skill_name, proficiency, required, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())`
		_, err := tx.ExecContext(ctx, query, eventID, skill.Skill, skill.Proficiency, skill.Required)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *EventStorePG) CreateSkillRequirement(ctx context.Context, req *event.SkillRequirement) error {
	query := `
		INSERT INTO event_skill_requirements (id, event_id, skill_name, proficiency, required, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query, req.EventID, req.Skill, req.Proficiency, req.Required).
		Scan(&req.ID, &req.CreatedAt)
}

func (s *EventStorePG) GetSkillRequirements(ctx context.Context, eventID string) ([]*event.SkillRequirement, error) {
	query := `
		SELECT id, event_id, skill_name, proficiency, required, created_at
		FROM event_skill_requirements
		WHERE event_id = $1
		ORDER BY created_at`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requirements []*event.SkillRequirement
	for rows.Next() {
		req := &event.SkillRequirement{}
		err := rows.Scan(&req.ID, &req.EventID, &req.Skill, &req.Proficiency, &req.Required, &req.CreatedAt)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, req)
	}
	return requirements, nil
}

func (s *EventStorePG) UpdateSkillRequirements(ctx context.Context, eventID string, requirements []*event.SkillRequirement) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Delete existing
	_, err = tx.ExecContext(ctx, "DELETE FROM event_skill_requirements WHERE event_id = $1", eventID)
	if err != nil {
		return err
	}

	// Insert new
	for _, req := range requirements {
		query := `
			INSERT INTO event_skill_requirements (id, event_id, skill_name, proficiency, required, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())`
		_, err = tx.ExecContext(ctx, query, eventID, req.Skill, req.Proficiency, req.Required)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *EventStorePG) DeleteSkillRequirements(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_skill_requirements WHERE event_id = $1", eventID)
	return err
}

// Training requirement methods
func (s *EventStorePG) insertTrainingRequirements(ctx context.Context, tx *sql.Tx, eventID string, training []event.TrainingRequirement) error {
	for _, req := range training {
		query := `
			INSERT INTO event_training_requirements (id, event_id, name, description, required, provided_by_organizer, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())`
		_, err := tx.ExecContext(ctx, query, eventID, req.Name, req.Description, req.Required, req.ProvidedByOrganizer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *EventStorePG) CreateTrainingRequirement(ctx context.Context, req *event.TrainingRequirement) error {
	query := `
		INSERT INTO event_training_requirements (id, event_id, name, description, required, provided_by_organizer, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query, req.EventID, req.Name, req.Description, req.Required, req.ProvidedByOrganizer).
		Scan(&req.ID, &req.CreatedAt)
}

func (s *EventStorePG) GetTrainingRequirements(ctx context.Context, eventID string) ([]*event.TrainingRequirement, error) {
	query := `
		SELECT id, event_id, name, description, required, provided_by_organizer, created_at
		FROM event_training_requirements
		WHERE event_id = $1
		ORDER BY created_at`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requirements []*event.TrainingRequirement
	for rows.Next() {
		req := &event.TrainingRequirement{}
		err := rows.Scan(&req.ID, &req.EventID, &req.Name, &req.Description, &req.Required, &req.ProvidedByOrganizer, &req.CreatedAt)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, req)
	}
	return requirements, nil
}

func (s *EventStorePG) UpdateTrainingRequirements(ctx context.Context, eventID string, requirements []*event.TrainingRequirement) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Delete existing
	_, err = tx.ExecContext(ctx, "DELETE FROM event_training_requirements WHERE event_id = $1", eventID)
	if err != nil {
		return err
	}

	// Insert new
	for _, req := range requirements {
		query := `
			INSERT INTO event_training_requirements (id, event_id, name, description, required, provided_by_organizer, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())`
		_, err = tx.ExecContext(ctx, query, eventID, req.Name, req.Description, req.Required, req.ProvidedByOrganizer)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *EventStorePG) DeleteTrainingRequirements(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_training_requirements WHERE event_id = $1", eventID)
	return err
}

// Interest requirement methods
func (s *EventStorePG) insertInterestRequirements(ctx context.Context, tx *sql.Tx, eventID string, interestIDs []string) error {
	if len(interestIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO event_interest_requirements (event_id, interest_id, created_at)
		VALUES ($1, unnest($2::uuid[]), NOW())`
	_, err := tx.ExecContext(ctx, query, eventID, pq.Array(interestIDs))
	return err
}

func (s *EventStorePG) AddInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error {
	if len(interestIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO event_interest_requirements (event_id, interest_id, created_at)
		VALUES ($1, unnest($2::uuid[]), NOW())`
	_, err := s.db.ExecContext(ctx, query, eventID, pq.Array(interestIDs))
	return err
}

func (s *EventStorePG) GetInterestRequirements(ctx context.Context, eventID string) ([]string, error) {
	query := `
		SELECT interest_id
		FROM event_interest_requirements
		WHERE event_id = $1
		ORDER BY created_at`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var interestIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		interestIDs = append(interestIDs, id)
	}

	return interestIDs, nil
}

func (s *EventStorePG) UpdateInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Delete existing
	_, err = tx.ExecContext(ctx, "DELETE FROM event_interest_requirements WHERE event_id = $1", eventID)
	if err != nil {
		return err
	}

	// Insert new
	if len(interestIDs) > 0 {
		query := `
			INSERT INTO event_interest_requirements (event_id, interest_id, created_at)
			VALUES ($1, unnest($2::uuid[]), NOW())`
		_, err = tx.ExecContext(ctx, query, eventID, pq.Array(interestIDs))
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *EventStorePG) RemoveInterestRequirements(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_interest_requirements WHERE event_id = $1", eventID)
	return err
}

// Event image methods
func (s *EventStorePG) CreateEventImage(ctx context.Context, image *event.EventImage) error {
	query := `
		INSERT INTO event_images (id, event_id, file_id, alt_text, is_primary, display_order, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query, image.EventID, image.FileID, image.AltText, image.IsPrimary, image.DisplayOrder).
		Scan(&image.ID, &image.CreatedAt)
}

func (s *EventStorePG) GetEventImages(ctx context.Context, eventID string) ([]*event.EventImage, error) {
	query := `
		SELECT ei.id, ei.event_id, ei.file_id, ei.alt_text, ei.is_primary, ei.display_order, ei.created_at,
		       COALESCE(fu.storage_path, '') as url
		FROM event_images ei
		LEFT JOIN file_uploads fu ON ei.file_id = fu.id
		WHERE ei.event_id = $1
		ORDER BY ei.display_order, ei.created_at`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*event.EventImage
	for rows.Next() {
		img := &event.EventImage{}
		err := rows.Scan(&img.ID, &img.EventID, &img.FileID, &img.AltText, &img.IsPrimary, &img.DisplayOrder, &img.CreatedAt, &img.URL)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, nil
}

func (s *EventStorePG) UpdateEventImage(ctx context.Context, image *event.EventImage) error {
	query := `
		UPDATE event_images 
		SET alt_text = $1, is_primary = $2, display_order = $3
		WHERE id = $4`
	_, err := s.db.ExecContext(ctx, query, image.AltText, image.IsPrimary, image.DisplayOrder, image.ID)
	return err
}

func (s *EventStorePG) DeleteEventImage(ctx context.Context, imageID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_images WHERE id = $1", imageID)
	return err
}

func (s *EventStorePG) SetPrimaryImage(ctx context.Context, eventID, imageID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Unset all primary images for this event
	_, err = tx.ExecContext(ctx, "UPDATE event_images SET is_primary = false WHERE event_id = $1", eventID)
	if err != nil {
		return err
	}

	// Set the specified image as primary
	_, err = tx.ExecContext(ctx, "UPDATE event_images SET is_primary = true WHERE id = $1 AND event_id = $2", imageID, eventID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Event announcement methods
func (s *EventStorePG) CreateAnnouncement(ctx context.Context, announcement *event.EventAnnouncement) error {
	query := `
		INSERT INTO event_announcements (id, event_id, title, content, is_urgent, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query, announcement.EventID, announcement.Title, announcement.Content, announcement.IsUrgent).
		Scan(&announcement.ID, &announcement.CreatedAt)
}

func (s *EventStorePG) GetAnnouncements(ctx context.Context, eventID string) ([]*event.EventAnnouncement, error) {
	query := `
		SELECT id, event_id, title, content, is_urgent, created_at
		FROM event_announcements
		WHERE event_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announcements []*event.EventAnnouncement
	for rows.Next() {
		ann := &event.EventAnnouncement{}
		err := rows.Scan(&ann.ID, &ann.EventID, &ann.Title, &ann.Content, &ann.IsUrgent, &ann.CreatedAt)
		if err != nil {
			return nil, err
		}
		announcements = append(announcements, ann)
	}

	return announcements, nil
}

func (s *EventStorePG) UpdateAnnouncement(ctx context.Context, announcement *event.EventAnnouncement) error {
	query := `
		UPDATE event_announcements 
		SET title = $1, content = $2, is_urgent = $3
		WHERE id = $4`
	_, err := s.db.ExecContext(ctx, query, announcement.Title, announcement.Content, announcement.IsUrgent, announcement.ID)
	return err
}

func (s *EventStorePG) DeleteAnnouncement(ctx context.Context, announcementID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_announcements WHERE id = $1", announcementID)
	return err
}

// Event update/audit log methods
func (s *EventStorePG) LogUpdate(ctx context.Context, update *event.EventUpdate) error {
	query := `
		INSERT INTO event_updates (id, event_id, updated_by, field_name, old_value, new_value, update_type, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`
	return s.db.QueryRowContext(ctx, query, update.EventID, update.UpdatedBy, update.FieldName, update.OldValue, update.NewValue, update.UpdateType).
		Scan(&update.ID, &update.CreatedAt)
}

func (s *EventStorePG) GetUpdateHistory(ctx context.Context, eventID string, limit, offset int) ([]*event.EventUpdate, error) {
	query := `
		SELECT id, event_id, updated_by, field_name, old_value, new_value, update_type, created_at
		FROM event_updates
		WHERE event_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, query, eventID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updates []*event.EventUpdate
	for rows.Next() {
		upd := &event.EventUpdate{}
		err := rows.Scan(&upd.ID, &upd.EventID, &upd.UpdatedBy, &upd.FieldName, &upd.OldValue, &upd.NewValue, &upd.UpdateType, &upd.CreatedAt)
		if err != nil {
			return nil, err
		}
		updates = append(updates, upd)
	}

	return updates, nil
}

// Recurring event methods
func (s *EventStorePG) GetEventInstances(ctx context.Context, parentEventID string) ([]*event.Event, error) {
	return []*event.Event{}, nil // TODO: Implement
}

func (s *EventStorePG) GetUpcomingInstances(ctx context.Context, parentEventID string, limit int) ([]*event.Event, error) {
	return []*event.Event{}, nil // TODO: Implement
}

// Capacity management methods
func (s *EventStorePG) GetCurrentCapacity(ctx context.Context, eventID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM registrations
		WHERE event_id = $1 AND status = 'CONFIRMED'`

	var count int
	err := s.db.QueryRowContext(ctx, query, eventID).Scan(&count)
	return count, err
}

func (s *EventStorePG) IsAtCapacity(ctx context.Context, eventID string) (bool, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'CONFIRMED') >= 
			(SELECT max_capacity FROM events WHERE id = $1)`

	var atCapacity bool
	err := s.db.QueryRowContext(ctx, query, eventID).Scan(&atCapacity)
	return atCapacity, err
}

// Utility methods
func (s *EventStorePG) EventExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

func (s *EventStorePG) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM events WHERE slug = $1)", slug).Scan(&exists)
	return exists, err
}

func (s *EventStorePG) GenerateUniqueSlug(ctx context.Context, title string) (string, error) {
	// Create base slug from title
	reg := regexp.MustCompile("[^a-zA-Z0-9]+")
	baseSlug := strings.ToLower(reg.ReplaceAllString(title, "-"))
	baseSlug = strings.Trim(baseSlug, "-")

	// Check if base slug exists
	exists, err := s.SlugExists(ctx, baseSlug)
	if err != nil {
		return "", err
	}

	if !exists {
		return baseSlug, nil
	}

	// Try with numbers appended
	for i := 1; i < 1000; i++ {
		candidateSlug := fmt.Sprintf("%s-%d", baseSlug, i)
		exists, err := s.SlugExists(ctx, candidateSlug)
		if err != nil {
			return "", err
		}

		if !exists {
			return candidateSlug, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique slug for title: %s", title)
}
