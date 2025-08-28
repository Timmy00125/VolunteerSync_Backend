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

// EventStore implements the event.Repository interface using PostgreSQL
type EventStore struct {
	db *sql.DB
}

// NewEventStore creates a new PostgreSQL event store
func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{db: db}
}

// Create creates a new event in the database
func (s *EventStore) Create(ctx context.Context, e *event.Event) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Convert recurrence rule to JSON if present
	var recurrenceRuleJSON []byte
	if e.RecurrenceRule != nil {
		recurrenceRuleJSON, err = json.Marshal(e.RecurrenceRule)
		if err != nil {
			return fmt.Errorf("failed to marshal recurrence rule: %w", err)
		}
	}

	// Insert main event record
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

	// Prepare event data for insertion
	eventData := map[string]interface{}{
		"id":                        e.ID,
		"title":                     e.Title,
		"description":               e.Description,
		"short_description":         e.ShortDescription,
		"organizer_id":              e.OrganizerID,
		"status":                    e.Status,
		"start_time":                e.StartTime,
		"end_time":                  e.EndTime,
		"location_name":             e.Location.Name,
		"location_address":          e.Location.Address,
		"location_city":             e.Location.City,
		"location_state":            e.Location.State,
		"location_country":          e.Location.Country,
		"location_zip_code":         e.Location.ZipCode,
		"location_latitude":         nil,
		"location_longitude":        nil,
		"location_instructions":     e.Location.Instructions,
		"is_remote":                 e.Location.IsRemote,
		"min_capacity":              e.Capacity.Minimum,
		"max_capacity":              e.Capacity.Maximum,
		"waitlist_enabled":          e.Capacity.WaitlistEnabled,
		"minimum_age":               e.Requirements.MinimumAge,
		"background_check_required": e.Requirements.BackgroundCheck,
		"physical_requirements":     e.Requirements.PhysicalRequirements,
		"category":                  e.Category,
		"time_commitment":           e.TimeCommitment,
		"tags":                      pq.Array(e.Tags),
		"registration_opens_at":     e.RegistrationSettings.OpensAt,
		"registration_closes_at":    e.RegistrationSettings.ClosesAt,
		"requires_approval":         e.RegistrationSettings.RequiresApproval,
		"confirmation_required":     e.RegistrationSettings.ConfirmationRequired,
		"cancellation_deadline":     e.RegistrationSettings.CancellationDeadline,
		"parent_event_id":           e.ParentEventID,
		"recurrence_rule":           recurrenceRuleJSON,
		"slug":                      e.Slug,
		"share_url":                 e.ShareURL,
		"created_at":                e.CreatedAt,
		"updated_at":                e.UpdatedAt,
		"published_at":              e.PublishedAt,
	}

	// Set coordinates if available
	if e.Location.Coordinates != nil {
		eventData["location_latitude"] = e.Location.Coordinates.Latitude
		eventData["location_longitude"] = e.Location.Coordinates.Longitude
	}

	// Prepare coordinates
	var lat, lng *float64
	if e.Location.Coordinates != nil {
		lat = &e.Location.Coordinates.Latitude
		lng = &e.Location.Coordinates.Longitude
	}

	_, err = tx.ExecContext(ctx, query,
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
		recurrenceRuleJSON, e.Slug, e.ShareURL, e.CreatedAt, e.UpdatedAt, e.PublishedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	// Insert skill requirements
	if len(e.Requirements.Skills) > 0 {
		for _, skill := range e.Requirements.Skills {
			skill.EventID = e.ID
			if err := s.createSkillRequirement(ctx, tx, &skill); err != nil {
				return fmt.Errorf("failed to create skill requirement: %w", err)
			}
		}
	}

	// Insert training requirements
	if len(e.Requirements.Training) > 0 {
		for _, training := range e.Requirements.Training {
			training.EventID = e.ID
			if err := s.createTrainingRequirement(ctx, tx, &training); err != nil {
				return fmt.Errorf("failed to create training requirement: %w", err)
			}
		}
	}

	// Insert interest requirements
	if len(e.Requirements.Interests) > 0 {
		if err := s.addInterestRequirements(ctx, tx, e.ID, e.Requirements.Interests); err != nil {
			return fmt.Errorf("failed to create interest requirements: %w", err)
		}
	}

	return tx.Commit()
}

