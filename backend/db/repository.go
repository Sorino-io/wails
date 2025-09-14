package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository handles all database operations
type Repository struct {
	db *DB
}

// NewRepository creates a new repository instance
func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

// Client operations

func (r *Repository) CreateClient(ctx context.Context, client Client) (*Client, error) {
	query := `
		INSERT INTO client (name, phone, email, address, created_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	result, err := r.db.ExecContext(ctx, query, client.Name, client.Phone, client.Email, client.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get client ID: %w", err)
	}

	client.ID = id
	client.CreatedAt = time.Now()
	return &client, nil
}

func (r *Repository) GetClient(ctx context.Context, id int64) (*Client, error) {
	query := `SELECT id, name, phone, email, address, created_at, updated_at FROM client WHERE id = ?`

	var client Client
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&client.ID, &client.Name, &client.Phone, &client.Email, &client.Address, &client.CreatedAt, &client.UpdatedAt)
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
		SELECT id, name, phone, email, address, created_at, updated_at 
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
		err := rows.Scan(&client.ID, &client.Name, &client.Phone, &client.Email, &client.Address, &client.CreatedAt, &client.UpdatedAt)
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
		SET name = ?, phone = ?, email = ?, address = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, client.Name, client.Phone, client.Email, client.Address, client.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	return r.GetClient(ctx, client.ID)
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
	_, err := r.db.ExecContext(ctx, query, product.SKU, product.Name, product.Description,
		product.UnitPriceCents, product.Currency, product.Active, product.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return r.GetProduct(ctx, product.ID)
}

// Order operations (simplified - full implementation would include items handling)

func (r *Repository) CreateOrder(ctx context.Context, draft OrderDraft) (*Order, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
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

	// Create order
	orderQuery := `
		INSERT INTO "order" (order_number, client_id, status, notes, discount_percent, issue_date, due_date, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	result, err := tx.ExecContext(ctx, orderQuery, orderNumber, draft.ClientID, OrderStatusPending,
		draft.Notes, draft.DiscountPercent, issueDate, draft.DueDate)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get order ID: %w", err)
	}

	// Create order items
	for _, item := range draft.Items {
		totalCents := int64(item.Qty) * item.UnitPriceCents
		itemQuery := `
			INSERT INTO order_item (order_id, product_id, name_snapshot, sku_snapshot, qty, unit_price_cents, discount_percent, currency, total_cents)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := tx.ExecContext(ctx, itemQuery, orderID, item.ProductID, item.NameSnapshot,
			item.SKUSnapshot, item.Qty, item.UnitPriceCents, item.DiscountPercent, item.Currency, totalCents)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return created order
	order := &Order{
		ID:              orderID,
		OrderNumber:     orderNumber,
		ClientID:        draft.ClientID,
		Status:          OrderStatusPending,
		Notes:           draft.Notes,
		DiscountPercent: draft.DiscountPercent,
		IssueDate:       issueDate,
		DueDate:         draft.DueDate,
		CreatedAt:       time.Now(),
	}

	return order, nil
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
				o.discount_percent, o.issue_date, o.due_date,
				o.created_at, o.updated_at,
				c.id, c.name, c.phone, c.email, c.address, c.created_at, c.updated_at
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
			&order.Order.IssueDate, &order.Order.DueDate, &order.Order.CreatedAt, &order.Order.UpdatedAt,
			&order.Client.ID, &order.Client.Name, &order.Client.Phone, &order.Client.Email,
			&order.Client.Address, &order.Client.CreatedAt, &order.Client.UpdatedAt,
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

		// Calculate totals
		subtotal, discount, _, total := CalcOrderTotals(items, order.Order.DiscountPercent, 0)
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
	query := `
		SELECT 
			o.id, o.order_number, o.client_id, o.status, o.notes, 
			o.discount_percent, o.issue_date, o.due_date,
			o.created_at, o.updated_at,
			c.id, c.name, c.phone, c.email, c.address, c.created_at, c.updated_at
		FROM "order" o
		JOIN client c ON o.client_id = c.id
		WHERE o.id = ?
	`

	var order OrderDetail
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.Order.ID, &order.Order.OrderNumber, &order.Order.ClientID, &order.Order.Status,
		&order.Order.Notes, &order.Order.DiscountPercent,
		&order.Order.IssueDate, &order.Order.DueDate, &order.Order.CreatedAt, &order.Order.UpdatedAt,
		&order.Client.ID, &order.Client.Name, &order.Client.Phone, &order.Client.Email,
		&order.Client.Address, &order.Client.CreatedAt, &order.Client.UpdatedAt,
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

	// Calculate totals
	subtotal, discount, _, total := CalcOrderTotals(items, order.Order.DiscountPercent, 0)
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
			   c.id, c.name, c.phone, c.email, c.address, c.created_at, c.updated_at
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
			&client.ID, &client.Name, &client.Phone, &client.Email, &client.Address, &client.CreatedAt, &client.UpdatedAt)
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
	if len(update.Items) > 0 {
		// Delete existing items
		_, err = tx.ExecContext(ctx, `DELETE FROM order_item WHERE order_id = ?`, update.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete existing items: %w", err)
		}

		// Insert new items
		for _, item := range update.Items {
			totalCents := int64(item.Qty) * item.UnitPriceCents
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

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get updated order
	var order Order
	query := `SELECT id, order_number, client_id, status, notes, discount_percent, issue_date, due_date, created_at, updated_at FROM "order" WHERE id = ?`
	err = r.db.QueryRowContext(ctx, query, update.ID).Scan(
		&order.ID, &order.OrderNumber, &order.ClientID, &order.Status, &order.Notes,
		&order.DiscountPercent, &order.IssueDate, &order.DueDate,
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
