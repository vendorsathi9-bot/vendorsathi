-- Migration 005 down: Remove updated_at column from business_hours
ALTER TABLE business_hours DROP COLUMN IF EXISTS updated_at;
