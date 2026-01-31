BEGIN;

-- 1. Add new columns (nullable first to avoid table rewrite issues)
ALTER TABLE transfers
  ADD COLUMN transfer_id VARCHAR(50),
  ADD COLUMN reference_number VARCHAR(50);

-- 2. Backfill existing rows (adjust logic if needed)
UPDATE transfers
SET
  transfer_id = id::text,
  reference_number = id::text
WHERE transfer_id IS NULL
   OR reference_number IS NULL;

COMMIT;
