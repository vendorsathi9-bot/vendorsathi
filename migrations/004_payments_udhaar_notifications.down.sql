-- Migration 004: Down migration

DROP TRIGGER IF EXISTS update_udhaar_accounts_updated_at ON udhaar_accounts;

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS sync_operations;
DROP TABLE IF EXISTS device_tokens;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS udhaar_entries;
DROP TABLE IF EXISTS udhaar_accounts;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payments;