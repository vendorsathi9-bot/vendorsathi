-- Migration 009: Login attempts table for authentication security
-- Tracks failed login attempts per phone number with lockout support

-- Create login_attempts table
CREATE TABLE login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone TEXT NOT NULL,
    failed_count INT NOT NULL DEFAULT 1,
    failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_until TIMESTAMPTZ,
    succeeded_at TIMESTAMPTZ,
    UNIQUE (phone)
);

-- Index for querying recent failed attempts
CREATE INDEX idx_login_attempts_phone ON login_attempts(phone);
CREATE INDEX idx_login_attempts_failed_at ON login_attempts(failed_at);

-- Comment: This table supports account lockout after 5 failed login attempts
-- within a 15-minute window. After 5 failures, the account is locked for 15 minutes.
-- On successful login, the row failure count is reset.
-- This ensures lockout persists across server restarts and in distributed deployments.
