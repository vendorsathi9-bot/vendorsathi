-- Migration 006: Features Expansion
-- Adds cost price to products, preferred language to users, expenses, reviews, and loyalty rewards

-- 1. Add preferred_language to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS preferred_language TEXT NOT NULL DEFAULT 'en';

-- 2. Add cost_price_paise to products
ALTER TABLE products ADD COLUMN IF NOT EXISTS cost_price_paise BIGINT NOT NULL DEFAULT 0 CHECK (cost_price_paise >= 0);

-- 3. Add discount_paise to orders for loyalty redemption
ALTER TABLE orders ADD COLUMN IF NOT EXISTS discount_paise BIGINT NOT NULL DEFAULT 0 CHECK (discount_paise >= 0);

-- 4. Expenses table for vendor net profit calculations
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vendor_id UUID NOT NULL REFERENCES vendor_profiles(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'OPERATIONAL',
    amount_paise BIGINT NOT NULL CHECK (amount_paise > 0),
    expense_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_expenses_vendor_date ON expenses(vendor_id, expense_date DESC);

-- 5. Ratings and reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID UNIQUE REFERENCES orders(id) ON DELETE SET NULL,
    customer_id UUID NOT NULL REFERENCES customer_profiles(user_id) ON DELETE CASCADE,
    vendor_id UUID NOT NULL REFERENCES vendor_profiles(id) ON DELETE CASCADE,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    is_hidden BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reviews_vendor ON reviews(vendor_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_customer ON reviews(customer_id);

-- 6. Loyalty Accounts and Transactions
CREATE TABLE IF NOT EXISTS loyalty_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID UNIQUE NOT NULL REFERENCES customer_profiles(user_id) ON DELETE CASCADE,
    points_balance INT NOT NULL DEFAULT 0 CHECK (points_balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_loyalty_accounts_customer ON loyalty_accounts(customer_id);

CREATE TABLE IF NOT EXISTS loyalty_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES loyalty_accounts(id) ON DELETE CASCADE,
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    points INT NOT NULL,
    transaction_type TEXT NOT NULL CHECK (transaction_type IN ('EARNED', 'REDEEMED', 'ADJUSTMENT')),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (account_id, order_id, transaction_type)
);
CREATE INDEX IF NOT EXISTS idx_loyalty_transactions_account ON loyalty_transactions(account_id, created_at DESC);