// GetByID retrieves an event by its ID
func (s *EventStore) GetByID(ctx context.Context, id string) (*event.Event, error) {
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

	var recurrenceRuleJSON []byte
	var tags pq.StringArray
	var lat, lng sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.Title, &e.Description, &e.ShortDescription, &e.OrganizerID, &e.Status,
		&e.StartTime, &e.EndTime, &e.Location.Name, &e.Location.Address, &e.Location.City,
		&e.Location.State, &e.Location.Country, &e.Location.ZipCode, &lat, &lng,
		&e.Location.Instructions, &e.Location.IsRemote, &e.Capacity.Minimum,
		&e.Capacity.Maximum, &e.Capacity.WaitlistEnabled, &e.Requirements.MinimumAge,
		&e.Requirements.BackgroundCheck, &e.Requirements.PhysicalRequirements,
		&e.Category, &e.TimeCommitment, &tags, &e.RegistrationSettings.OpensAt,
		&e.RegistrationSettings.ClosesAt, &e.RegistrationSettings.RequiresApproval,
		&e.RegistrationSettings.ConfirmationRequired, &e.RegistrationSettings.CancellationDeadline,
		&e.ParentEventID, &recurrenceRuleJSON, &e.Slug, &e.ShareURL,
		&e.CreatedAt, &e.UpdatedAt, &e.PublishedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Set coordinates if available
	if lat.Valid && lng.Valid {
		e.Location.Coordinates = &event.Coordinates{
			Latitude:  lat.Float64,
			Longitude: lng.Float64,
		}
	}

	// Convert tags
	e.Tags = []string(tags)

	// Parse recurrence rule if present
	if len(recurrenceRuleJSON) > 0 {
		var rule event.RecurrenceRule
		if err := json.Unmarshal(recurrenceRuleJSON, &rule); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recurrence rule: %w", err)
		}
		e.RecurrenceRule = &rule
	}

	// Load related data
	if err := s.loadEventRelations(ctx, e); err != nil {
		return nil, fmt.Errorf("failed to load event relations: %w", err)
	}

	return e, nil
}

// GetBySlug retrieves an event by its slug
func (s *EventStore) GetBySlug(ctx context.Context, slug string) (*event.Event, error) {
	// First get the ID by slug
	var id string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM events WHERE slug = $1", slug).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found with slug: %s", slug)
		}
		return nil, fmt.Errorf("failed to get event by slug: %w", err)
	}

	return s.GetByID(ctx, id)
}

