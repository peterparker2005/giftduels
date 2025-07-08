-- Drop Tables (in reverse dependency order)
DROP TABLE IF EXISTS gift_events CASCADE;
DROP TABLE IF EXISTS gifts CASCADE;
DROP TABLE IF EXISTS gift_attributes CASCADE;

-- Drop ENUM Types
DROP TYPE IF EXISTS gift_status CASCADE;
DROP TYPE IF EXISTS gift_attribute_type CASCADE;
