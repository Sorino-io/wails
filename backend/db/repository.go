package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

// Repository handles all database operations
type Repository struct {
	db *DB
}

// NewRepository creates a new repository instance
func NewRepository(db *DB) *Repository { return &Repository{db: db} }

// Client operations

func (r *Repository) CreateClient(ctx context.Context, client Client) (*Client, error) {
	debug := false
	if v, ok := os.LookupEnv("DEBUG_CLIENTS"); ok && v != "" && v != "0" {
		debug = true
	}
	if debug {
		log.Printf("[clients] repo CreateClient inserting name=%q phone=%v address=%v debt=%d", client.Name, client.Phone, client.Address, client.DebtCents)
	}
	query := `
		INSERT INTO client (name, phone, address, debt_cents, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	result, err := r.db.ExecContext(ctx, query, client.Name, client.Phone, client.Address, client.DebtCents)
	if err != nil {
		if debug { log.Printf("[clients] repo CreateClient SQL error: %v", err) }
		log.Printf("[clients] repo CreateClient FAILED name=%q err=%v", client.Name, err)
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get client ID: %w", err)
	}

	client.ID = id
	client.CreatedAt = time.Now()
	if debug {
		log.Printf("[clients] repo CreateClient success id=%d", client.ID)
	}
	return &client, nil
}

func (r *Repository) GetClient(ctx context.Context, id int64) (*Client, error) {
	query := `SELECT id, name, phone, address, debt_cents, created_at, updated_at FROM client WHERE id = ?`

	var client Client
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&client.ID, &client.Name, &client.Phone, &client.Address, &client.DebtCents, &client.CreatedAt, &client.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("client not found")
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &client, nil
}

func (r *Repository) ListClients(ctx context.Context, query string, limit, offset int) ([]Client, int, error) {
	var clients []Client
	var total int

	// Count total
	countQuery := `SELECT COUNT(*) FROM client WHERE name LIKE ?`
	searchPattern := "%" + query + "%"
	err := r.db.QueryRowContext(ctx, countQuery, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count clients: %w", err)
	}

	// Get clients
	listQuery := `
		SELECT id, name, phone, address, debt_cents, created_at, updated_at 
		FROM client 
		WHERE name LIKE ? 
		ORDER BY name 
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, listQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clients: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var client Client
		err := rows.Scan(&client.ID, &client.Name, &client.Phone, &client.Address, &client.DebtCents, &client.CreatedAt, &client.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan client: %w", err)
		}
		clients = append(clients, client)
	}

	return clients, total, nil
}

func (r *Repository) UpdateClient(ctx context.Context, client Client) (*Client, error) {
	query := `
		UPDATE client 
		SET name = ?, phone = ?, address = ?, debt_cents = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, client.Name, client.Phone, client.Address, client.DebtCents, client.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}
	return r.GetClient(ctx, client.ID)
}

// DeleteClient attempts to delete a client. It will fail if FK constraints (orders/invoices) reference it.
func (r *Repository) DeleteClient(ctx context.Context, id int64) error {
	// Quick existence check
	if _, err := r.GetClient(ctx, id); err != nil { return err }
	// Attempt delete (will error if referenced due to foreign keys)
	_, err := r.db.ExecContext(ctx, `DELETE FROM client WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}
	return nil
}

// Product operations

func (r *Repository) CreateProduct(ctx context.Context, product Product) (*Product, error) {
	query := `
		INSERT INTO product (sku, name, description, unit_price_cents, currency, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	result, err := r.db.ExecContext(ctx, query, product.SKU, product.Name, product.Description,
		product.UnitPriceCents, product.Currency, product.Active)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get product ID: %w", err)
	}

	product.ID = id
	product.CreatedAt = time.Now()
	return &product, nil
}

func (r *Repository) GetProduct(ctx context.Context, id int64) (*Product, error) {
	query := `SELECT id, sku, name, description, unit_price_cents, currency, active, created_at, updated_at FROM product WHERE id = ?`

	var product Product
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&product.ID, &product.SKU, &product.Name, &product.Description,
		&product.UnitPriceCents, &product.Currency, &product.Active, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *Repository) ListProducts(ctx context.Context, query string, active *bool, limit, offset int) ([]Product, int, error) {
	var products []Product
	var total int

	// Build WHERE clause
	whereClause := "WHERE (name LIKE ? OR sku LIKE ?)"
	args := []interface{}{"%" + query + "%", "%" + query + "%"}

	if active != nil {
		whereClause += " AND active = ?"
		args = append(args, *active)
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM product %s", whereClause)
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Get products
	listQuery := fmt.Sprintf(`
		SELECT id, sku, name, description, unit_price_cents, currency, active, created_at, updated_at 
		FROM product 
		%s 
		ORDER BY name 
		LIMIT ? OFFSET ?
	`, whereClause)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.Description,
			&product.UnitPriceCents, &product.Currency, &product.Active, &product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	return products, total, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, product Product) (*Product, error) {
	query := `
		UPDATE product 
		SET sku = ?, name = ?, description = ?, unit_price_cents = ?, currency = ?, active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, product.SKU, product.Name, product.Description, product.UnitPriceCents, product.Currency, product.Active, product.ID)
	if err != nil { return nil, fmt.Errorf("failed to update product: %w", err) }

	return r.GetProduct(ctx, product.ID)
}

