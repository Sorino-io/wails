package pdf

import (
	"bytes"
	"fmt"
	"myproject/backend/db"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
)

// OrderPDFGenerator generates PDF documents for orders
type OrderPDFGenerator struct{}

// NewOrderPDFGenerator creates a new order PDF generator
func NewOrderPDFGenerator() *OrderPDFGenerator {
	return &OrderPDFGenerator{}
}

// GenerateOrderPDF generates a PDF for the given order
func (g *OrderPDFGenerator) GenerateOrderPDF(orderDetail db.OrderDetail) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	// Add Arabic font support (if available)
	// For now, use core Helvetica font to avoid missing-font issues
	pdf.SetFont("Helvetica", "B", 16)

	// Header - Company Information
	pdf.SetXY(20, 20)
	pdf.CellFormat(170, 10, "Order Management System", "", 1, "L", false, 0, "")
	pdf.Ln(10)

	// Order Information
	pdf.SetFont("Helvetica", "B", 14)
	pdf.Cell(40, 10, fmt.Sprintf("Order #: %s", orderDetail.Order.OrderNumber))
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("Issue Date: %s", orderDetail.Order.IssueDate.Format("2006-01-02")))
	pdf.Ln(6)

	if orderDetail.Order.DueDate != nil {
		pdf.Cell(40, 6, fmt.Sprintf("Due Date: %s", orderDetail.Order.DueDate.Format("2006-01-02")))
		pdf.Ln(6)
	}

	pdf.Cell(40, 6, fmt.Sprintf("Status: %s", orderDetail.Order.Status))
	pdf.Ln(10)

	// Client Information
	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(40, 8, "Client Information")
	pdf.Ln(8)

	pdf.SetFont("Helvetica", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("Name: %s", orderDetail.Client.Name))
	pdf.Ln(6)

	if orderDetail.Client.Phone != nil {
		pdf.Cell(40, 6, fmt.Sprintf("Phone: %s", *orderDetail.Client.Phone))
		pdf.Ln(6)
	}

	if orderDetail.Client.Email != nil {
		pdf.Cell(40, 6, fmt.Sprintf("Email: %s", *orderDetail.Client.Email))
		pdf.Ln(6)
	}

	if orderDetail.Client.Address != nil {
		pdf.Cell(40, 6, fmt.Sprintf("Address: %s", *orderDetail.Client.Address))
		pdf.Ln(6)
	}
	pdf.Ln(5)

	// Items Table Header
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(240, 240, 240)

	// Table headers
	pdf.CellFormat(80, 8, "Description", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Unit Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Total", "1", 1, "C", true, 0, "")

	// Items
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetFillColor(255, 255, 255)

	for _, item := range orderDetail.Items {
		pdf.CellFormat(80, 7, item.NameSnapshot, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 7, fmt.Sprintf("%d", item.Qty), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 7, db.FormatCurrency(item.UnitPriceCents, item.Currency), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, db.FormatCurrency(item.TotalCents, item.Currency), "1", 1, "R", false, 0, "")
	}

	// Totals
	pdf.Ln(5)
	pdf.SetFont("Helvetica", "B", 10)

	// Calculate totals
	subtotal, discount, tax, total := db.CalcOrderTotals(orderDetail.Items, orderDetail.Order.DiscountPercent, orderDetail.Order.TaxPercent)

	// Subtotal
	pdf.CellFormat(135, 7, "Subtotal:", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 7, db.FormatCurrency(subtotal, "USD"), "1", 1, "R", false, 0, "")

	// Discount
	if orderDetail.Order.DiscountPercent > 0 {
		pdf.CellFormat(135, 7, fmt.Sprintf("Discount (%d%%):", orderDetail.Order.DiscountPercent), "", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, fmt.Sprintf("-%s", db.FormatCurrency(discount, "USD")), "1", 1, "R", false, 0, "")
	}

	// Tax
	if orderDetail.Order.TaxPercent > 0 {
		pdf.CellFormat(135, 7, fmt.Sprintf("Tax (%d%%):", orderDetail.Order.TaxPercent), "", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, db.FormatCurrency(tax, "USD"), "1", 1, "R", false, 0, "")
	}

	// Total
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(135, 8, "Total:", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, db.FormatCurrency(total, "USD"), "1", 1, "R", false, 0, "")

	// Notes
	if orderDetail.Order.Notes != nil && *orderDetail.Order.Notes != "" {
		pdf.Ln(10)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.Cell(40, 6, "Notes:")
		pdf.Ln(6)
		pdf.SetFont("Helvetica", "", 9)
		pdf.MultiCell(170, 5, *orderDetail.Order.Notes, "", "L", false)
	}

	// Footer
	pdf.Ln(15)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.Cell(170, 5, fmt.Sprintf("Generated on %s", time.Now().Format("02/01/2006 15:04")))

	// Return PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}

// InvoicePDFGenerator generates PDF documents for invoices
type InvoicePDFGenerator struct{}

// NewInvoicePDFGenerator creates a new invoice PDF generator
func NewInvoicePDFGenerator() *InvoicePDFGenerator {
	return &InvoicePDFGenerator{}
}

// GenerateInvoicePDF generates a PDF for the given invoice
func (g *InvoicePDFGenerator) GenerateInvoicePDF(invoiceDetail db.InvoiceDetail) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	// Header - Company Information
	pdf.SetFont("Arial", "B", 16)
	pdf.SetXY(20, 20)
	pdf.CellFormat(170, 10, "Invoice Management System", "", 1, "L", false, 0, "")
	pdf.Ln(10)

	// Invoice Information
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, fmt.Sprintf("Invoice #: %s", invoiceDetail.Invoice.InvoiceNumber))
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("Issue Date: %s", invoiceDetail.Invoice.IssueDate.Format("2006-01-02")))
	pdf.Ln(6)

	if invoiceDetail.Invoice.DueDate != nil {
		pdf.Cell(40, 6, fmt.Sprintf("Due Date: %s", invoiceDetail.Invoice.DueDate.Format("2006-01-02")))
		pdf.Ln(6)
	}

	pdf.Cell(40, 6, fmt.Sprintf("Status: %s", invoiceDetail.Invoice.Status))
	pdf.Ln(10)

	// Client Information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Client Information")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("Name: %s", invoiceDetail.Client.Name))
	pdf.Ln(6)

	if invoiceDetail.Client.Phone != nil {
		pdf.Cell(40, 6, fmt.Sprintf("Phone: %s", *invoiceDetail.Client.Phone))
		pdf.Ln(6)
	}

	if invoiceDetail.Client.Email != nil {
		pdf.Cell(40, 6, fmt.Sprintf("Email: %s", *invoiceDetail.Client.Email))
		pdf.Ln(6)
	}

	if invoiceDetail.Client.Address != nil {
		pdf.Cell(40, 6, fmt.Sprintf("Address: %s", *invoiceDetail.Client.Address))
		pdf.Ln(6)
	}
	pdf.Ln(5)

	// Items Table Header
	pdf.SetFont("Arial", "B", 10)
	pdf.SetFillColor(240, 240, 240)

	// Table headers
	pdf.CellFormat(80, 8, "Description", "1", 0, "C", true, 0, "")
	pdf.CellFormat(20, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Unit Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Total", "1", 1, "C", true, 0, "")

	// Items
	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(255, 255, 255)

	for _, item := range invoiceDetail.Items {
		pdf.CellFormat(80, 7, item.NameSnapshot, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 7, fmt.Sprintf("%d", item.Qty), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 7, db.FormatCurrency(item.UnitPriceCents, item.Currency), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, db.FormatCurrency(item.TotalCents, item.Currency), "1", 1, "R", false, 0, "")
	}

	// Totals
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 10)

	// Subtotal
	pdf.CellFormat(135, 7, "Subtotal:", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 7, db.FormatCurrency(invoiceDetail.Invoice.SubtotalCents, invoiceDetail.Invoice.Currency), "1", 1, "R", false, 0, "")

	// Discount
	if invoiceDetail.Invoice.DiscountPercent > 0 {
		discountAmount := (invoiceDetail.Invoice.SubtotalCents * int64(invoiceDetail.Invoice.DiscountPercent)) / 100
		pdf.CellFormat(135, 7, fmt.Sprintf("Discount (%d%%):", invoiceDetail.Invoice.DiscountPercent), "", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, fmt.Sprintf("-%s", db.FormatCurrency(discountAmount, invoiceDetail.Invoice.Currency)), "1", 1, "R", false, 0, "")
	}

	// Tax
	if invoiceDetail.Invoice.TaxPercent > 0 {
		taxableAmount := invoiceDetail.Invoice.SubtotalCents
		if invoiceDetail.Invoice.DiscountPercent > 0 {
			discountAmount := (invoiceDetail.Invoice.SubtotalCents * int64(invoiceDetail.Invoice.DiscountPercent)) / 100
			taxableAmount -= discountAmount
		}
		taxAmount := (taxableAmount * int64(invoiceDetail.Invoice.TaxPercent)) / 100
		pdf.CellFormat(135, 7, fmt.Sprintf("Tax (%d%%):", invoiceDetail.Invoice.TaxPercent), "", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, db.FormatCurrency(taxAmount, invoiceDetail.Invoice.Currency), "1", 1, "R", false, 0, "")
	}

	// Total
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(135, 8, "Total:", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, db.FormatCurrency(invoiceDetail.Invoice.TotalCents, invoiceDetail.Invoice.Currency), "1", 1, "R", false, 0, "")

	// Payment Summary
	if len(invoiceDetail.Payments) > 0 {
		pdf.Ln(10)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(40, 8, "Payments")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 9)
		for _, payment := range invoiceDetail.Payments {
			pdf.Cell(40, 6, fmt.Sprintf("%s: %s (%s)",
				payment.PaidAt.Format("2006-01-02"),
				db.FormatCurrency(payment.AmountCents, invoiceDetail.Invoice.Currency),
				payment.Method))
			pdf.Ln(6)
		}

		pdf.Ln(3)
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(135, 7, "Paid Amount:", "", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, db.FormatCurrency(invoiceDetail.PaidCents, invoiceDetail.Invoice.Currency), "1", 1, "R", false, 0, "")

		pdf.CellFormat(135, 7, "Balance:", "", 0, "R", false, 0, "")
		pdf.CellFormat(35, 7, db.FormatCurrency(invoiceDetail.BalanceCents, invoiceDetail.Invoice.Currency), "1", 1, "R", false, 0, "")
	}

	// Notes
	if invoiceDetail.Invoice.Notes != nil && *invoiceDetail.Invoice.Notes != "" {
		pdf.Ln(10)
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(40, 6, "Notes:")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(170, 5, *invoiceDetail.Invoice.Notes, "", "L", false)
	}

	// Footer with signature area
	pdf.Ln(15)
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(85, 10, "Customer Signature")
	pdf.Cell(85, 10, "Company Signature")
	pdf.Ln(20)

	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(170, 5, fmt.Sprintf("Generated on %s", time.Now().Format("02/01/2006 15:04")))

	// Return PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}
