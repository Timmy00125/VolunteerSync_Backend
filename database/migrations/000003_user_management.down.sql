-- Phase 3: User Management - revert schema updates

-- Drop new tables first due to FKs
DROP TABLE IF EXISTS file_uploads;
DROP TABLE IF EXISTS user_activity_logs;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS user_skills;
DROP TABLE IF EXISTS user_interests;
DROP TABLE IF EXISTS interests;
DROP TABLE IF EXISTS interest_categories;

-- Remove columns from users table
ALTER TABLE users DROP COLUMN IF EXISTS is_verified;
ALTER TABLE users DROP COLUMN IF EXISTS last_active_at;
ALTER TABLE users DROP COLUMN IF EXISTS sms_notifications;
ALTER TABLE users DROP COLUMN IF EXISTS push_notifications;
ALTER TABLE users DROP COLUMN IF EXISTS email_notifications;
ALTER TABLE users DROP COLUMN IF EXISTS profile_visibility;
ALTER TABLE users DROP COLUMN IF EXISTS show_email;
ALTER TABLE users DROP COLUMN IF EXISTS show_location;
ALTER TABLE users DROP COLUMN IF EXISTS allow_messaging;
ALTER TABLE users DROP COLUMN IF EXISTS longitude;
ALTER TABLE users DROP COLUMN IF EXISTS latitude;
ALTER TABLE users DROP COLUMN IF EXISTS country;
ALTER TABLE users DROP COLUMN IF EXISTS state;
ALTER TABLE users DROP COLUMN IF EXISTS city;
ALTER TABLE users DROP COLUMN IF EXISTS profile_picture_url;
ALTER TABLE users DROP COLUMN IF EXISTS bio;
ALTER TABLE users DROP COLUMN IF EXISTS event_reminders;
ALTER TABLE users DROP COLUMN IF EXISTS new_opportunities;
ALTER TABLE users DROP COLUMN IF EXISTS newsletter_subscription;
