-- Migration 009 down: Remove login attempts table

DROP INDEX IF EXISTS idx_login_attempts_failed_at;
DROP INDEX IF EXISTS idx_login_attempts_phone;
DROP TABLE IF EXISTS login_attempts;