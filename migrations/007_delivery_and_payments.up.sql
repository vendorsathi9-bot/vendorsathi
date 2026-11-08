-- Add DELIVERY_AGENT role
INSERT INTO roles (code, description) VALUES ('DELIVERY_AGENT', 'Delivery agent who handles order deliveries') ON CONFLICT (code) DO NOTHING;

-- Add delivery_agent_id column to orders table
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_agent_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Delivery assignments table
CREATE TABLE IF NOT EXISTS delivery_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    agent_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'ASSIGNED' CHECK (status IN ('ASSIGNED','ACCEPTED','PICKED_UP','OUT_FOR_DELIVERY','DELIVERED')),
    assigned_by UUID REFERENCES users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMPTZ,
    picked_up_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(order_id)
);
CREATE INDEX IF NOT EXISTS idx_delivery_assignments_agent ON delivery_assignments(agent_user_id, status);

-- Live delivery tracking table (one latest position per active order)
CREATE TABLE IF NOT EXISTS delivery_tracking (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    agent_user_id UUID NOT NULL REFERENCES users(id),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(order_id)
);
