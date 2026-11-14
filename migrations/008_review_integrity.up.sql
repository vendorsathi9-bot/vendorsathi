-- Migration 008: Review integrity and query indexes.
--
-- This migration is additive and non-destructive. It deliberately does not
-- rewrite or delete existing reviews. The application also validates order
-- ownership and completed-order status before inserting a review; these
-- indexes enforce the invariants at the database level and keep vendor review
-- listings fast.

-- One review per order. The original table already declared order_id UNIQUE,
-- but a named partial index makes the intent explicit and covers databases
-- where the inline constraint may have been dropped.
CREATE UNIQUE INDEX IF NOT EXISTS uq_reviews_order_id
    ON reviews(order_id)
    WHERE order_id IS NOT NULL;

-- Fast public vendor review listings (hidden reviews excluded).
CREATE INDEX IF NOT EXISTS idx_reviews_vendor_visible
    ON reviews(vendor_id, created_at DESC)
    WHERE is_hidden = FALSE;

-- Fast admin moderation listing.
CREATE INDEX IF NOT EXISTS idx_reviews_hidden_created
    ON reviews(is_hidden, created_at DESC);
