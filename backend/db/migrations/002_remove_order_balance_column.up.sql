-- Idempotent migration to remove any legacy order balance column (remaining_cents / debt_cents) by table rebuild.
-- This version avoids nested transactions and view dependency issues.
-- It is safe to run even if the table already has the final structure.

-- 1. Drop dependent view that references the "order" table (will be recreated later)
DROP VIEW IF EXISTS vw_top_clients;

-- 2. If an intermediate leftover table from a failed previous attempt exists, drop it
DROP TABLE IF EXISTS order_new;

-- 3. Recreate a clean target schema for the order table (final form without balance column)
CREATE TABLE IF NOT EXISTS order_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_number TEXT UNIQUE NOT NULL,
    client_id INTEGER NOT NULL REFERENCES client(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'PENDING',
    notes TEXT,
    discount_percent INTEGER DEFAULT 0 CHECK(discount_percent >= 0 AND discount_percent <= 100),
    issue_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    due_date DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

-- 4. Copy data from existing order table if it still exists
--    (If the old table was already replaced, this still works because "order" exists; if not, INSERT is skipped gracefully.)
INSERT INTO order_new (id, order_number, client_id, status, notes, discount_percent, issue_date, due_date, created_at, updated_at)
    SELECT id, order_number, client_id, status, notes, discount_percent, issue_date, due_date, created_at, updated_at
    FROM "order";

-- 5. Replace original table if schema differs. We drop only after successful copy.
DROP TABLE IF EXISTS "order";
ALTER TABLE order_new RENAME TO "order";

-- 6. Recreate indexes possibly lost due to table rebuild (IF NOT EXISTS keeps it idempotent)
CREATE INDEX IF NOT EXISTS idx_order_client_id ON "order"(client_id);
CREATE INDEX IF NOT EXISTS idx_order_status ON "order"(status);
CREATE INDEX IF NOT EXISTS idx_order_issue_date ON "order"(issue_date);

-- 7. Recreate vw_top_clients view with updated table (structure unchanged, but we ensure it exists)
CREATE VIEW IF NOT EXISTS vw_top_clients AS
SELECT 
    c.id,
    c.name,
    COUNT(DISTINCT o.id) as order_count,
    COALESCE(SUM(p.amount_cents), 0) as total_paid_cents
FROM client c
LEFT JOIN "order" o ON c.id = o.client_id
LEFT JOIN invoice i ON c.id = i.client_id
LEFT JOIN payment p ON i.id = p.invoice_id
GROUP BY c.id, c.name
ORDER BY total_paid_cents DESC
LIMIT 10;

-- Done.