// Update updates an existing event
func (s *EventStore) Update(ctx context.Context, e *event.Event) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Convert recurrence rule to JSON if present
	var recurrenceRuleJSON []byte
	if e.RecurrenceRule != nil {
		recurrenceRuleJSON, err = json.Marshal(e.RecurrenceRule)
		if err != nil {
			return fmt.Errorf("failed to marshal recurrence rule: %w", err)
		}
	}

	// Update main event record
	query := `
		UPDATE events SET
			title = :title, description = :description, short_description = :short_description,
			location_name = :location_name, location_address = :location_address,
			location_city = :location_city, location_state = :location_state,
			location_country = :location_country, location_zip_code = :location_zip_code,
			location_latitude = :location_latitude, location_longitude = :location_longitude,
			location_instructions = :location_instructions, is_remote = :is_remote,
			min_capacity = :min_capacity, max_capacity = :max_capacity,
			waitlist_enabled = :waitlist_enabled, minimum_age = :minimum_age,
			background_check_required = :background_check_required,
			physical_requirements = :physical_requirements, category = :category,
			time_commitment = :time_commitment, tags = :tags,
			registration_opens_at = :registration_opens_at,
			registration_closes_at = :registration_closes_at,
			requires_approval = :requires_approval,
			confirmation_required = :confirmation_required,
			cancellation_deadline = :cancellation_deadline,
			recurrence_rule = :recurrence_rule, updated_at = NOW()
		WHERE id = :id`

	// Prepare event data for update
	eventData := map[string]interface{}{
		"id":                        e.ID,
		"title":                     e.Title,
		"description":               e.Description,
		"short_description":         e.ShortDescription,
		"location_name":             e.Location.Name,
		"location_address":          e.Location.Address,
		"location_city":             e.Location.City,
		"location_state":            e.Location.State,
		"location_country":          e.Location.Country,
		"location_zip_code":         e.Location.ZipCode,
		"location_latitude":         nil,
		"location_longitude":        nil,
		"location_instructions":     e.Location.Instructions,
		"is_remote":                 e.Location.IsRemote,
		"min_capacity":              e.Capacity.Minimum,
		"max_capacity":              e.Capacity.Maximum,
		"waitlist_enabled":          e.Capacity.WaitlistEnabled,
		"minimum_age":               e.Requirements.MinimumAge,
		"background_check_required": e.Requirements.BackgroundCheck,
		"physical_requirements":     e.Requirements.PhysicalRequirements,
		"category":                  e.Category,
		"time_commitment":           e.TimeCommitment,
		"tags":                      pq.Array(e.Tags),
		"registration_opens_at":     e.RegistrationSettings.OpensAt,
		"registration_closes_at":    e.RegistrationSettings.ClosesAt,
		"requires_approval":         e.RegistrationSettings.RequiresApproval,
		"confirmation_required":     e.RegistrationSettings.ConfirmationRequired,
		"cancellation_deadline":     e.RegistrationSettings.CancellationDeadline,
		"recurrence_rule":           recurrenceRuleJSON,
	}

	// Set coordinates if available
	if e.Location.Coordinates != nil {
		eventData["location_latitude"] = e.Location.Coordinates.Latitude
		eventData["location_longitude"] = e.Location.Coordinates.Longitude
	}

	_, err = tx.NamedExec(query, eventData)
	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	return tx.Commit()
}

