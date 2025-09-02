-- Registrations table
CREATE TABLE registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    event_id UUID NOT NULL REFERENCES events(id),
    status TEXT NOT NULL DEFAULT 'PENDING_APPROVAL' CHECK (status IN (
        'PENDING_APPROVAL', 'CONFIRMED', 'WAITLISTED', 'CANCELLED', 'DECLINED', 'NO_SHOW', 'COMPLETED'
    )),
    personal_message TEXT,
    approval_notes TEXT,
    cancellation_reason TEXT,
    attendance_status TEXT NOT NULL DEFAULT 'REGISTERED' CHECK (attendance_status IN (
        'REGISTERED', 'CHECKED_IN', 'COMPLETED', 'NO_SHOW', 'CANCELLED'
    )),

    -- Important timestamps
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    checked_in_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    -- Waitlist management
    waitlist_position INTEGER,
    waitlist_promoted_at TIMESTAMPTZ,
    promotion_offered_at TIMESTAMPTZ,
    promotion_expires_at TIMESTAMPTZ,
    auto_promote BOOLEAN DEFAULT TRUE,

    -- Additional information
    emergency_contact_name TEXT,
    emergency_contact_phone TEXT,
    dietary_restrictions TEXT,
    accessibility_needs TEXT,

    -- Metadata
    checked_in_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Constraints
    UNIQUE(user_id, event_id),
    CONSTRAINT valid_confirmation CHECK (
        (status = 'CONFIRMED' AND confirmed_at IS NOT NULL) OR
        (status != 'CONFIRMED')
    ),
    CONSTRAINT valid_cancellation CHECK (
        (status = 'CANCELLED' AND cancelled_at IS NOT NULL) OR
        (status != 'CANCELLED')
    )
);

-- Registration skills snapshot (skills at time of registration)
CREATE TABLE registration_skills (
    registration_id UUID NOT NULL REFERENCES registrations(id) ON DELETE CASCADE,
    skill_name TEXT NOT NULL,
    proficiency TEXT NOT NULL CHECK (proficiency IN ('BEGINNER', 'INTERMEDIATE', 'ADVANCED', 'EXPERT')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (registration_id, skill_name)
);

-- Registration interests snapshot (interests at time of registration)
CREATE TABLE registration_interests (
    registration_id UUID NOT NULL REFERENCES registrations(id) ON DELETE CASCADE,
    interest_id UUID NOT NULL REFERENCES interests(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (registration_id, interest_id)
);

-- Attendance records (detailed check-in/out tracking)
CREATE TABLE attendance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_id UUID NOT NULL REFERENCES registrations(id),
    status TEXT NOT NULL CHECK (status IN ('CHECKED_IN', 'CHECKED_OUT', 'COMPLETED', 'NO_SHOW')),
    checked_in_at TIMESTAMPTZ,
    checked_out_at TIMESTAMPTZ,
    checked_in_by UUID REFERENCES users(id),
    location_verified BOOLEAN DEFAULT FALSE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Registration status changes (audit trail)
CREATE TABLE registration_status_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_id UUID NOT NULL REFERENCES registrations(id),
    old_status TEXT,
    new_status TEXT NOT NULL,
    changed_by UUID REFERENCES users(id),
    reason TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Waitlist management
CREATE TABLE waitlist_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_id UUID NOT NULL REFERENCES registrations(id) UNIQUE,
    position INTEGER NOT NULL,
    priority_score DECIMAL DEFAULT 0,
    auto_promote BOOLEAN DEFAULT TRUE,
    promotion_offered_at TIMESTAMPTZ,
    promotion_expires_at TIMESTAMPTZ,
    declined_promotion BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Registration conflicts
CREATE TABLE registration_conflicts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    primary_event_id UUID NOT NULL REFERENCES events(id),
    conflicting_event_id UUID NOT NULL REFERENCES events(id),
    conflict_type TEXT NOT NULL CHECK (conflict_type IN (
        'TIME_OVERLAP', 'LOCATION_CONFLICT', 'TRAVEL_TIME_CONFLICT', 'SKILL_OVERCOMMITMENT'
    )),
    severity TEXT NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    resolved BOOLEAN DEFAULT FALSE,
    resolution_notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_registrations_user_id ON registrations(user_id);
CREATE INDEX idx_registrations_event_id ON registrations(event_id);
CREATE INDEX idx_registrations_status ON registrations(status);
CREATE INDEX idx_registrations_applied_at ON registrations(applied_at);
CREATE INDEX idx_registrations_waitlist ON registrations(waitlist_position) WHERE waitlist_position IS NOT NULL;
CREATE INDEX idx_attendance_records_registration_id ON attendance_records(registration_id);
CREATE INDEX idx_waitlist_entries_position ON waitlist_entries(position);
CREATE INDEX idx_registration_conflicts_user_event ON registration_conflicts(user_id, primary_event_id);
