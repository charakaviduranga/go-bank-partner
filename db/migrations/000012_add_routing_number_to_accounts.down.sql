BEGIN;

-- 1. Drop routing_number
ALTER TABLE accounts
  DROP COLUMN routing_number;

-- 2. Shrink account_number (ensure data fits before running)
ALTER TABLE accounts
  ALTER COLUMN account_number TYPE VARCHAR(10);

COMMIT;
