-- Consolidated, idempotent schema for the application
-- This file represents the final desired schema. It can be run on a fresh DB
-- or an older one; it only creates objects if missing and adds needed columns.

PRAGMA foreign_keys = ON;

-- Tables
CREATE TABLE IF NOT EXISTS client (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    phone TEXT,
    address TEXT,
    debt_cents INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS product (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sku TEXT UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    unit_price_cents INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'DZD',
    active INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS "order" (
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

-- Ensure snapshot column exists on order (added in later versions)
ALTER TABLE "order" ADD COLUMN IF NOT EXISTS client_debt_snapshot_cents INTEGER;

CREATE TABLE IF NOT EXISTS order_item (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL REFERENCES "order"(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES product(id) ON DELETE SET NULL,
    name_snapshot TEXT NOT NULL,
    sku_snapshot TEXT,
    qty INTEGER NOT NULL CHECK(qty > 0),
    unit_price_cents INTEGER NOT NULL CHECK(unit_price_cents >= 0),
    discount_percent INTEGER NOT NULL DEFAULT 0 CHECK(discount_percent >= 0 AND discount_percent <= 100),
    currency TEXT NOT NULL DEFAULT 'DZD',
    total_cents INTEGER NOT NULL CHECK(total_cents >= 0)
);

-- Ensure order_item has discount_percent for older databases
ALTER TABLE order_item
    ADD COLUMN IF NOT EXISTS discount_percent INTEGER NOT NULL DEFAULT 0
        CHECK(discount_percent >= 0 AND discount_percent <= 100);

CREATE TABLE IF NOT EXISTS invoice (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_number TEXT UNIQUE NOT NULL,
    order_id INTEGER REFERENCES "order"(id) ON DELETE SET NULL,
    client_id INTEGER NOT NULL REFERENCES client(id),
    status TEXT NOT NULL DEFAULT 'DRAFT',
    issue_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    due_date DATETIME,
    notes TEXT,
    subtotal_cents INTEGER NOT NULL DEFAULT 0,
    discount_percent INTEGER DEFAULT 0 CHECK(discount_percent >= 0 AND discount_percent <= 100),
    tax_percent INTEGER DEFAULT 0 CHECK(tax_percent >= 0 AND tax_percent <= 100),
    total_cents INTEGER NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'DZD',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS invoice_item (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id INTEGER NOT NULL REFERENCES invoice(id) ON DELETE CASCADE,
    product_id INTEGER REFERENCES product(id) ON DELETE SET NULL,
    name_snapshot TEXT NOT NULL,
    sku_snapshot TEXT,
    qty INTEGER NOT NULL CHECK(qty > 0),
    unit_price_cents INTEGER NOT NULL CHECK(unit_price_cents >= 0),
    currency TEXT NOT NULL DEFAULT 'DZD',
    total_cents INTEGER NOT NULL CHECK(total_cents >= 0)
);

CREATE TABLE IF NOT EXISTS payment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id INTEGER NOT NULL REFERENCES invoice(id) ON DELETE CASCADE,
    amount_cents INTEGER NOT NULL CHECK(amount_cents > 0),
    method TEXT NOT NULL,
    reference TEXT,
    paid_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes (idempotent)
CREATE INDEX IF NOT EXISTS idx_client_name ON client(name);
CREATE INDEX IF NOT EXISTS idx_product_name ON product(name);
CREATE INDEX IF NOT EXISTS idx_product_sku ON product(sku);
CREATE INDEX IF NOT EXISTS idx_product_active ON product(active);
CREATE INDEX IF NOT EXISTS idx_order_client_id ON "order"(client_id);
CREATE INDEX IF NOT EXISTS idx_order_status ON "order"(status);
CREATE INDEX IF NOT EXISTS idx_order_issue_date ON "order"(issue_date);
CREATE INDEX IF NOT EXISTS idx_order_item_order_id ON order_item(order_id);
CREATE INDEX IF NOT EXISTS idx_order_item_product_id ON order_item(product_id);
CREATE INDEX IF NOT EXISTS idx_invoice_client_id ON invoice(client_id);
CREATE INDEX IF NOT EXISTS idx_invoice_order_id ON invoice(order_id);
CREATE INDEX IF NOT EXISTS idx_invoice_status ON invoice(status);
CREATE INDEX IF NOT EXISTS idx_invoice_issue_date ON invoice(issue_date);
CREATE INDEX IF NOT EXISTS idx_invoice_item_invoice_id ON invoice_item(invoice_id);
CREATE INDEX IF NOT EXISTS idx_invoice_item_product_id ON invoice_item(product_id);
CREATE INDEX IF NOT EXISTS idx_payment_invoice_id ON payment(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payment_paid_at ON payment(paid_at);

-- Recreate views safely
DROP VIEW IF EXISTS vw_revenue_by_month;
CREATE VIEW vw_revenue_by_month AS
SELECT 
    strftime('%Y-%m', paid_at) as month,
    SUM(amount_cents) as revenue_cents
FROM payment
WHERE paid_at >= date('now', '-12 months')
GROUP BY strftime('%Y-%m', paid_at)
ORDER BY month;

DROP VIEW IF EXISTS vw_top_clients;
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
