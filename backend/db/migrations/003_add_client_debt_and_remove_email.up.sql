-- Atomic migration: rebuild table to remove email and add debt_cents without partial side-effects
PRAGMA foreign_keys=off;
BEGIN TRANSACTION;

-- Drop dependent views to avoid errors while renaming/replacing tables
DROP VIEW IF EXISTS vw_top_clients;
DROP VIEW IF EXISTS vw_revenue_by_month;

-- Create new client table without email and with debt_cents
CREATE TABLE client_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    phone TEXT,
    address TEXT,
    debt_cents INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

-- Copy data from old table; initialize debt_cents to 0
INSERT INTO client_new (id, name, phone, address, debt_cents, created_at, updated_at)
SELECT id, name, phone, address, 0 AS debt_cents, created_at, updated_at FROM client;

-- Replace old table
DROP TABLE client;
ALTER TABLE client_new RENAME TO client;

-- Recreate index
CREATE INDEX IF NOT EXISTS idx_client_name ON client(name);

-- Recreate views (same definitions as initial migration)
CREATE VIEW vw_revenue_by_month AS
SELECT 
    strftime('%Y-%m', paid_at) as month,
    SUM(amount_cents) as revenue_cents
FROM payment
WHERE paid_at >= date('now', '-12 months')
GROUP BY strftime('%Y-%m', paid_at)
ORDER BY month;

CREATE VIEW vw_top_clients AS
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

COMMIT;
PRAGMA foreign_keys=on;
