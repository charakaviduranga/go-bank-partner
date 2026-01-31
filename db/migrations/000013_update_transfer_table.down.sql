BEGIN;

-- 1. Revert status check constraint to previous values
ALTER TABLE transfers
  DROP CONSTRAINT IF EXISTS transfers_status_check,
  ADD CONSTRAINT transfers_status_check
    CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED'));

-- 2. Drop added columns
ALTER TABLE transfers
  DROP COLUMN IF EXISTS transfer_id,
  DROP COLUMN IF EXISTS reference_number;

COMMIT;
