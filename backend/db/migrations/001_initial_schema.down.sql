-- Drop all tables and views in reverse order
DROP VIEW IF EXISTS vw_top_clients;
DROP VIEW IF EXISTS vw_revenue_by_month;
DROP TABLE IF EXISTS payment;
DROP TABLE IF EXISTS invoice_item;
DROP TABLE IF EXISTS invoice;
DROP TABLE IF EXISTS order_item;
DROP TABLE IF EXISTS "order";
DROP TABLE IF EXISTS product;
DROP TABLE IF EXISTS client;