// DeleteProduct deletes a product (will not delete existing order/invoice snapshots since they store name/sku snapshots)
func (r *Repository) DeleteProduct(ctx context.Context, id int64) error {
	if _, err := r.GetProduct(ctx, id); err != nil { return err }
	_, err := r.db.ExecContext(ctx, `DELETE FROM product WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// Order operations (simplified - full implementation would include items handling)

func (r *Repository) CreateOrder(ctx context.Context, draft OrderDraft) (*Order, error) {
	// TEMP DIAGNOSTIC LOGGING
	fmt.Printf("[CreateOrder] START client_id=%d items=%d discount=%d\n", draft.ClientID, len(draft.Items), draft.DiscountPercent)
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Printf("[CreateOrder] begin tx error: %v\n", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Generate order number
	orderNumber, err := r.generateOrderNumber(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate order number: %w", err)
	}

	// Set default issue date if not provided
	issueDate := time.Now()
	if draft.IssueDate != nil {
		issueDate = *draft.IssueDate
	}

	// Get client's current debt
	var clientDebtCents int64
	err = tx.QueryRowContext(ctx, `SELECT debt_cents FROM client WHERE id = ?`, draft.ClientID).Scan(&clientDebtCents)
	if err != nil {
		fmt.Printf("[CreateOrder] select client debt error: %v\n", err)
		return nil, fmt.Errorf("failed to get client debt: %w", err)
	}

	// Create order (snapshot column initially NULL, will fill after possible debt update)
	orderQuery := `
		INSERT INTO "order" (order_number, client_id, status, notes, discount_percent, issue_date, due_date, client_debt_snapshot_cents, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	result, err := tx.ExecContext(ctx, orderQuery, orderNumber, draft.ClientID, OrderStatusPending,
		draft.Notes, draft.DiscountPercent, issueDate, draft.DueDate)
	if err != nil {
		fmt.Printf("[CreateOrder] insert order error: %v\n", err)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get order ID: %w", err)
	}

	var orderTotalCents int64
	// Create order items
	for idx, item := range draft.Items {
		totalCents := int64(item.Qty) * item.UnitPriceCents
		orderTotalCents += totalCents - (totalCents*int64(item.DiscountPercent))/100
		itemQuery := `
			INSERT INTO order_item (order_id, product_id, name_snapshot, sku_snapshot, qty, unit_price_cents, discount_percent, currency, total_cents)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := tx.ExecContext(ctx, itemQuery, orderID, item.ProductID, item.NameSnapshot,
			item.SKUSnapshot, item.Qty, item.UnitPriceCents, item.DiscountPercent, item.Currency, totalCents)
		if err != nil {
			fmt.Printf("[CreateOrder] insert order item error: %v (idx=%d)\n", err, idx)
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	// Global discount is NOT applied to orderTotalCents - it's just a UI helper for setting item discounts
	// The orderTotalCents already includes item-level discounts applied above

	// Increment client's debt by order total (business rule retained)
	if orderTotalCents > 0 {
		if _, err := tx.ExecContext(ctx, `UPDATE client SET debt_cents = COALESCE(debt_cents,0) + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, orderTotalCents, draft.ClientID); err != nil {
			fmt.Printf("[CreateOrder] update client debt error: %v\n", err)
			return nil, fmt.Errorf("failed to update client debt: %w", err)
		}
	}

	// Store snapshot after increment (or unchanged if zero total) so PDF has consistent historical value
	if _, err := tx.ExecContext(ctx, `UPDATE "order" SET client_debt_snapshot_cents = (SELECT debt_cents FROM client WHERE id = ?) WHERE id = ?`, draft.ClientID, orderID); err != nil {
		return nil, fmt.Errorf("failed to set debt snapshot: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
			fmt.Printf("[CreateOrder] commit error: %v\n", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return created order
	// Load snapshot for return
	var snapshot *int64
	_ = r.db.QueryRowContext(ctx, `SELECT client_debt_snapshot_cents FROM "order" WHERE id = ?`, orderID).Scan(&snapshot)
	order := &Order{
		ID:                       orderID,
		OrderNumber:              orderNumber,
		ClientID:                 draft.ClientID,
		Status:                   OrderStatusPending,
		Notes:                    draft.Notes,
		DiscountPercent:          draft.DiscountPercent,
		IssueDate:                issueDate,
		DueDate:                  draft.DueDate,
		CreatedAt:                time.Now(),
		ClientDebtSnapshotCents:  snapshot,
	}

	fmt.Printf("[CreateOrder] SUCCESS order_id=%d total=%d items=%d\n", order.ID, orderTotalCents, len(draft.Items))
	return order, nil
}

// CancelOrderAndAdjustDebt sets order status to CANCELED and subtracts its total from client's debt if no invoices/payments exist.
// It returns the amount subtracted (could be 0 if invoices/payments prevent adjustment).
func (r *Repository) CancelOrderAndAdjustDebt(ctx context.Context, orderID int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return 0, fmt.Errorf("begin tx: %w", err) }
	defer tx.Rollback()

	// Load order + items and client id
	var status string
	var clientID int64
	err = tx.QueryRowContext(ctx, `SELECT status, client_id FROM "order" WHERE id = ?`, orderID).Scan(&status, &clientID)
	if err != nil {
		if err == sql.ErrNoRows { return 0, fmt.Errorf("order not found") }
		return 0, fmt.Errorf("load order: %w", err)
	}
	if status == OrderStatusCanceled { // already canceled
		return 0, nil
	}

	// Compute order total (same logic used in listing)
	rows, err := tx.QueryContext(ctx, `SELECT qty, unit_price_cents, discount_percent FROM order_item WHERE order_id = ?`, orderID)
	if err != nil { return 0, fmt.Errorf("query items: %w", err) }
	defer rows.Close()
	var total int64
	for rows.Next() {
		var qty int32; var price int64; var disc int32
		if err := rows.Scan(&qty, &price, &disc); err != nil { return 0, fmt.Errorf("scan item: %w", err) }
		line := int64(qty) * price
		line = line - (line*int64(disc))/100
		total += line
	}

	// Order-level discount is NOT applied - it's just a UI helper for setting item discounts
	// Total already includes item-level discounts from the loop above

	// Check for invoices/payments referencing this order
	var invoiceCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoice WHERE order_id = ?`, orderID).Scan(&invoiceCount); err != nil {
		return 0, fmt.Errorf("count invoices: %w", err)
	}
	var paymentCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(p.id) FROM payment p JOIN invoice i ON p.invoice_id = i.id WHERE i.order_id = ?`, orderID).Scan(&paymentCount); err != nil {
		return 0, fmt.Errorf("count payments: %w", err)
	}

	// Update order status
	if _, err := tx.ExecContext(ctx, `UPDATE "order" SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, OrderStatusCanceled, orderID); err != nil {
		return 0, fmt.Errorf("update order: %w", err)
	}

	var adjusted int64
	if invoiceCount == 0 && paymentCount == 0 && total > 0 { // Safe to roll back debt
		if _, err := tx.ExecContext(ctx, `UPDATE client SET debt_cents = CASE WHEN debt_cents - ? < 0 THEN 0 ELSE debt_cents - ? END, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, total, total, clientID); err != nil {
			return 0, fmt.Errorf("adjust debt: %w", err)
		}
		adjusted = total
	}

	if err := tx.Commit(); err != nil { return 0, fmt.Errorf("commit: %w", err) }
	return adjusted, nil
}

// DeleteCanceledOrdersForClient removes all canceled orders (and their items) for a client. Returns count removed.
func (r *Repository) DeleteCanceledOrdersForClient(ctx context.Context, clientID int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil { return 0, fmt.Errorf("begin tx: %w", err) }
	defer tx.Rollback()

	// Collect canceled order ids
	rows, err := tx.QueryContext(ctx, `SELECT id FROM "order" WHERE client_id = ? AND status = ?`, clientID, OrderStatusCanceled)
	if err != nil { return 0, fmt.Errorf("query canceled orders: %w", err) }
	defer rows.Close()
	var ids []int64
	for rows.Next() { var oid int64; if err := rows.Scan(&oid); err != nil { return 0, err }; ids = append(ids, oid) }
	if len(ids) == 0 { return 0, nil }

	// Delete items then orders
	// Use IN clause. For large sets we could chunk, but expected small.
	placeholders := ""
	for i := range ids { if i>0 { placeholders += "," }; placeholders += "?" }
	args := make([]interface{}, len(ids))
	for i,v := range ids { args[i] = v }
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DELETE FROM order_item WHERE order_id IN (%s)`, placeholders), args...); err != nil {
		return 0, fmt.Errorf("delete items: %w", err)
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`DELETE FROM "order" WHERE id IN (%s)`, placeholders), args...); err != nil {
		return 0, fmt.Errorf("delete orders: %w", err)
	}
	if err := tx.Commit(); err != nil { return 0, fmt.Errorf("commit: %w", err) }
	return int64(len(ids)), nil
}

// ProductOrderUsageStats returns counts of total and active (non-canceled) orders referencing a product
func (r *Repository) ProductOrderUsageStats(ctx context.Context, productID int64) (total int64, active int64, err error) {
	queryTotal := `SELECT COUNT(*) FROM order_item WHERE product_id = ?`
	if err = r.db.QueryRowContext(ctx, queryTotal, productID).Scan(&total); err != nil { return }
	queryActive := `SELECT COUNT(DISTINCT o.id) FROM order_item oi JOIN "order" o ON oi.order_id = o.id WHERE oi.product_id = ? AND o.status != ?`
	if err = r.db.QueryRowContext(ctx, queryActive, productID, OrderStatusCanceled).Scan(&active); err != nil { return }
	return
}

// HasActiveOrdersForClient returns true if client has any non-canceled orders
func (r *Repository) HasActiveOrdersForClient(ctx context.Context, clientID int64) (bool, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM "order" WHERE client_id = ? AND status != ?`, clientID, OrderStatusCanceled).Scan(&count)
	if err != nil { return false, fmt.Errorf("check active orders: %w", err) }
	return count > 0, nil
}

