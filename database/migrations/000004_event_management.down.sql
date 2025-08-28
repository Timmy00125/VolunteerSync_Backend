-- Phase 4: Event Management - Rollback
-- This file removes all event management related tables and functions

-- Drop triggers and functions
DROP TRIGGER IF EXISTS update_events_updated_at ON events;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS event_updates CASCADE;
DROP TABLE IF EXISTS event_announcements CASCADE;
DROP TABLE IF EXISTS event_images CASCADE;
DROP TABLE IF EXISTS event_training_requirements CASCADE;
DROP TABLE IF EXISTS event_interest_requirements CASCADE;
DROP TABLE IF EXISTS event_skill_requirements CASCADE;
DROP TABLE IF EXISTS events CASCADE;