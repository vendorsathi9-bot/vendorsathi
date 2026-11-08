DROP TABLE IF EXISTS loyalty_transactions;
DROP TABLE IF EXISTS loyalty_accounts;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS expenses;
ALTER TABLE orders DROP COLUMN IF EXISTS discount_paise;
ALTER TABLE products DROP COLUMN IF EXISTS cost_price_paise;
ALTER TABLE users DROP COLUMN IF EXISTS preferred_language;