// Delete soft deletes an event by setting its status to ARCHIVED
func (s *EventStore) Delete(ctx context.Context, id string) error {
	query := `UPDATE events SET status = 'ARCHIVED', updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	return nil
}

// List retrieves events with filtering, sorting, and pagination
func (s *EventStore) List(ctx context.Context, filter event.EventSearchFilter, sort *event.EventSortInput, limit, offset int) (*event.EventConnection, error) {
	// Build the base query
	baseQuery := `
		FROM events e 
		WHERE e.status != 'ARCHIVED'`

	args := []interface{}{}
	argCount := 0

	// Apply filters
	whereConditions := []string{}

	// Text search filter
	if filter.Query != nil && *filter.Query != "" {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf(`
			to_tsvector('english', e.title || ' ' || e.description || ' ' || COALESCE(e.short_description, ''))
			@@ plainto_tsquery('english', $%d)`, argCount))
		args = append(args, *filter.Query)
	}

	// Location filter
	if filter.Location != nil {
		argCount += 3
		whereConditions = append(whereConditions, fmt.Sprintf(`
			e.location_latitude IS NOT NULL 
			AND e.location_longitude IS NOT NULL
			AND ST_DWithin(
				ST_Point(e.location_longitude, e.location_latitude)::geography,
				ST_Point($%d, $%d)::geography,
				$%d * 1000
			)`, argCount-2, argCount-1, argCount))
		args = append(args, filter.Location.Center.Longitude, filter.Location.Center.Latitude, filter.Location.Radius)
	}

	// Date range filter
	if filter.DateRange != nil {
		argCount += 2
		whereConditions = append(whereConditions, fmt.Sprintf(`
			e.start_time >= $%d AND e.start_time <= $%d`, argCount-1, argCount))
		args = append(args, filter.DateRange.StartDate, filter.DateRange.EndDate)
	}

	// Category filter
	if len(filter.Categories) > 0 {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf(`e.category = ANY($%d)`, argCount))
		categories := make([]string, len(filter.Categories))
		for i, cat := range filter.Categories {
			categories[i] = string(cat)
		}
		args = append(args, pq.Array(categories))
	}

	// Time commitment filter
	if len(filter.TimeCommitment) > 0 {
		argCount++
		whereConditions = append(whereConditions, fmt.Sprintf(`e.time_commitment = ANY($%d)`, argCount))
		commitments := make([]string, len(filter.TimeCommitment))
		for i, tc := range filter.TimeCommitment {
			commitments[i] = string(tc)
		}
		args = append(args, pq.Array(commitments))
	}

	// Available spots filter
	if filter.HasAvailableSpots != nil && *filter.HasAvailableSpots {
		whereConditions = append(whereConditions, `
			(SELECT COUNT(*) FROM registrations r WHERE r.event_id = e.id AND r.status = 'CONFIRMED') < e.max_capacity`)
	}

	// Combine where conditions
	if len(whereConditions) > 0 {
		baseQuery += " AND " + strings.Join(whereConditions, " AND ")
	}

	// Add ordering
	orderBy := "ORDER BY e.start_time ASC"
	if sort != nil {
		direction := "ASC"
		if sort.Direction == event.SortDirectionDESC {
			direction = "DESC"
		}

		switch sort.Field {
		case event.EventSortFieldStartTime:
			orderBy = fmt.Sprintf("ORDER BY e.start_time %s", direction)
		case event.EventSortFieldCreatedAt:
			orderBy = fmt.Sprintf("ORDER BY e.created_at %s", direction)
		case event.EventSortFieldCapacityRemaining:
			orderBy = fmt.Sprintf(`ORDER BY (e.max_capacity - COALESCE((
				SELECT COUNT(*) FROM registrations r 
				WHERE r.event_id = e.id AND r.status = 'CONFIRMED'
			), 0)) %s`, direction)
		}
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int
	err := s.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get events
	selectQuery := `
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
		` + baseQuery + " " + orderBy + fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := s.db.QueryxContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	events := []*event.Event{}
	for rows.Next() {
		e := &event.Event{}
		var recurrenceRuleJSON []byte
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
			&e.ParentEventID, &recurrenceRuleJSON, &e.Slug, &e.ShareURL,
			&e.CreatedAt, &e.UpdatedAt, &e.PublishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Set coordinates if available
		if lat.Valid && lng.Valid {
			e.Location.Coordinates = &event.Coordinates{
				Latitude:  lat.Float64,
				Longitude: lng.Float64,
			}
		}

		// Convert tags
		e.Tags = []string(tags)

		// Parse recurrence rule if present
		if len(recurrenceRuleJSON) > 0 {
			var rule event.RecurrenceRule
			if err := json.Unmarshal(recurrenceRuleJSON, &rule); err != nil {
				return nil, fmt.Errorf("failed to unmarshal recurrence rule: %w", err)
			}
			e.RecurrenceRule = &rule
		}

		// Load related data
		if err := s.loadEventRelations(ctx, e); err != nil {
			return nil, fmt.Errorf("failed to load event relations: %w", err)
		}

		events = append(events, e)
	}

	// Build response
	edges := make([]event.EventEdge, len(events))
	for i, e := range events {
		edges[i] = event.EventEdge{
			Node:   *e,
			Cursor: e.ID, // Simple cursor implementation
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
	}, nil
}

// Helper functions for loading event relations
func (s *EventStore) loadEventRelations(ctx context.Context, e *event.Event) error {
	// Load skill requirements
	skillReqs, err := s.GetSkillRequirements(ctx, e.ID)
	if err != nil {
		return err
	}
	e.Requirements.Skills = make([]event.SkillRequirement, len(skillReqs))
	for i, req := range skillReqs {
		e.Requirements.Skills[i] = *req
	}

	// Load training requirements
	trainingReqs, err := s.GetTrainingRequirements(ctx, e.ID)
	if err != nil {
		return err
	}
	e.Requirements.Training = make([]event.TrainingRequirement, len(trainingReqs))
	for i, req := range trainingReqs {
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

	// Get current capacity from registrations
	currentCapacity, err := s.GetCurrentCapacity(ctx, e.ID)
	if err != nil {
		return err
	}
	e.Capacity.Current = currentCapacity

	return nil
}

// Additional method stubs - implementing remaining interface methods
func (s *EventStore) GetByOrganizer(ctx context.Context, organizerID string) ([]*event.Event, error) {
	// Implementation would be similar to List but with organizer filter
	// For brevity, returning empty slice for now
	return []*event.Event{}, nil
}

func (s *EventStore) GetFeatured(ctx context.Context, limit int) ([]*event.Event, error) {
	// Implementation would fetch featured events based on criteria
	return []*event.Event{}, nil
}

func (s *EventStore) GetNearby(ctx context.Context, lat, lng, radius float64, limit int) ([]*event.Event, error) {
	// Implementation would use PostGIS for nearby search
	return []*event.Event{}, nil
}

func (s *EventStore) UpdateStatus(ctx context.Context, eventID string, status event.EventStatus) error {
	query := `UPDATE events SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, status, eventID)
	return err
}

