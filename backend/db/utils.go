package db

import (
	"fmt"
	"time"
)

// CalcOrderTotals calculates order totals based on items, discount and tax percentages
func CalcOrderTotals(items []OrderItem, discountPct, taxPct int) (subtotal, discount, tax, total int64) {
	// Calculate subtotal and discount from items
	for _, item := range items {
		itemSubtotal := item.TotalCents
		itemDiscount := (itemSubtotal * int64(item.DiscountPercent)) / 100
		subtotal += itemSubtotal
		discount += itemDiscount
	}

	// Apply order-level discount
	orderDiscount := (subtotal * int64(discountPct)) / 100
	discount += orderDiscount

	// Calculate total
	total = subtotal - discount

	return subtotal, discount, 0, total
}

// CalcInvoiceTotals calculates invoice totals based on items, discount and tax percentages
func CalcInvoiceTotals(items []InvoiceItem, discountPct, taxPct int) (subtotal, discount, tax, total int64) {
	// Calculate subtotal
	for _, item := range items {
		subtotal += item.TotalCents
	}

	// Calculate discount
	if discountPct > 0 && discountPct <= 100 {
		discount = (subtotal * int64(discountPct)) / 100
	}

	// Calculate tax on (subtotal - discount)
	taxableAmount := subtotal - discount
	if taxPct > 0 && taxPct <= 100 {
		tax = (taxableAmount * int64(taxPct)) / 100
	}

	// Calculate total
	total = subtotal - discount + tax

	return subtotal, discount, tax, total
}

// NewOrderNumber generates a new order number in format ORD-YYYY-####
func NewOrderNumber() string {
	year := time.Now().Year()
	// This is a simplified version - in real implementation, you'd query the DB for the next sequence
	return fmt.Sprintf("ORD-%d-%04d", year, 1)
}

// NewInvoiceNumber generates a new invoice number in format INV-YYYY-####
func NewInvoiceNumber() string {
	year := time.Now().Year()
	// This is a simplified version - in real implementation, you'd query the DB for the next sequence
	return fmt.Sprintf("INV-%d-%04d", year, 1)
}

// FormatCents converts cents to display format (e.g., 12345 -> "123.45")
func FormatCents(cents int64) string {
	return fmt.Sprintf("%.2f", float64(cents)/100)
}

// FormatCurrency formats cents with currency symbol
func FormatCurrency(cents int64, currency string) string {
	amount := FormatCents(cents)
	// switch currency {
	// case "DZD":
	return fmt.Sprintf("%s DZD", amount)
	// case "USD":
	// 	return fmt.Sprintf("$%s", amount)
	// case "EUR":
	// 	return fmt.Sprintf("€%s", amount)
	// default:
	// 	return fmt.Sprintf("%s %s", amount, currency)
	// }
}

// ParseCentsFromFloat converts float to cents (e.g., 123.45 -> 12345)
func ParseCentsFromFloat(amount float64) int64 {
	return int64(amount * 100)
}

// FormatDateArabic formats date in Arabic format (dd/MM/yyyy HH:mm)
func FormatDateArabic(t time.Time) string {
	return t.Format("02/01/2006 15:04")
}

// FormatDateForDB formats date for SQLite storage (ISO format)
func FormatDateForDB(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseDateFromDB parses date from SQLite storage
func ParseDateFromDB(dateStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateStr)
}

// ValidateDiscountPercent validates discount percentage (0-100)
func ValidateDiscountPercent(pct int) error {
	if pct < 0 || pct > 100 {
		return fmt.Errorf("discount percentage must be between 0 and 100")
	}
	return nil
}

// ValidateTaxPercent validates tax percentage (0-100)
func ValidateTaxPercent(pct int) error {
	if pct < 0 || pct > 100 {
		return fmt.Errorf("tax percentage must be between 0 and 100")
	}
	return nil
}

// ValidateQuantity validates item quantity (must be positive)
func ValidateQuantity(qty int) error {
	if qty <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	return nil
}

// ValidatePrice validates price in cents (must be non-negative)
func ValidatePrice(priceCents int64) error {
	if priceCents < 0 {
		return fmt.Errorf("price cannot be negative")
	}
	return nil
}

// IsValidOrderStatus checks if order status is valid
func IsValidOrderStatus(status string) bool {
	validStatuses := []string{OrderStatusPending, OrderStatusConfirmed, OrderStatusCanceled, OrderStatusCompleted}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidInvoiceStatus checks if invoice status is valid
func IsValidInvoiceStatus(status string) bool {
	validStatuses := []string{InvoiceStatusDraft, InvoiceStatusIssued, InvoiceStatusPaid, InvoiceStatusCanceled}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// IsValidPaymentMethod checks if payment method is valid
func IsValidPaymentMethod(method string) bool {
	validMethods := []string{PaymentMethodCash, PaymentMethodCard, PaymentMethodTransfer, PaymentMethodOther}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

// CalculateItemTotal calculates the total for an order/invoice item
func CalculateItemTotal(qty int, unitPriceCents int64) int64 {
	return int64(qty) * unitPriceCents
}

// CalculateInvoiceBalance calculates remaining balance for an invoice
func CalculateInvoiceBalance(totalCents int64, payments []Payment) int64 {
	var paidCents int64
	for _, payment := range payments {
		paidCents += payment.AmountCents
	}
	return totalCents - paidCents
}
