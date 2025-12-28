-- Migration: Add is_all_day column to events table
-- This allows events to be marked as all-day events

ALTER TABLE events ADD COLUMN IF NOT EXISTS is_all_day BOOLEAN DEFAULT FALSE;

-- Update existing events to default is_all_day to false
UPDATE events SET is_all_day = FALSE WHERE is_all_day IS NULL;
