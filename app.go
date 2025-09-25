package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"myproject/backend/db"
	"myproject/backend/pdf"
	"myproject/backend/services"
	"os"
	"path/filepath"
)

//go:embed backend/db/migrations/*.sql
var migrationFiles embed.FS

// App struct
type App struct {
	ctx            context.Context
	db             *db.DB
	repo           *db.Repository
	clientService  *services.ClientService
	productService *services.ProductService
	orderService   *services.OrderService
	orderPDF       *pdf.OrderPDFGenerator
	amiriFont      embed.FS
	// initialization state
	initialized    bool
	initErr        error
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context provided
// will be cancelled when the app stops.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Get user config directory for the application
	configDir, err := os.UserConfigDir()
	if err != nil {
		a.initErr = fmt.Errorf("failed to get user config dir: %w", err)
		log.Printf(a.initErr.Error())
		return
	}
	appDir := filepath.Join(configDir, "myproject")

	// Database file path
	dbPath := filepath.Join(appDir, "data.db")
	log.Printf("Connecting to database at: %s", dbPath)
	database, err := db.Connect(dbPath)
	if err != nil {
		a.initErr = fmt.Errorf("failed to connect to database: %w", err)
		log.Printf(a.initErr.Error())
		return
	}
	a.db = database
	log.Printf("✓ Database connected successfully!")

	// Run migrations
	migrationsDir := "./backend/db/migrations"
	log.Printf("Running migrations from embedded files...")

	// Try to use embedded migrations first (for production builds)
	if err := a.db.RunEmbeddedMigrations(migrationFiles, "backend/db/migrations"); err != nil {
		log.Printf("Embedded migrations failed, trying file system: %v", err)
		// Fallback to file system migrations (for development)
		if err := a.db.RunMigrations(migrationsDir); err != nil {
			a.initErr = fmt.Errorf("failed to run migrations: %w", err)
			log.Printf(a.initErr.Error())
			return
		}
	}
	log.Printf("✓ Migrations completed successfully!")

	// Initialize repository and services
	a.repo = db.NewRepository(a.db)
	a.clientService = services.NewClientService(a.repo)
	a.productService = services.NewProductService(a.repo)
	a.orderService = services.NewOrderService(a.repo)
	log.Printf("✓ Services initialized successfully!")

	// Initialize PDF generators
	a.orderPDF = pdf.NewOrderPDFGenerator()
	log.Printf("✓ PDF generators initialized successfully!")

	a.initialized = true
	log.Printf("🎉 Application startup completed successfully!")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Client operations

// CreateClient creates a new client
func (a *App) CreateClient(name, phone, address string) (*db.Client, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	var phonePtr, addressPtr *string
	if phone != "" {
		phonePtr = &phone
	}
	if address != "" {
		addressPtr = &address
	}
	if v, ok := os.LookupEnv("DEBUG_CLIENTS"); ok && v != "" && v != "0" {
		log.Printf("[clients] CreateClient req name=%q phone=%q address=%q", name, phone, address)
	}
	client := db.Client{
		Name:      name,
		Phone:     phonePtr,
		Address:   addressPtr,
		DebtCents: 0,
	}
	res, err := a.clientService.Create(a.ctx, client)
	if err != nil {
		if v, ok := os.LookupEnv("DEBUG_CLIENTS"); ok && v != "" && v != "0" {
			log.Printf("[clients] CreateClient error: %v", err)
		}
		return nil, err
	}
	if v, ok := os.LookupEnv("DEBUG_CLIENTS"); ok && v != "" && v != "0" {
		log.Printf("[clients] CreateClient success id=%d", res.ID)
	}
	return res, nil
}

// GetClients retrieves clients with pagination and search
func (a *App) GetClients(query string, limit, offset int) (*db.PaginatedResult[db.Client], error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	clients, total, err := a.clientService.List(a.ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return &db.PaginatedResult[db.Client]{
		Data:  clients,
		Total: total,
	}, nil
}

// GetClient retrieves a client by ID
func (a *App) GetClient(id int) (*db.Client, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	return a.clientService.Get(a.ctx, int64(id))
}

// UpdateClient updates an existing client
func (a *App) UpdateClient(id int, name, phone, address string, debtCents int64) (*db.Client, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	var phonePtr, addressPtr *string
	if phone != "" {
		phonePtr = &phone
	}
	if address != "" {
		addressPtr = &address
	}

	client := db.Client{
		ID:        int64(id),
		Name:      name,
		Phone:     phonePtr,
		Address:   addressPtr,
		DebtCents: debtCents,
	}
	return a.clientService.Update(a.ctx, client)
}

// AdjustClientDebt adjusts a client's debt by delta cents
func (a *App) AdjustClientDebt(id int, deltaCents int64) (*db.Client, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	return a.clientService.AdjustDebt(a.ctx, int64(id), deltaCents)
}

// DeleteClient deletes a client
func (a *App) DeleteClient(id int) error {
	if err := a.ensureReady(); err != nil { return err }
	return a.clientService.Delete(a.ctx, int64(id))
}

// Product operations

// CreateProduct creates a new product
func (a *App) CreateProduct(name, description string, price float64, sku string) (*db.Product, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	var descPtr, skuPtr *string
	if description != "" {
		descPtr = &description
	}
	if sku != "" {
		skuPtr = &sku
	}

	product := db.Product{
		Name:           name,
		Description:    descPtr,
		SKU:            skuPtr,
		UnitPriceCents: int64(price), // Convert dollars to cents
		Currency:       "USD",
		Active:         true,
	}
	return a.productService.Create(a.ctx, product)
}

// GetProducts retrieves products with pagination and search
func (a *App) GetProducts(query string, limit, offset int) (*db.PaginatedResult[db.Product], error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	products, total, err := a.productService.List(a.ctx, query, nil, limit, offset)
	if err != nil {
		return nil, err
	}
	return &db.PaginatedResult[db.Product]{
		Data:  products,
		Total: total,
	}, nil
}

// GetProduct retrieves a product by ID
func (a *App) GetProduct(id int) (*db.Product, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	return a.productService.Get(a.ctx, int64(id))
}

// UpdateProduct updates an existing product
func (a *App) UpdateProduct(id int, name, description string, price float64, sku string) (*db.Product, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	var descPtr, skuPtr *string
	if description != "" {
		descPtr = &description
	}
	if sku != "" {
		skuPtr = &sku
	}

	product := db.Product{
		ID:             int64(id),
		Name:           name,
		Description:    descPtr,
		SKU:            skuPtr,
		UnitPriceCents: int64(price), // Convert dollars to cents
		Currency:       "USD",
		Active:         true,
	}
	return a.productService.Update(a.ctx, product)
}

// DeleteProduct deletes a product
func (a *App) DeleteProduct(id int) error {
	if err := a.ensureReady(); err != nil { return err }
	return a.productService.Delete(a.ctx, int64(id))
}

// Dashboard operations

// GetDashboardMetrics retrieves dashboard metrics and data
func (a *App) GetDashboardMetrics(timeRange string) (*db.DashboardData, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	return a.repo.GetDashboardMetrics(a.ctx, timeRange)
}

// Order operations

// CreateOrder creates a new order
func (a *App) CreateOrder(clientID int, notes string, discountPercent int, items []map[string]interface{}) (*db.Order, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	// Convert items from frontend format
	orderItems := make([]db.OrderItemDraft, len(items))
	for i, item := range items {
		var productID *int64
		if id, ok := item["product_id"].(float64); ok && id > 0 {
			idInt := int64(id)
			productID = &idInt
		}

		var skuSnapshot *string
		if sku, ok := item["sku_snapshot"].(string); ok && sku != "" {
			skuSnapshot = &sku
		}

		nameSnapshot, _ := item["name_snapshot"].(string)
		qty, _ := item["qty"].(float64)
		unitPriceCents, _ := item["unit_price_cents"].(float64)
		discountPercent, _ := item["discount_percent"].(float64)
		currency, _ := item["currency"].(string)
		if currency == "" {
			currency = "USD"
		}

		orderItems[i] = db.OrderItemDraft{
			ProductID:       productID,
			NameSnapshot:    nameSnapshot,
			SKUSnapshot:     skuSnapshot,
			Qty:             int(qty),
			UnitPriceCents:  int64(unitPriceCents),
			DiscountPercent: int(discountPercent),
			Currency:        currency,
		}
	}

	draft := db.OrderDraft{
		ClientID:        int64(clientID),
		Notes:           notesPtr,
		DiscountPercent: discountPercent,
		Items:           orderItems,
	}

	return a.orderService.Create(a.ctx, draft)
}

// GetOrders retrieves orders with pagination and search
func (a *App) GetOrders(query string, clientID int, status string, limit, offset int, sort string) (*db.PaginatedResult[db.OrderDetail], error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	filters := db.OrderFilters{}
	if query != "" {
		filters.Query = &query
	}
	if clientID > 0 {
		clientIDInt64 := int64(clientID)
		filters.ClientID = &clientIDInt64
	}
	if status != "" {
		filters.Status = &status
	}
	if sort != "" {
		filters.Sort = &sort
	}

	orders, total, err := a.orderService.List(a.ctx, filters, limit, offset)
	if err != nil {
		return nil, err
	}
	return &db.PaginatedResult[db.OrderDetail]{
		Data:  orders,
		Total: total,
	}, nil
}

// GetOrder retrieves an order by ID
func (a *App) GetOrder(id int) (*db.OrderDetail, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	return a.orderService.Get(a.ctx, int64(id))
}

// UpdateOrder updates an existing order
func (a *App) UpdateOrder(id int, status, notes string, discountPercent *int, items []map[string]interface{}) (*db.Order, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	update := db.OrderUpdate{
		ID: int64(id),
	}

	if status != "" {
		update.Status = &status
	}
	if notes != "" {
		update.Notes = &notes
	}
	if discountPercent != nil {
		update.DiscountPercent = discountPercent
	}

	// Convert items if provided
	if len(items) > 0 {
		orderItems := make([]db.OrderItemDraft, len(items))
		for i, item := range items {
			var productID *int64
			if id, ok := item["product_id"].(float64); ok && id > 0 {
				idInt := int64(id)
				productID = &idInt
			}

			var skuSnapshot *string
			if sku, ok := item["sku_snapshot"].(string); ok && sku != "" {
				skuSnapshot = &sku
			}

			nameSnapshot, _ := item["name_snapshot"].(string)
			qty, _ := item["qty"].(float64)
			unitPriceCents, _ := item["unit_price_cents"].(float64)
			discountPercent, _ := item["discount_percent"].(float64)
			currency, _ := item["currency"].(string)
			if currency == "" {
				currency = "USD"
			}

			orderItems[i] = db.OrderItemDraft{
				ProductID:       productID,
				NameSnapshot:    nameSnapshot,
				SKUSnapshot:     skuSnapshot,
				Qty:             int(qty),
				UnitPriceCents:  int64(unitPriceCents),
				DiscountPercent: int(discountPercent),
				Currency:        currency,
			}
		}
		update.Items = orderItems
	}

	return a.orderService.Update(a.ctx, update)
}

// DeleteOrder deletes an order (cancels it)
func (a *App) DeleteOrder(id int) error {
	if err := a.ensureReady(); err != nil {
		return err
	}
	return a.orderService.Delete(a.ctx, int64(id))
}

// GetOrderStatuses returns available order statuses
func (a *App) GetOrderStatuses() []string {
	if !a.initialized || a.orderService == nil {
		return []string{}
	}
	return a.orderService.GetOrderStatuses()
}

// DebugSchema dumps current DB schema (tables -> columns) for diagnostics
func (a *App) DebugSchema() (map[string][]string, error) {
	if err := a.ensureReady(); err != nil { return nil, err }
	return a.repo.DebugSchema(a.ctx)
}

// ExportOrderPDF generates and exports an order as PDF
func (a *App) ExportOrderPDF(orderID int) ([]byte, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	orderDetail, err := a.orderService.Get(a.ctx, int64(orderID))
	if err != nil {
		return nil, err
	}

	// Generate PDF bytes
	pdfBytes, err := a.orderPDF.GenerateOrderPDF(*orderDetail)
	if err != nil {
		return nil, err
	}

	// Log info for debugging: ensure we have items and bytes length
	log.Printf("ExportOrderPDF: orderID=%d items=%d bytes=%d\n", orderID, len(orderDetail.Items), len(pdfBytes))

	if len(pdfBytes) == 0 {
		return nil, fmt.Errorf("generated PDF is empty for order %d", orderID)
	}

	return pdfBytes, nil
}

// ensureReady verifies backend initialization before handling a request
func (a *App) ensureReady() error {
	if a.initialized && a.repo != nil && a.clientService != nil && a.productService != nil && a.orderService != nil {
		return nil
	}
	if a.initErr != nil {
		return fmt.Errorf("backend not initialized: %w", a.initErr)
	}
	return fmt.Errorf("backend not initialized yet, please wait")
}
