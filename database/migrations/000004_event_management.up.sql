-- Phase 4: Event Management - Complete event system
-- Using gen_random_uuid() for better randomness (available in PostgreSQL 13+)

-- Drop the existing simple events table if it exists
DROP TABLE IF EXISTS events CASCADE;

-- Events table with comprehensive structure
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    short_description TEXT,
    organizer_id UUID NOT NULL REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'PUBLISHED', 'CANCELLED', 'COMPLETED', 'ARCHIVED')),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,

    -- Location fields
    location_name TEXT NOT NULL,
    location_address TEXT NOT NULL,
    location_city TEXT NOT NULL,
    location_state TEXT,
    location_country TEXT NOT NULL,
    location_zip_code TEXT,
    location_latitude DECIMAL(9,6),
    location_longitude DECIMAL(9,6),
    location_instructions TEXT,
    is_remote BOOLEAN DEFAULT FALSE,

    -- Capacity fields
    min_capacity INTEGER NOT NULL DEFAULT 1,
    max_capacity INTEGER NOT NULL,
    waitlist_enabled BOOLEAN DEFAULT FALSE,

    -- Requirements
    minimum_age INTEGER,
    background_check_required BOOLEAN DEFAULT FALSE,
    physical_requirements TEXT,

    -- Event metadata
    category TEXT NOT NULL CHECK (category IN (
        'ENVIRONMENT', 'EDUCATION', 'HEALTH', 'COMMUNITY_SERVICE', 
        'DISASTER_RELIEF', 'ANIMAL_WELFARE', 'ARTS_CULTURE', 'TECHNOLOGY',
        'SPORTS_RECREATION', 'SENIOR_CARE', 'YOUTH_MENTORING', 'FOOD_SECURITY'
    )),
    time_commitment TEXT NOT NULL CHECK (time_commitment IN (
        'ONE_TIME', 'SHORT_TERM', 'MEDIUM_TERM', 'LONG_TERM', 'ONGOING'
    )),
    tags TEXT[], -- PostgreSQL array

    -- Registration settings
    registration_opens_at TIMESTAMPTZ,
    registration_closes_at TIMESTAMPTZ NOT NULL,
    requires_approval BOOLEAN DEFAULT FALSE,
    confirmation_required BOOLEAN DEFAULT TRUE,
    cancellation_deadline TIMESTAMPTZ,

    -- Recurrence (for recurring events)
    parent_event_id UUID REFERENCES events(id),
    recurrence_rule JSONB,

    -- SEO and sharing
    slug TEXT UNIQUE,
    share_url TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    published_at TIMESTAMPTZ,

    -- Constraints
    CONSTRAINT valid_time_range CHECK (end_time > start_time),
    CONSTRAINT valid_capacity CHECK (max_capacity >= min_capacity),
    CONSTRAINT valid_registration_timing CHECK (registration_closes_at <= start_time)
);

-- Event skill requirements
CREATE TABLE event_skill_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    skill_name TEXT NOT NULL,
    proficiency TEXT NOT NULL CHECK (proficiency IN ('BEGINNER', 'INTERMEDIATE', 'ADVANCED', 'EXPERT')),
    required BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Event interest requirements (junction table with interests)
CREATE TABLE event_interest_requirements (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    interest_id UUID NOT NULL REFERENCES interests(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (event_id, interest_id)
);

-- Event training requirements
CREATE TABLE event_training_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    required BOOLEAN DEFAULT FALSE,
    provided_by_organizer BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Event images
CREATE TABLE event_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES file_uploads(id),
    alt_text TEXT,
    is_primary BOOLEAN DEFAULT FALSE,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Event announcements
CREATE TABLE event_announcements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    is_urgent BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Event updates log (audit trail)
CREATE TABLE event_updates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    updated_by UUID NOT NULL REFERENCES users(id),
    field_name TEXT NOT NULL,
    old_value TEXT,
    new_value TEXT,
    update_type TEXT NOT NULL CHECK (update_type IN ('MINOR', 'MAJOR', 'STATUS_CHANGE')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Performance indexes
CREATE INDEX idx_events_organizer_id ON events(organizer_id);
CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_location ON events(location_city, location_state, location_country);
CREATE INDEX idx_events_category ON events(category);
CREATE INDEX idx_events_published_at ON events(published_at) WHERE published_at IS NOT NULL;
CREATE INDEX idx_events_tags ON events USING GIN(tags);
CREATE INDEX idx_events_slug ON events(slug) WHERE slug IS NOT NULL;

-- Location indexes for geographic queries (using standard PostgreSQL)
CREATE INDEX idx_events_location_lat_lng ON events(location_latitude, location_longitude) 
    WHERE location_latitude IS NOT NULL AND location_longitude IS NOT NULL;
CREATE INDEX idx_events_location_city ON events(location_city);
CREATE INDEX idx_events_location_state ON events(location_state);
CREATE INDEX idx_events_location_country ON events(location_country);

-- Full-text search index
CREATE INDEX idx_events_search ON events USING GIN(
    to_tsvector('english', title || ' ' || description || ' ' || COALESCE(short_description, ''))
);

-- Indexes for related tables
CREATE INDEX idx_event_skill_requirements_event_id ON event_skill_requirements(event_id);
CREATE INDEX idx_event_skill_requirements_skill_name ON event_skill_requirements(skill_name);

CREATE INDEX idx_event_interest_requirements_event_id ON event_interest_requirements(event_id);
CREATE INDEX idx_event_interest_requirements_interest_id ON event_interest_requirements(interest_id);

CREATE INDEX idx_event_training_requirements_event_id ON event_training_requirements(event_id);

CREATE INDEX idx_event_images_event_id ON event_images(event_id);
CREATE INDEX idx_event_images_is_primary ON event_images(event_id) WHERE is_primary = true;

CREATE INDEX idx_event_announcements_event_id ON event_announcements(event_id);
CREATE INDEX idx_event_announcements_created_at ON event_announcements(created_at);

CREATE INDEX idx_event_updates_event_id ON event_updates(event_id);
CREATE INDEX idx_event_updates_created_at ON event_updates(created_at);

-- Add trigger to update updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_events_updated_at 
    BEFORE UPDATE ON events 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();