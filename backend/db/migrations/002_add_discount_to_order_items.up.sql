ALTER TABLE order_item ADD COLUMN discount_percent INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "order" DROP COLUMN tax_percent;
