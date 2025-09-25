package db

import (
	"time"
)

// Common pagination result
type PaginatedResult[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
}

// Client represents a customer
type Client struct {
	ID        int64      `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Phone     *string    `json:"phone" db:"phone"`
	DebtCents int64      `json:"debt_cents" db:"debt_cents"`
	Address   *string    `json:"address" db:"address"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

// Product represents a sellable item
type Product struct {
	ID             int64      `json:"id" db:"id"`
	SKU            *string    `json:"sku" db:"sku"`
	Name           string     `json:"name" db:"name"`
	Description    *string    `json:"description" db:"description"`
	UnitPriceCents int64      `json:"unit_price_cents" db:"unit_price_cents"`
	Currency       string     `json:"currency" db:"currency"`
	Active         bool       `json:"active" db:"active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at" db:"updated_at"`
}

// Order represents a customer order
type Order struct {
	ID              int64      `json:"id" db:"id"`
	OrderNumber     string     `json:"order_number" db:"order_number"`
	ClientID        int64      `json:"client_id" db:"client_id"`
	Status          string     `json:"status" db:"status"`
	Notes           *string    `json:"notes" db:"notes"`
	DiscountPercent int        `json:"discount_percent" db:"discount_percent"`
	IssueDate       time.Time  `json:"issue_date" db:"issue_date"`
	DueDate         *time.Time `json:"due_date" db:"due_date"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at" db:"updated_at"`
	RemainingCents  int64      `json:"remaining_cents" db:"remaining_cents"`
}

// OrderItem represents a line item in an order
type OrderItem struct {
	ID              int64   `json:"id" db:"id"`
	OrderID         int64   `json:"order_id" db:"order_id"`
	ProductID       *int64  `json:"product_id" db:"product_id"`
	NameSnapshot    string  `json:"name_snapshot" db:"name_snapshot"`
	SKUSnapshot     *string `json:"sku_snapshot" db:"sku_snapshot"`
	Qty             int     `json:"qty" db:"qty"`
	UnitPriceCents  int64   `json:"unit_price_cents" db:"unit_price_cents"`
	DiscountPercent int     `json:"discount_percent" db:"discount_percent"`
	Currency        string  `json:"currency" db:"currency"`
	TotalCents      int64   `json:"total_cents" db:"total_cents"`
}

// Invoice represents a bill sent to customer
type Invoice struct {
	ID              int64      `json:"id" db:"id"`
	InvoiceNumber   string     `json:"invoice_number" db:"invoice_number"`
	OrderID         *int64     `json:"order_id" db:"order_id"`
	ClientID        int64      `json:"client_id" db:"client_id"`
	Status          string     `json:"status" db:"status"`
	IssueDate       time.Time  `json:"issue_date" db:"issue_date"`
	DueDate         *time.Time `json:"due_date" db:"due_date"`
	Notes           *string    `json:"notes" db:"notes"`
	SubtotalCents   int64      `json:"subtotal_cents" db:"subtotal_cents"`
	DiscountPercent int        `json:"discount_percent" db:"discount_percent"`
	TaxPercent      int        `json:"tax_percent" db:"tax_percent"`
	TotalCents      int64      `json:"total_cents" db:"total_cents"`
	Currency        string     `json:"currency" db:"currency"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at" db:"updated_at"`
}

// InvoiceItem represents a line item in an invoice
type InvoiceItem struct {
	ID             int64   `json:"id" db:"id"`
	InvoiceID      int64   `json:"invoice_id" db:"invoice_id"`
	ProductID      *int64  `json:"product_id" db:"product_id"`
	NameSnapshot   string  `json:"name_snapshot" db:"name_snapshot"`
	SKUSnapshot    *string `json:"sku_snapshot" db:"sku_snapshot"`
	Qty            int     `json:"qty" db:"qty"`
	UnitPriceCents int64   `json:"unit_price_cents" db:"unit_price_cents"`
	Currency       string  `json:"currency" db:"currency"`
	TotalCents     int64   `json:"total_cents" db:"total_cents"`
}

// Payment represents a payment made against an invoice
type Payment struct {
	ID          int64     `json:"id" db:"id"`
	InvoiceID   int64     `json:"invoice_id" db:"invoice_id"`
	AmountCents int64     `json:"amount_cents" db:"amount_cents"`
	Method      string    `json:"method" db:"method"`
	Reference   *string   `json:"reference" db:"reference"`
	PaidAt      time.Time `json:"paid_at" db:"paid_at"`
	Notes       *string   `json:"notes" db:"notes"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// DTOs for complex operations

// OrderDetail includes order with client and items
type OrderDetail struct {
	Order         Order       `json:"order"`
	Client        Client      `json:"client"`
	Items         []OrderItem `json:"items"`
	SubtotalCents int64       `json:"subtotal_cents"`
	DiscountCents int64       `json:"discount_cents"`
	TaxCents      int64       `json:"tax_cents"`
	TotalCents    int64       `json:"total_cents"`
}

// InvoiceDetail includes invoice with client, items and payments
type InvoiceDetail struct {
	Invoice      Invoice       `json:"invoice"`
	Client       Client        `json:"client"`
	Items        []InvoiceItem `json:"items"`
	Payments     []Payment     `json:"payments"`
	PaidCents    int64         `json:"paid_cents"`
	BalanceCents int64         `json:"balance_cents"`
}

// OrderDraft for creating new orders
type OrderDraft struct {
	ClientID        int64            `json:"client_id"`
	Notes           *string          `json:"notes"`
	DiscountPercent int              `json:"discount_percent"`
	IssueDate       *time.Time       `json:"issue_date"`
	DueDate         *time.Time       `json:"due_date"`
	Items           []OrderItemDraft `json:"items"`
}

// OrderItemDraft for creating order items
type OrderItemDraft struct {
	ProductID       *int64  `json:"product_id"`
	NameSnapshot    string  `json:"name_snapshot"`
	SKUSnapshot     *string `json:"sku_snapshot"`
	Qty             int     `json:"qty"`
	UnitPriceCents  int64   `json:"unit_price_cents"`
	DiscountPercent int     `json:"discount_percent"`
	Currency        string  `json:"currency"`
}

// OrderUpdate for updating orders
type OrderUpdate struct {
	ID              int64            `json:"id"`
	Status          *string          `json:"status"`
	Notes           *string          `json:"notes"`
	DiscountPercent *int             `json:"discount_percent"`
	DueDate         *time.Time       `json:"due_date"`
	Items           []OrderItemDraft `json:"items"`
}

// InvoiceOverrides for creating invoice from order
type InvoiceOverrides struct {
	IssueDate       *time.Time `json:"issue_date"`
	DueDate         *time.Time `json:"due_date"`
	Notes           *string    `json:"notes"`
	DiscountPercent *int       `json:"discount_percent"`
	TaxPercent      *int       `json:"tax_percent"`
}

// InvoiceDraft for creating new invoices
type InvoiceDraft struct {
	OrderID         *int64             `json:"order_id"`
	ClientID        int64              `json:"client_id"`
	Notes           *string            `json:"notes"`
	DiscountPercent int                `json:"discount_percent"`
	TaxPercent      int                `json:"tax_percent"`
	IssueDate       *time.Time         `json:"issue_date"`
	DueDate         *time.Time         `json:"due_date"`
	Items           []InvoiceItemDraft `json:"items"`
	Currency        string             `json:"currency"`
}

// InvoiceUpdate for updating invoices
type InvoiceUpdate struct {
	ID              int64              `json:"id"`
	Status          *string            `json:"status"`
	Notes           *string            `json:"notes"`
	DiscountPercent *int               `json:"discount_percent"`
	TaxPercent      *int               `json:"tax_percent"`
	DueDate         *time.Time         `json:"due_date"`
	Items           []InvoiceItemDraft `json:"items"`
}

// InvoiceItemDraft for creating/updating invoice items
type InvoiceItemDraft struct {
	ProductID      *int64  `json:"product_id"`
	NameSnapshot   string  `json:"name_snapshot"`
	SKUSnapshot    *string `json:"sku_snapshot"`
	Qty            int     `json:"qty"`
	UnitPriceCents int64   `json:"unit_price_cents"`
	Currency       string  `json:"currency"`
}

// OrderFilters for filtering orders list
type OrderFilters struct {
	ClientID *int64  `json:"client_id"`
	Status   *string `json:"status"`
	Query    *string `json:"query"`
	Sort     *string `json:"sort"`
}

// DashboardData for dashboard metrics
type DashboardData struct {
	TotalOrdersMonth            int              `json:"total_orders_month"`
	TotalInvoicesMonth          int              `json:"total_invoices_month"`
	PaymentsCollectedMonthCents int64            `json:"payments_collected_month_cents"`
	OutstandingInvoicesCount    int              `json:"outstanding_invoices_count"`
	RevenueByMonth              []RevenueByMonth `json:"revenue_by_month"`
	TopClients                  []TopClient      `json:"top_clients"`
}

// RevenueByMonth for dashboard charts
type RevenueByMonth struct {
	Month        string `json:"month" db:"month"`
	RevenueCents int64  `json:"revenue_cents" db:"revenue_cents"`
}

// TopClient for dashboard
type TopClient struct {
	ID             int64  `json:"id" db:"id"`
	Name           string `json:"name" db:"name"`
	OrderCount     int    `json:"order_count" db:"order_count"`
	TotalPaidCents int64  `json:"total_paid_cents" db:"total_paid_cents"`
}

// Constants for statuses
const (
	OrderStatusPending   = "PENDING"
	OrderStatusConfirmed = "CONFIRMED"
	OrderStatusCanceled  = "CANCELED"
	OrderStatusCompleted = "COMPLETED"

	InvoiceStatusDraft    = "DRAFT"
	InvoiceStatusIssued   = "ISSUED"
	InvoiceStatusPaid     = "PAID"
	InvoiceStatusCanceled = "CANCELED"

	PaymentMethodCash     = "CASH"
	PaymentMethodCard     = "CARD"
	PaymentMethodTransfer = "TRANSFER"
	PaymentMethodOther    = "OTHER"
)
