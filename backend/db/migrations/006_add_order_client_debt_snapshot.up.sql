-- Add client debt snapshot column to order
ALTER TABLE "order" ADD COLUMN client_debt_snapshot_cents INTEGER;

-- Backfill existing rows with current client debt at time of migration for consistency
UPDATE "order" o
SET client_debt_snapshot_cents = (
  SELECT debt_cents FROM client c WHERE c.id = o.client_id
)
WHERE client_debt_snapshot_cents IS NULL;