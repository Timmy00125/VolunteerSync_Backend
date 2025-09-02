DROP INDEX IF EXISTS idx_registration_conflicts_user_event;
DROP INDEX IF EXISTS idx_waitlist_entries_position;
DROP INDEX IF EXISTS idx_attendance_records_registration_id;
DROP INDEX IF EXISTS idx_registrations_waitlist;
DROP INDEX IF EXISTS idx_registrations_applied_at;
DROP INDEX IF EXISTS idx_registrations_status;
DROP INDEX IF EXISTS idx_registrations_event_id;
DROP INDEX IF EXISTS idx_registrations_user_id;

DROP TABLE IF EXISTS registration_conflicts;
DROP TABLE IF EXISTS waitlist_entries;
DROP TABLE IF EXISTS registration_status_changes;
DROP TABLE IF EXISTS attendance_records;
DROP TABLE IF EXISTS registration_interests;
DROP TABLE IF EXISTS registration_skills;
DROP TABLE IF EXISTS registrations;
