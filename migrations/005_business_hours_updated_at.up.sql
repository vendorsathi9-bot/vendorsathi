-- Migration 005: Add updated_at column to business_hours to fix trigger
ALTER TABLE business_hours ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
