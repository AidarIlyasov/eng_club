-- Add type field to events table
ALTER TABLE events 
ADD COLUMN type ENUM('many_activities', 'single_activity') NOT NULL DEFAULT 'many_activities',
ADD INDEX idx_events_type (type);