// DebugSchema returns a map of table -> columns for diagnostics
func (r *Repository) DebugSchema(ctx context.Context) (map[string][]string, error) {
	result := make(map[string][]string)
	rows, err := r.db.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil { return nil, err }
	defer rows.Close()
	var tables []string
	for rows.Next() { var name string; if err := rows.Scan(&name); err != nil { return nil, err }; tables = append(tables, name) }
	for _, t := range tables {
		crow, err := r.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, t))
		if err != nil { continue }
		defer crow.Close()
		cols := []string{}
		for crow.Next() { var cid int; var name, ctype string; var notnull, pk int; var dflt interface{}; if err := crow.Scan(&cid,&name,&ctype,&notnull,&dflt,&pk); err==nil { cols = append(cols, name) } }
		result[t] = cols
	}
	return result, nil
}

// ListOrders retrieves orders with pagination and filters
func (r *Repository) ListOrders(ctx context.Context, filters OrderFilters, limit, offset int) ([]OrderDetail, int, error) {
	// Build WHERE clause
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if filters.ClientID != nil {
		whereClause += " AND o.client_id = ?"
		args = append(args, *filters.ClientID)
	}

	if filters.Status != nil {
		whereClause += " AND o.status = ?"
		args = append(args, *filters.Status)
	}

	if filters.Query != nil && *filters.Query != "" {
		whereClause += " AND (o.order_number LIKE ? OR c.name LIKE ?)"
		queryPattern := "%" + *filters.Query + "%"
		args = append(args, queryPattern, queryPattern)
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM "order" o 
		JOIN client c ON o.client_id = c.id 
		%s
	`, whereClause)

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get orders count: %w", err)
	}

	// Determine sort order
	sortClause := "ORDER BY o.id DESC"
	if filters.Sort != nil && *filters.Sort != "" {
		sortClause = fmt.Sprintf("ORDER BY %s", *filters.Sort)
	}

	// Get orders with details
	query := fmt.Sprintf(`
			SELECT 
				o.id, o.order_number, o.client_id, o.status, o.notes, 
				o.discount_percent, o.issue_date, o.due_date, o.client_debt_snapshot_cents,
				o.created_at, o.updated_at,
				c.id, c.name, c.phone, c.address, c.debt_cents, c.created_at, c.updated_at
			FROM "order" o
			JOIN client c ON o.client_id = c.id
			%s
			%s
			LIMIT ? OFFSET ?
		`, whereClause, sortClause)

	// Add limit and offset to args
	queryArgs := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []OrderDetail
	for rows.Next() {
		var order OrderDetail
		err := rows.Scan(
			&order.Order.ID, &order.Order.OrderNumber, &order.Order.ClientID, &order.Order.Status,
			&order.Order.Notes, &order.Order.DiscountPercent,
			&order.Order.IssueDate, &order.Order.DueDate, &order.Order.ClientDebtSnapshotCents, &order.Order.CreatedAt, &order.Order.UpdatedAt,
			&order.Client.ID, &order.Client.Name, &order.Client.Phone,
			&order.Client.Address, &order.Client.DebtCents, &order.Client.CreatedAt, &order.Client.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}

		// Get order items
		items, err := r.getOrderItems(ctx, order.Order.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get order items: %w", err)
		}
		order.Items = items

		// Calculate totals - pass 0 for discount since global discount is just UI helper
		subtotal, discount, _, total := CalcOrderTotals(items, 0, 0)
		order.SubtotalCents = subtotal
		order.DiscountCents = discount
		order.TaxCents = 0
		order.TotalCents = total

		orders = append(orders, order)
	}

	return orders, total, nil
}

// GetOrderDetail retrieves a single order with full details
func (r *Repository) GetOrderDetail(ctx context.Context, id int64) (*OrderDetail, error) {
	query := fmt.Sprintf(`
		SELECT 
			o.id, o.order_number, o.client_id, o.status, o.notes, 
			o.discount_percent, o.issue_date, o.due_date, o.client_debt_snapshot_cents,
			o.created_at, o.updated_at,
			c.id, c.name, c.phone, c.address, c.debt_cents, c.created_at, c.updated_at
		FROM "order" o
		JOIN client c ON o.client_id = c.id
		WHERE o.id = ?
	`)

	var order OrderDetail
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.Order.ID, &order.Order.OrderNumber, &order.Order.ClientID, &order.Order.Status,
		&order.Order.Notes, &order.Order.DiscountPercent,
		&order.Order.IssueDate, &order.Order.DueDate, &order.Order.ClientDebtSnapshotCents, &order.Order.CreatedAt, &order.Order.UpdatedAt,
		&order.Client.ID, &order.Client.Name, &order.Client.Phone,
		&order.Client.Address, &order.Client.DebtCents, &order.Client.CreatedAt, &order.Client.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get order items
	items, err := r.getOrderItems(ctx, order.Order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	order.Items = items

	// Calculate totals - pass 0 for discount since global discount is just UI helper
	subtotal, discount, _, total := CalcOrderTotals(items, 0, 0)
	order.SubtotalCents = subtotal
	order.DiscountCents = discount
	order.TaxCents = 0
	order.TotalCents = total

	return &order, nil
}

// CreateInvoice creates a new invoice and its items
func (r *Repository) CreateInvoice(ctx context.Context, draft InvoiceDraft) (*Invoice, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	invoiceNumber, err := r.generateInvoiceNumber(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	issueDate := time.Now()
	if draft.IssueDate != nil {
		issueDate = *draft.IssueDate
	}

	query := `
		INSERT INTO invoice (invoice_number, order_id, client_id, status, issue_date, due_date, notes, subtotal_cents, discount_percent, tax_percent, total_cents, currency, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	// For now, subtotal and total will be 0; caller may update later
	result, err := tx.ExecContext(ctx, query, invoiceNumber, draft.OrderID, draft.ClientID, "DRAFT", issueDate, draft.DueDate, draft.Notes, 0, draft.DiscountPercent, draft.TaxPercent, 0, draft.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	invoiceID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice ID: %w", err)
	}

	// Insert items
	for _, item := range draft.Items {
		totalCents := int64(item.Qty) * item.UnitPriceCents
		itemQuery := `
			INSERT INTO invoice_item (invoice_id, product_id, name_snapshot, sku_snapshot, qty, unit_price_cents, currency, total_cents)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := tx.ExecContext(ctx, itemQuery, invoiceID, item.ProductID, item.NameSnapshot, item.SKUSnapshot, item.Qty, item.UnitPriceCents, item.Currency, totalCents)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	inv := &Invoice{
		ID:              invoiceID,
		InvoiceNumber:   invoiceNumber,
		OrderID:         draft.OrderID,
		ClientID:        draft.ClientID,
		Status:          "DRAFT",
		IssueDate:       issueDate,
		DueDate:         draft.DueDate,
		Notes:           draft.Notes,
		SubtotalCents:   0,
		DiscountPercent: draft.DiscountPercent,
		TaxPercent:      draft.TaxPercent,
		TotalCents:      0,
		Currency:        draft.Currency,
		CreatedAt:       time.Now(),
	}

	return inv, nil
}

// ListInvoices retrieves invoices with simple pagination
func (r *Repository) ListInvoices(ctx context.Context, limit, offset int) ([]InvoiceDetail, int, error) {
	var invoices []InvoiceDetail
	var total int

	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoice`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	query := `
		SELECT i.id, i.invoice_number, i.order_id, i.client_id, i.status, i.issue_date, i.due_date, i.notes, i.subtotal_cents, i.discount_percent, i.tax_percent, i.total_cents, i.currency, i.created_at, i.updated_at,
			   c.id, c.name, c.phone, c.address, c.debt_cents, c.created_at, c.updated_at
		FROM invoice i
		JOIN client c ON i.client_id = c.id
		ORDER BY i.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var inv Invoice
		var client Client
		// scan invoice fields then client fields
		err := rows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.OrderID, &inv.ClientID, &inv.Status, &inv.IssueDate, &inv.DueDate, &inv.Notes, &inv.SubtotalCents, &inv.DiscountPercent, &inv.TaxPercent, &inv.TotalCents, &inv.Currency, &inv.CreatedAt, &inv.UpdatedAt,
			&client.ID, &client.Name, &client.Phone, &client.Address, &client.DebtCents, &client.CreatedAt, &client.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan invoice with client: %w", err)
		}

		detail := InvoiceDetail{
			Invoice:      inv,
			Client:       client,
			Items:        []InvoiceItem{},
			Payments:     []Payment{},
			PaidCents:    0,
			BalanceCents: inv.TotalCents,
		}
		invoices = append(invoices, detail)
	}

	return invoices, total, nil
}

// GetInvoiceDetail retrieves full invoice with items and payments
func (r *Repository) GetInvoiceDetail(ctx context.Context, id int64) (*InvoiceDetail, error) {
	query := `SELECT id, invoice_number, order_id, client_id, status, issue_date, due_date, notes, subtotal_cents, discount_percent, tax_percent, total_cents, currency, created_at, updated_at FROM invoice WHERE id = ?`
	var inv Invoice
	err := r.db.QueryRowContext(ctx, query, id).Scan(&inv.ID, &inv.InvoiceNumber, &inv.OrderID, &inv.ClientID, &inv.Status, &inv.IssueDate, &inv.DueDate, &inv.Notes, &inv.SubtotalCents, &inv.DiscountPercent, &inv.TaxPercent, &inv.TotalCents, &inv.Currency, &inv.CreatedAt, &inv.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Get client
	client, err := r.GetClient(ctx, inv.ClientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Get items
	rows, err := r.db.QueryContext(ctx, `SELECT id, invoice_id, product_id, name_snapshot, sku_snapshot, qty, unit_price_cents, currency, total_cents FROM invoice_item WHERE invoice_id = ? ORDER BY id`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice items: %w", err)
	}
	defer rows.Close()

	var items []InvoiceItem
	for rows.Next() {
		var it InvoiceItem
		if err := rows.Scan(&it.ID, &it.InvoiceID, &it.ProductID, &it.NameSnapshot, &it.SKUSnapshot, &it.Qty, &it.UnitPriceCents, &it.Currency, &it.TotalCents); err != nil {
			return nil, fmt.Errorf("failed to scan invoice item: %w", err)
		}
		items = append(items, it)
	}

	// Get payments
	payRows, err := r.db.QueryContext(ctx, `SELECT id, invoice_id, amount_cents, method, reference, paid_at, notes, created_at FROM payment WHERE invoice_id = ? ORDER BY paid_at`, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments: %w", err)
	}
	defer payRows.Close()

	var payments []Payment
	var paidCents int64
	for payRows.Next() {
		var p Payment
		if err := payRows.Scan(&p.ID, &p.InvoiceID, &p.AmountCents, &p.Method, &p.Reference, &p.PaidAt, &p.Notes, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, p)
		paidCents += p.AmountCents
	}

	detail := &InvoiceDetail{
		Invoice:      inv,
		Client:       *client,
		Items:        items,
		Payments:     payments,
		PaidCents:    paidCents,
		BalanceCents: inv.TotalCents - paidCents,
	}

	return detail, nil
}

// UpdateOrder updates an existing order
func (r *Repository) UpdateOrder(ctx context.Context, update OrderUpdate) (*Order, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Capture existing order state for debt diff
	var existingStatus string
	var clientID int64
	if err := tx.QueryRowContext(ctx, `SELECT status, client_id FROM "order" WHERE id = ?`, update.ID).Scan(&existingStatus, &clientID); err != nil {
		if err == sql.ErrNoRows { return nil, fmt.Errorf("order not found") }
		return nil, fmt.Errorf("failed to load existing order: %w", err)
	}

	// Compute current total before changes (only if we might need debt adjustment)
	var oldTotal int64
	{
		rows, err := tx.QueryContext(ctx, `SELECT qty, unit_price_cents, discount_percent FROM order_item WHERE order_id = ?`, update.ID)
		if err != nil { return nil, fmt.Errorf("failed to load existing items: %w", err) }
		for rows.Next() {
			var qty int32; var price int64; var disc int32
			if err := rows.Scan(&qty, &price, &disc); err != nil { rows.Close(); return nil, fmt.Errorf("scan existing item: %w", err) }
			line := int64(qty) * price
			line = line - (line*int64(disc))/100
			oldTotal += line
		}
		rows.Close()
		// Order-level discount is NOT applied - it's just a UI helper for setting item discounts
		// oldTotal already includes item-level discounts from the loop above
	}

	// Build UPDATE query dynamically
	setParts := []string{}
	args := []interface{}{}

	if update.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *update.Status)
	}

	if update.Notes != nil {
		setParts = append(setParts, "notes = ?")
		args = append(args, *update.Notes)
	}

	if update.DiscountPercent != nil {
		setParts = append(setParts, "discount_percent = ?")
		args = append(args, *update.DiscountPercent)
	}

	if update.DueDate != nil {
		setParts = append(setParts, "due_date = ?")
		args = append(args, *update.DueDate)
	}

	if len(setParts) > 0 {
		setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
		args = append(args, update.ID)

		setClause := setParts[0]
		for i := 1; i < len(setParts); i++ {
			setClause += ", " + setParts[i]
		}

		query := fmt.Sprintf(`UPDATE "order" SET %s WHERE id = ?`, setClause)

		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to update order: %w", err)
		}
	}

	// Update items if provided
	var orderTotalCents int64
	if len(update.Items) > 0 {
		// Delete existing items
		_, err = tx.ExecContext(ctx, `DELETE FROM order_item WHERE order_id = ?`, update.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete existing items: %w", err)
		}

		// Insert new items
		for _, item := range update.Items {
			totalCents := int64(item.Qty) * item.UnitPriceCents
			orderTotalCents += totalCents - (totalCents*int64(item.DiscountPercent))/100
			itemQuery := `
				INSERT INTO order_item (order_id, product_id, name_snapshot, sku_snapshot, qty, unit_price_cents, discount_percent, currency, total_cents)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`
			_, err := tx.ExecContext(ctx, itemQuery, update.ID, item.ProductID, item.NameSnapshot,
				item.SKUSnapshot, item.Qty, item.UnitPriceCents, item.DiscountPercent, item.Currency, totalCents)
			if err != nil {
				return nil, fmt.Errorf("failed to create order item: %w", err)
			}
		}
	}

	// Order-level discount is NOT applied to orderTotalCents - it's just a UI helper
	// The orderTotalCents already includes item-level discounts applied above

	// Adjust client debt if order influences debt (not canceled) and no invoices/payments, and data actually changed
	/*
	if existingStatus != OrderStatusCanceled {
		var invoiceCount, paymentCount int64
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM invoice WHERE order_id = ?`, update.ID).Scan(&invoiceCount); err != nil {
			return nil, fmt.Errorf("count invoices: %w", err)
		}
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(p.id) FROM payment p JOIN invoice i ON p.invoice_id = i.id WHERE i.order_id = ?`, update.ID).Scan(&paymentCount); err != nil {
			return nil, fmt.Errorf("count payments: %w", err)
		}
		if invoiceCount == 0 && paymentCount == 0 && (len(update.Items) > 0 || update.DiscountPercent != nil) {
			var newTotal int64 = oldTotal
			if len(update.Items) > 0 {
				newTotal = orderTotalCents
			} else if update.DiscountPercent != nil {
				var base int64
				rows, err := tx.QueryContext(ctx, `SELECT qty, unit_price_cents, discount_percent FROM order_item WHERE order_id = ?`, update.ID)
				if err != nil { return nil, fmt.Errorf("recalc items for discount: %w", err) }
				for rows.Next() {
					var qty int32; var price int64; var disc int32
					if err := rows.Scan(&qty, &price, &disc); err != nil { rows.Close(); return nil, fmt.Errorf("scan recalc item: %w", err) }
					line := int64(qty) * price
					line = line - (line*int64(disc))/100
					base += line
				}
				rows.Close()
				if update.DiscountPercent != nil && *update.DiscountPercent > 0 && *update.DiscountPercent <= 100 {
					newTotal = base - (base*int64(*update.DiscountPercent))/100
				} else {
					newTotal = base
				}
			}
			if newTotal != oldTotal {
				diff := newTotal - oldTotal
				if diff != 0 {
					if diff > 0 {
						if _, err := tx.ExecContext(ctx, `UPDATE client SET debt_cents = COALESCE(debt_cents,0) + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, diff, clientID); err != nil {
							return nil, fmt.Errorf("increase client debt: %w", err)
						}
					} else {
						if _, err := tx.ExecContext(ctx, `UPDATE client SET debt_cents = CASE WHEN debt_cents + ? < 0 THEN 0 ELSE debt_cents + ? END, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, diff, diff, clientID); err != nil {
							return nil, fmt.Errorf("decrease client debt: %w", err)
						}
					}
				}
			}
		}
	}
	*/

	// After any potential debt adjustments (logic currently commented), always refresh snapshot so PDFs are consistent
	if _, err := tx.ExecContext(ctx, `UPDATE "order" SET client_debt_snapshot_cents = (SELECT debt_cents FROM client WHERE id = ?) WHERE id = ?`, clientID, update.ID); err != nil {
		return nil, fmt.Errorf("failed to update debt snapshot: %w", err)
	}

	// (Removed balance column) - if future outstanding tracking is needed, compute via invoices/payments

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get updated order
	var order Order
	query := `SELECT id, order_number, client_id, status, notes, discount_percent, issue_date, due_date, client_debt_snapshot_cents, created_at, updated_at FROM "order" WHERE id = ?`
	err = r.db.QueryRowContext(ctx, query, update.ID).Scan(
		&order.ID, &order.OrderNumber, &order.ClientID, &order.Status, &order.Notes,
		&order.DiscountPercent, &order.IssueDate, &order.DueDate, &order.ClientDebtSnapshotCents,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated order: %w", err)
	}

	return &order, nil
}

// getOrderItems is a helper function to get items for an order
func (r *Repository) getOrderItems(ctx context.Context, orderID int64) ([]OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, name_snapshot, sku_snapshot, qty, unit_price_cents, discount_percent, currency, total_cents
		FROM order_item 
		WHERE order_id = ?
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.NameSnapshot,
			&item.SKUSnapshot, &item.Qty, &item.UnitPriceCents, &item.DiscountPercent, &item.Currency, &item.TotalCents,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) generateOrderNumber(ctx context.Context, tx *sql.Tx) (string, error) {
	year := time.Now().Year()

	var count int
	query := `SELECT COUNT(*) FROM "order" WHERE order_number LIKE ?`
	pattern := fmt.Sprintf("ORD-%d-%%", year)

	err := tx.QueryRowContext(ctx, query, pattern).Scan(&count)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ORD-%d-%04d", year, count+1), nil
}

func (r *Repository) generateInvoiceNumber(ctx context.Context, tx *sql.Tx) (string, error) {
	year := time.Now().Year()

	var count int
	query := `SELECT COUNT(*) FROM invoice WHERE invoice_number LIKE ?`
	pattern := fmt.Sprintf("INV-%d-%%", year)

	err := tx.QueryRowContext(ctx, query, pattern).Scan(&count)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("INV-%d-%04d", year, count+1), nil
}

// GetDashboardMetrics retrieves dashboard data
func (r *Repository) GetDashboardMetrics(ctx context.Context, timeRange string) (*DashboardData, error) {
	var data DashboardData

	// Get orders count for current month
	ordersQuery := `
		SELECT COUNT(*) FROM "order" 
		WHERE strftime('%Y-%m', issue_date) = strftime('%Y-%m', 'now')
	`
	err := r.db.QueryRowContext(ctx, ordersQuery).Scan(&data.TotalOrdersMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders count: %w", err)
	}

	// Get invoices count for current month
	invoicesQuery := `
		SELECT COUNT(*) FROM invoice 
		WHERE strftime('%Y-%m', issue_date) = strftime('%Y-%m', 'now')
	`
	err = r.db.QueryRowContext(ctx, invoicesQuery).Scan(&data.TotalInvoicesMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices count: %w", err)
	}

	// Get payments collected this month
	paymentsQuery := `
		SELECT COALESCE(SUM(amount_cents), 0) FROM payment 
		WHERE strftime('%Y-%m', paid_at) = strftime('%Y-%m', 'now')
	`
	err = r.db.QueryRowContext(ctx, paymentsQuery).Scan(&data.PaymentsCollectedMonthCents)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments: %w", err)
	}

	// Get outstanding invoices count
	outstandingQuery := `
		SELECT COUNT(*) FROM invoice 
		WHERE status IN ('DRAFT', 'ISSUED')
	`
	err = r.db.QueryRowContext(ctx, outstandingQuery).Scan(&data.OutstandingInvoicesCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get outstanding invoices: %w", err)
	}

	// Get revenue by month
	revenueQuery := `SELECT month, revenue_cents FROM vw_revenue_by_month ORDER BY month`
	rows, err := r.db.QueryContext(ctx, revenueQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue by month: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var revenue RevenueByMonth
		err := rows.Scan(&revenue.Month, &revenue.RevenueCents)
		if err != nil {
			return nil, fmt.Errorf("failed to scan revenue: %w", err)
		}
		data.RevenueByMonth = append(data.RevenueByMonth, revenue)
	}

	// Get top clients
	clientsQuery := `SELECT id, name, order_count, total_paid_cents FROM vw_top_clients`
	rows, err = r.db.QueryContext(ctx, clientsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get top clients: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var client TopClient
		err := rows.Scan(&client.ID, &client.Name, &client.OrderCount, &client.TotalPaidCents)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		data.TopClients = append(data.TopClients, client)
	}

	return &data, nil
}
