-- Global consolidated schema (reset)
-- Client table (email removed, debt_cents added)
CREATE TABLE client (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    phone TEXT,
    address TEXT,
    debt_cents INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

-- Create product table
CREATE TABLE product (
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

CREATE TABLE "order" (
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

CREATE TABLE order_item (
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

-- Create invoice table
CREATE TABLE invoice (
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

-- Create invoice_item table
CREATE TABLE invoice_item (
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

-- Create payment table
CREATE TABLE payment (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    invoice_id INTEGER NOT NULL REFERENCES invoice(id) ON DELETE CASCADE,
    amount_cents INTEGER NOT NULL CHECK(amount_cents > 0),
    method TEXT NOT NULL,
    reference TEXT,
    paid_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indices for performance
CREATE INDEX idx_client_name ON client(name);
CREATE INDEX idx_product_name ON product(name);
CREATE INDEX idx_product_sku ON product(sku);
CREATE INDEX idx_product_active ON product(active);
CREATE INDEX idx_order_client_id ON "order"(client_id);
CREATE INDEX idx_order_status ON "order"(status);
CREATE INDEX idx_order_issue_date ON "order"(issue_date);
CREATE INDEX idx_order_item_order_id ON order_item(order_id);
CREATE INDEX idx_order_item_product_id ON order_item(product_id);
CREATE INDEX idx_invoice_client_id ON invoice(client_id);
CREATE INDEX idx_invoice_order_id ON invoice(order_id);
CREATE INDEX idx_invoice_status ON invoice(status);
CREATE INDEX idx_invoice_issue_date ON invoice(issue_date);
CREATE INDEX idx_invoice_item_invoice_id ON invoice_item(invoice_id);
CREATE INDEX idx_invoice_item_product_id ON invoice_item(product_id);
CREATE INDEX idx_payment_invoice_id ON payment(invoice_id);
CREATE INDEX idx_payment_paid_at ON payment(paid_at);

-- Create views for dashboard metrics
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
