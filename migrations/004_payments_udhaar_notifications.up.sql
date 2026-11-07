-- Migration 004: Payments, Refunds, Udhaar, Notifications, Device Tokens, Sync Operations, Audit Logs

-- Payments
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    method TEXT NOT NULL CHECK (method IN ('CASH', 'UPI')),
    amount_paise BIGINT NOT NULL CHECK (amount_paise > 0),
    status TEXT NOT NULL DEFAULT 'PENDING_VERIFICATION' CHECK (status IN (
        'PENDING_VERIFICATION', 'CONFIRMED', 'REJECTED', 'REFUNDED_PARTIAL', 'REFUNDED_FULL'
    )),
    verification_status TEXT NOT NULL DEFAULT 'PENDING' CHECK (verification_status IN ('PENDING', 'VERIFIED', 'REJECTED')),
    recorded_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reference TEXT,
    evidence JSONB,
    idempotency_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (order_id, idempotency_key)
);

-- Refunds
CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE RESTRICT,
    amount_paise BIGINT NOT NULL CHECK (amount_paise > 0),
    status TEXT NOT NULL DEFAULT 'RECORDED' CHECK (status IN ('RECORDED', 'PROCESSED', 'FAILED')),
    reason TEXT,
    idempotency_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (payment_id, idempotency_key)
);

-- Udhaar Accounts
CREATE TABLE udhaar_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id UUID NOT NULL REFERENCES vendor_profiles(id) ON DELETE RESTRICT,
    customer_id UUID NOT NULL REFERENCES customer_profiles(user_id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CLOSED', 'SUSPENDED')),
    credit_limit_paise BIGINT CHECK (credit_limit_paise >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (vendor_id, customer_id)
);

-- Udhaar Entries
CREATE TABLE udhaar_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES udhaar_accounts(id) ON DELETE RESTRICT,
    entry_type TEXT NOT NULL CHECK (entry_type IN ('CREDIT', 'REPAYMENT', 'ADJUSTMENT')),
    amount_paise BIGINT NOT NULL CHECK (amount_paise > 0),
    reference_type TEXT,
    reference_id UUID,
    idempotency_key TEXT,
    actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (account_id, idempotency_key)
);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    payload JSONB,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Device Tokens (FCM)
CREATE TABLE device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('ANDROID', 'IOS', 'WEB')),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Sync Operations
CREATE TABLE sync_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id TEXT NOT NULL,
    client_operation_id UUID NOT NULL,
    operation_type TEXT NOT NULL,
    payload_version INT NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENDING', 'ACCEPTED', 'REJECTED', 'NEEDS_REVIEW')),
    payload JSONB NOT NULL,
    result JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    UNIQUE (user_id, client_operation_id)
);

-- Audit Logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id UUID,
    reason TEXT,
    request_id UUID,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_payments_order ON payments(order_id);
CREATE INDEX idx_payments_idempotency ON payments(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_refunds_payment ON refunds(payment_id);
CREATE INDEX idx_refunds_idempotency ON refunds(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_udhaar_accounts_vendor ON udhaar_accounts(vendor_id, status);
CREATE INDEX idx_udhaar_accounts_customer ON udhaar_accounts(customer_id, status);
CREATE INDEX idx_udhaar_entries_account ON udhaar_entries(account_id, created_at DESC);
CREATE INDEX idx_udhaar_entries_idempotency ON udhaar_entries(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_unread ON notifications(user_id, read_at) WHERE read_at IS NULL;
CREATE INDEX idx_device_tokens_user ON device_tokens(user_id);
CREATE INDEX idx_device_tokens_token ON device_tokens(token);
CREATE INDEX idx_sync_operations_user ON sync_operations(user_id, device_id);
CREATE INDEX idx_sync_operations_status ON sync_operations(status) WHERE status IN ('PENDING', 'SENDING');
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_user_id, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);

-- Updated at triggers
CREATE TRIGGER update_udhaar_accounts_updated_at BEFORE UPDATE ON udhaar_accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();