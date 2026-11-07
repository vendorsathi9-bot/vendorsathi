-- Migration 002: Down migration

DROP TRIGGER IF EXISTS update_inventory_updated_at ON inventory;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS update_customer_addresses_updated_at ON customer_addresses;
DROP TRIGGER IF EXISTS update_customer_profiles_updated_at ON customer_profiles;
DROP TRIGGER IF EXISTS update_business_hours_updated_at ON business_hours;
DROP TRIGGER IF EXISTS update_vendor_locations_updated_at ON vendor_locations;
DROP TRIGGER IF EXISTS update_vendor_profiles_updated_at ON vendor_profiles;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS inventory_movements;
DROP TABLE IF EXISTS inventory;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS customer_addresses;
DROP TABLE IF EXISTS customer_profiles;
DROP TABLE IF EXISTS business_hours;
DROP TABLE IF EXISTS vendor_locations;
DROP TABLE IF EXISTS vendor_profiles;