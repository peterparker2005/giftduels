-- Migration: trigger_updated_at
-- Created at: 2025-07-19 14:39:08
-- Description: Add your migration description here

CREATE OR REPLACE FUNCTION set_updated_at_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at := NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_gifts_set_updated_at
BEFORE UPDATE ON gifts
FOR EACH ROW
EXECUTE FUNCTION set_updated_at_timestamp();
