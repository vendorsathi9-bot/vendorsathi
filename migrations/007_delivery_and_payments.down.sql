DROP TABLE IF EXISTS delivery_tracking;
DROP TABLE IF EXISTS delivery_assignments;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_agent_id;
DELETE FROM roles WHERE code = 'DELIVERY_AGENT';
