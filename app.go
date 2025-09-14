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
		log.Printf("failed to get user config dir: %v", err)
		return
	}
	appDir := filepath.Join(configDir, "myproject")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		log.Printf("failed to create app config dir: %v", err)
		return
	}

	// Database file path
	dbPath := filepath.Join(appDir, "data.db")
	log.Printf("Connecting to database at: %s", dbPath)
	database, err := db.Connect(dbPath)
	if err != nil {
		log.Printf("failed to connect to database: %v", err)
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
			log.Printf("failed to run migrations: %v", err)
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

	log.Printf("🎉 Application startup completed successfully!")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Client operations

// CreateClient creates a new client
func (a *App) CreateClient(name, phone, email, address string) (*db.Client, error) {
	var phonePtr, emailPtr, addressPtr *string
	if phone != "" {
		phonePtr = &phone
	}
	if email != "" {
		emailPtr = &email
	}
	if address != "" {
		addressPtr = &address
	}

	client := db.Client{
		Name:    name,
		Phone:   phonePtr,
		Email:   emailPtr,
		Address: addressPtr,
	}
	return a.clientService.Create(a.ctx, client)
}

// GetClients retrieves clients with pagination and search
func (a *App) GetClients(query string, limit, offset int) (*db.PaginatedResult[db.Client], error) {
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
	return a.clientService.Get(a.ctx, int64(id))
}

// UpdateClient updates an existing client
func (a *App) UpdateClient(id int, name, phone, email, address string) (*db.Client, error) {
	var phonePtr, emailPtr, addressPtr *string
	if phone != "" {
		phonePtr = &phone
	}
	if email != "" {
		emailPtr = &email
	}
	if address != "" {
		addressPtr = &address
	}

	client := db.Client{
		ID:      int64(id),
		Name:    name,
		Phone:   phonePtr,
		Email:   emailPtr,
		Address: addressPtr,
	}
	return a.clientService.Update(a.ctx, client)
}

// Product operations

// CreateProduct creates a new product
func (a *App) CreateProduct(name, description string, price float64, sku string) (*db.Product, error) {
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
		UnitPriceCents: int64(price * 100), // Convert dollars to cents
		Currency:       "USD",
		Active:         true,
	}
	return a.productService.Create(a.ctx, product)
}

// GetProducts retrieves products with pagination and search
func (a *App) GetProducts(query string, limit, offset int) (*db.PaginatedResult[db.Product], error) {
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
	return a.productService.Get(a.ctx, int64(id))
}

// UpdateProduct updates an existing product
func (a *App) UpdateProduct(id int, name, description string, price float64, sku string) (*db.Product, error) {
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
		UnitPriceCents: int64(price * 100), // Convert dollars to cents
		Currency:       "USD",
		Active:         true,
	}
	return a.productService.Update(a.ctx, product)
}

// Dashboard operations

// GetDashboardMetrics retrieves dashboard metrics and data
func (a *App) GetDashboardMetrics(timeRange string) (*db.DashboardData, error) {
	return a.repo.GetDashboardMetrics(a.ctx, timeRange)
}

// Order operations

// CreateOrder creates a new order
func (a *App) CreateOrder(clientID int, notes string, discountPercent int, items []map[string]interface{}) (*db.Order, error) {
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
	return a.orderService.Get(a.ctx, int64(id))
}

// UpdateOrder updates an existing order
func (a *App) UpdateOrder(id int, status, notes string, discountPercent *int, items []map[string]interface{}) (*db.Order, error) {
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
	return a.orderService.Delete(a.ctx, int64(id))
}

// GetOrderStatuses returns available order statuses
func (a *App) GetOrderStatuses() []string {
	return a.orderService.GetOrderStatuses()
}

// ExportOrderPDF generates and exports an order as PDF
func (a *App) ExportOrderPDF(orderID int) ([]byte, error) {
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