func (s *EventStore) GetByStatus(ctx context.Context, status event.EventStatus, limit, offset int) ([]*event.Event, error) {
	return []*event.Event{}, nil
}

// Skill requirement methods
func (s *EventStore) CreateSkillRequirement(ctx context.Context, req *event.SkillRequirement) error {
	return s.createSkillRequirement(ctx, s.db, req)
}

func (s *EventStore) createSkillRequirement(ctx context.Context, tx sqlx.ExtContext, req *event.SkillRequirement) error {
	query := `
		INSERT INTO event_skill_requirements (id, event_id, skill_name, proficiency, required, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		RETURNING id, created_at`

	err := tx.QueryRowxContext(ctx, query, req.EventID, req.Skill, req.Proficiency, req.Required).
		Scan(&req.ID, &req.CreatedAt)

	return err
}

func (s *EventStore) GetSkillRequirements(ctx context.Context, eventID string) ([]*event.SkillRequirement, error) {
	query := `
		SELECT id, event_id, skill_name, proficiency, required, created_at
		FROM event_skill_requirements
		WHERE event_id = $1
		ORDER BY created_at`

	rows, err := s.db.QueryxContext(ctx, query, eventID)
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

func (s *EventStore) UpdateSkillRequirements(ctx context.Context, eventID string, requirements []*event.SkillRequirement) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing requirements
	_, err = tx.ExecContext(ctx, "DELETE FROM event_skill_requirements WHERE event_id = $1", eventID)
	if err != nil {
		return err
	}

	// Insert new requirements
	for _, req := range requirements {
		req.EventID = eventID
		if err := s.createSkillRequirement(ctx, tx, req); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *EventStore) DeleteSkillRequirements(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_skill_requirements WHERE event_id = $1", eventID)
	return err
}

// Training requirement methods
func (s *EventStore) CreateTrainingRequirement(ctx context.Context, req *event.TrainingRequirement) error {
	return s.createTrainingRequirement(ctx, s.db, req)
}

func (s *EventStore) createTrainingRequirement(ctx context.Context, tx sqlx.ExtContext, req *event.TrainingRequirement) error {
	query := `
		INSERT INTO event_training_requirements (id, event_id, name, description, required, provided_by_organizer, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`

	err := tx.QueryRowxContext(ctx, query, req.EventID, req.Name, req.Description, req.Required, req.ProvidedByOrganizer).
		Scan(&req.ID, &req.CreatedAt)

	return err
}

func (s *EventStore) GetTrainingRequirements(ctx context.Context, eventID string) ([]*event.TrainingRequirement, error) {
	query := `
		SELECT id, event_id, name, description, required, provided_by_organizer, created_at
		FROM event_training_requirements
		WHERE event_id = $1
		ORDER BY created_at`

	rows, err := s.db.QueryxContext(ctx, query, eventID)
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

func (s *EventStore) UpdateTrainingRequirements(ctx context.Context, eventID string, requirements []*event.TrainingRequirement) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing requirements
	_, err = tx.ExecContext(ctx, "DELETE FROM event_training_requirements WHERE event_id = $1", eventID)
	if err != nil {
		return err
	}

	// Insert new requirements
	for _, req := range requirements {
		req.EventID = eventID
		if err := s.createTrainingRequirement(ctx, tx, req); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *EventStore) DeleteTrainingRequirements(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_training_requirements WHERE event_id = $1", eventID)
	return err
}

// Interest requirement methods
func (s *EventStore) AddInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error {
	return s.addInterestRequirements(ctx, s.db, eventID, interestIDs)
}

func (s *EventStore) addInterestRequirements(ctx context.Context, tx sqlx.ExtContext, eventID string, interestIDs []string) error {
	if len(interestIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO event_interest_requirements (event_id, interest_id, created_at)
		VALUES ($1, unnest($2::uuid[]), NOW())`

	_, err := tx.ExecContext(ctx, query, eventID, pq.Array(interestIDs))
	return err
}

func (s *EventStore) GetInterestRequirements(ctx context.Context, eventID string) ([]string, error) {
	query := `
		SELECT interest_id
		FROM event_interest_requirements
		WHERE event_id = $1
		ORDER BY created_at`

	var interestIDs []string
	err := s.db.SelectContext(ctx, &interestIDs, query, eventID)
	return interestIDs, err
}

func (s *EventStore) UpdateInterestRequirements(ctx context.Context, eventID string, interestIDs []string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing requirements
	_, err = tx.ExecContext(ctx, "DELETE FROM event_interest_requirements WHERE event_id = $1", eventID)
	if err != nil {
		return err
	}

	// Insert new requirements
	if err := s.addInterestRequirements(ctx, tx, eventID, interestIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *EventStore) RemoveInterestRequirements(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_interest_requirements WHERE event_id = $1", eventID)
	return err
}

// Event image methods
func (s *EventStore) CreateEventImage(ctx context.Context, image *event.EventImage) error {
	query := `
		INSERT INTO event_images (id, event_id, file_id, alt_text, is_primary, display_order, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`

	err := s.db.QueryRowxContext(ctx, query, image.EventID, image.FileID, image.AltText, image.IsPrimary, image.DisplayOrder).
		Scan(&image.ID, &image.CreatedAt)

	return err
}

func (s *EventStore) GetEventImages(ctx context.Context, eventID string) ([]*event.EventImage, error) {
	query := `
		SELECT ei.id, ei.event_id, ei.file_id, ei.alt_text, ei.is_primary, ei.display_order, ei.created_at,
		       fu.storage_path as url
		FROM event_images ei
		JOIN file_uploads fu ON ei.file_id = fu.id
		WHERE ei.event_id = $1
		ORDER BY ei.display_order, ei.created_at`

	rows, err := s.db.QueryxContext(ctx, query, eventID)
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

func (s *EventStore) UpdateEventImage(ctx context.Context, image *event.EventImage) error {
	query := `
		UPDATE event_images 
		SET alt_text = $1, is_primary = $2, display_order = $3
		WHERE id = $4`

	_, err := s.db.ExecContext(ctx, query, image.AltText, image.IsPrimary, image.DisplayOrder, image.ID)
	return err
}

func (s *EventStore) DeleteEventImage(ctx context.Context, imageID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_images WHERE id = $1", imageID)
	return err
}

func (s *EventStore) SetPrimaryImage(ctx context.Context, eventID, imageID string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
func (s *EventStore) CreateAnnouncement(ctx context.Context, announcement *event.EventAnnouncement) error {
	query := `
		INSERT INTO event_announcements (id, event_id, title, content, is_urgent, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		RETURNING id, created_at`

	err := s.db.QueryRowxContext(ctx, query, announcement.EventID, announcement.Title, announcement.Content, announcement.IsUrgent).
		Scan(&announcement.ID, &announcement.CreatedAt)

	return err
}

func (s *EventStore) GetAnnouncements(ctx context.Context, eventID string) ([]*event.EventAnnouncement, error) {
	query := `
		SELECT id, event_id, title, content, is_urgent, created_at
		FROM event_announcements
		WHERE event_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.QueryxContext(ctx, query, eventID)
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

func (s *EventStore) UpdateAnnouncement(ctx context.Context, announcement *event.EventAnnouncement) error {
	query := `
		UPDATE event_announcements 
		SET title = $1, content = $2, is_urgent = $3
		WHERE id = $4`

	_, err := s.db.ExecContext(ctx, query, announcement.Title, announcement.Content, announcement.IsUrgent, announcement.ID)
	return err
}

func (s *EventStore) DeleteAnnouncement(ctx context.Context, announcementID string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM event_announcements WHERE id = $1", announcementID)
	return err
}

// Event update/audit log methods
func (s *EventStore) LogUpdate(ctx context.Context, update *event.EventUpdate) error {
	query := `
		INSERT INTO event_updates (id, event_id, updated_by, field_name, old_value, new_value, update_type, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`

	err := s.db.QueryRowxContext(ctx, query, update.EventID, update.UpdatedBy, update.FieldName, update.OldValue, update.NewValue, update.UpdateType).
		Scan(&update.ID, &update.CreatedAt)

	return err
}

func (s *EventStore) GetUpdateHistory(ctx context.Context, eventID string, limit, offset int) ([]*event.EventUpdate, error) {
	query := `
		SELECT id, event_id, updated_by, field_name, old_value, new_value, update_type, created_at
		FROM event_updates
		WHERE event_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryxContext(ctx, query, eventID, limit, offset)
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
func (s *EventStore) GetEventInstances(ctx context.Context, parentEventID string) ([]*event.Event, error) {
	// Implementation would fetch all events with the given parent_event_id
	return []*event.Event{}, nil
}

func (s *EventStore) GetUpcomingInstances(ctx context.Context, parentEventID string, limit int) ([]*event.Event, error) {
	// Implementation would fetch upcoming events with the given parent_event_id
	return []*event.Event{}, nil
}

// Capacity management methods
func (s *EventStore) GetCurrentCapacity(ctx context.Context, eventID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM registrations
		WHERE event_id = $1 AND status = 'CONFIRMED'`

	var count int
	err := s.db.GetContext(ctx, &count, query, eventID)
	return count, err
}

func (s *EventStore) IsAtCapacity(ctx context.Context, eventID string) (bool, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM registrations WHERE event_id = $1 AND status = 'CONFIRMED') >= 
			(SELECT max_capacity FROM events WHERE id = $1)`

	var atCapacity bool
	err := s.db.GetContext(ctx, &atCapacity, query, eventID)
	return atCapacity, err
}

// Utility methods
func (s *EventStore) EventExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := s.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM events WHERE id = $1)", id)
	return exists, err
}

func (s *EventStore) SlugExists(ctx context.Context, slug string) (bool, error) {
	var exists bool
	err := s.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM events WHERE slug = $1)", slug)
	return exists, err
}

func (s *EventStore) GenerateUniqueSlug(ctx context.Context, title string) (string, error) {
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
