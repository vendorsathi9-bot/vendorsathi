-- Migration 001: Initial schema for VendorSaathi
-- Creates core tables: roles, permissions, users, user_roles, refresh_sessions

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Permissions table
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Role-Permission mapping
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACTIVE', 'SUSPENDED', 'CLOSED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- User-Role mapping
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Refresh Sessions table
CREATE TABLE refresh_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    device_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_refresh_sessions_user_id ON refresh_sessions(user_id);
CREATE INDEX idx_refresh_sessions_token_hash ON refresh_sessions(token_hash);
CREATE INDEX idx_refresh_sessions_expires_at ON refresh_sessions(expires_at) WHERE revoked_at IS NULL;

-- Seed initial roles
INSERT INTO roles (code, description) VALUES
    ('CUSTOMER', 'Customer user'),
    ('VENDOR', 'Vendor user'),
    ('ADMIN', 'Administrator'),
    ('SUPER_ADMIN', 'Super administrator')
ON CONFLICT (code) DO NOTHING;

-- Seed basic permissions
INSERT INTO permissions (code, description) VALUES
    ('vendor:read', 'Read vendor information'),
    ('vendor:write', 'Create/update vendor profile'),
    ('vendor:manage', 'Full vendor management (admin)'),
    ('product:read', 'Read products'),
    ('product:write', 'Create/update products'),
    ('product:delete', 'Archive products'),
    ('inventory:read', 'Read inventory'),
    ('inventory:write', 'Adjust inventory'),
    ('order:read', 'Read orders'),
    ('order:write', 'Create orders (customer)'),
    ('order:manage', 'Accept/reject/fulfil orders (vendor)'),
    ('order:admin', 'Full order oversight (admin)'),
    ('payment:read', 'Read payments'),
    ('payment:write', 'Record payments'),
    ('payment:verify', 'Verify UPI payments'),
    ('payment:refund', 'Process refunds'),
    ('udhaar:read', 'Read udhaar accounts'),
    ('udhaar:write', 'Manage udhaar entries'),
    ('notification:read', 'Read notifications'),
    ('admin:users', 'Manage users'),
    ('admin:vendors', 'Manage vendors'),
    ('admin:audit', 'View audit logs'),
    ('admin:permissions', 'Manage admin permissions')
ON CONFLICT (code) DO NOTHING;

-- Assign permissions to roles
-- CUSTOMER permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'CUSTOMER' AND p.code IN (
    'vendor:read', 'product:read', 'order:write', 'order:read',
    'payment:read', 'notification:read'
)
ON CONFLICT DO NOTHING;

-- VENDOR permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'VENDOR' AND p.code IN (
    'vendor:read', 'vendor:write',
    'product:read', 'product:write', 'product:delete',
    'inventory:read', 'inventory:write',
    'order:read', 'order:manage',
    'payment:read', 'payment:write', 'payment:verify', 'payment:refund',
    'udhaar:read', 'udhaar:write',
    'notification:read'
)
ON CONFLICT DO NOTHING;

-- ADMIN permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'ADMIN' AND p.code IN (
    'vendor:read', 'vendor:manage',
    'product:read', 'order:read', 'order:admin',
    'payment:read', 'payment:refund',
    'udhaar:read',
    'notification:read',
    'admin:users', 'admin:vendors', 'admin:audit'
)
ON CONFLICT DO NOTHING;

-- SUPER_ADMIN permissions (all)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code = 'SUPER_ADMIN'
ON CONFLICT DO NOTHING;