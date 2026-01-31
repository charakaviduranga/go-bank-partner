BEGIN;

-- 1. Widen account_number (fast, metadata-only)
ALTER TABLE accounts
  ALTER COLUMN account_number TYPE VARCHAR(50);

-- 2. Add routing_number safely for existing rows
ALTER TABLE accounts
  ADD COLUMN routing_number VARCHAR(20);

-- 3. Backfill existing records
UPDATE accounts
  SET routing_number = '123456789'
  WHERE routing_number IS NULL;

COMMIT;
