-- Attempt to revert client table to include email and remove debt_cents
PRAGMA foreign_keys=off;
BEGIN TRANSACTION;

-- Recreate old client table with email column and without debt_cents
CREATE TABLE client_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    address TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

-- Copy data back (email will be NULL for all rows)
INSERT INTO client_old (id, name, email, phone, address, created_at, updated_at)
SELECT id, name, NULL AS email, phone, address, created_at, updated_at FROM client;

DROP TABLE client;
ALTER TABLE client_old RENAME TO client;

CREATE INDEX idx_client_name ON client(name);

COMMIT;
PRAGMA foreign_keys=on;
