ALTER TABLE order_items DROP COLUMN discount_percent;
ALTER TABLE orders ADD COLUMN tax_percent INTEGER NOT NULL DEFAULT 0;
