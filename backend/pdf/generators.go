package pdf

import (
	"bytes"
	"fmt"
	"myproject/backend/db"
	"os"
	"time"

	"github.com/01walid/goarabic"
	"github.com/go-pdf/fpdf"
)

// OrderPDFGenerator generates PDF documents for orders
type OrderPDFGenerator struct{}

// NewOrderPDFGenerator creates a new order PDF generator
func NewOrderPDFGenerator() *OrderPDFGenerator {
	return &OrderPDFGenerator{}
}

// GenerateOrderPDF generates a PDF for the given order
func (g *OrderPDFGenerator) GenerateOrderPDF(orderDetail db.OrderDetail) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	// Register Arabic-supporting font
	fontPaths := []string{
		"frontend/src/assets/fonts/frontendsrcassetsfontsAmiri-Regular.ttf",    // Dev path
		"../frontend/src/assets/fonts/frontendsrcassetsfontsAmiri-Regular.ttf", // Build path
		"./frontendsrcassetsfontsAmiri-Regular.ttf",
		// "./embedded/fonts/frontendsrcassetsfontsAmiri-Regular.ttf", // Embedded path
	}

	var fontLoaded bool
	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			pdf.AddUTF8Font("Amiri", "", path)
			fontLoaded = true
			break
		}
	}

	if !fontLoaded {
		return nil, fmt.Errorf("could not find font file in any of these locations: %v", fontPaths)
	}

	pdf.SetFont("Amiri", "", 16)

	// Helper for Arabic text (RTL)
	arabicCell := func(w, h float64, txt string, borderStr string, ln int, fill bool, link int) {
		txt = goarabic.Reverse(goarabic.ToGlyph(txt))
		pdf.CellFormat(w, h, txt, borderStr, ln, "R", fill, link, "")
	}
	ltrCell := func(w, h float64, txt string, borderStr string, ln int, fill bool, link int) {
		pdf.CellFormat(w, h, txt, borderStr, ln, "L", fill, link, "")
	}

	arabicLabelLtrValueCell := func(w, h float64, rtlLabel, ltrValue string) {
		processedRtlLabel := goarabic.Reverse(goarabic.ToGlyph(rtlLabel))
		rtlLabelWidth := pdf.GetStringWidth(processedRtlLabel)
		ltrValueWidth := pdf.GetStringWidth(ltrValue)

		x, y := pdf.GetXY()

		// Calculate start of text for right alignment
		textStartX := x + w - rtlLabelWidth - ltrValueWidth

		// Set position and draw LTR value
		pdf.SetXY(textStartX, y)
		pdf.CellFormat(ltrValueWidth, h, ltrValue, "", 0, "L", false, 0, "")

		// Set position and draw RTL label
		pdf.SetXY(textStartX+ltrValueWidth, y)
		pdf.CellFormat(rtlLabelWidth, h, processedRtlLabel, "", 0, "L", false, 0, "")

		// Move cursor to next line, preserving X
		pdf.SetXY(x, y+h)
	}

	// Header
	pdf.SetXY(120, 20)
	arabicCell(70, 10, "مصنع البركة للأنواني", "", 1, false, 0)

	// Client and Order Information
	y := pdf.GetY()
	if y < 35 {
		y = 35
	}

	// Client Information
	pdf.SetXY(20, y)
	pdf.SetFont("Amiri", "", 12)
	arabicCell(70, 8, "معلومات العميل:", "", 2, false, 0)
	pdf.SetFont("Amiri", "", 10)
	arabicCell(70, 6, "الاسم: "+orderDetail.Client.Name, "", 2, false, 0)
	if orderDetail.Client.Phone != nil {
		arabicLabelLtrValueCell(70, 6, "الهاتف: ", *orderDetail.Client.Phone)
	}
	if orderDetail.Client.Email != nil {
		arabicLabelLtrValueCell(70, 6, "البريد الإلكتروني: ", *orderDetail.Client.Email)
	}
	if orderDetail.Client.Address != nil {
		arabicCell(70, 6, "العنوان: "+*orderDetail.Client.Address, "", 2, false, 0)
	}
	yClient := pdf.GetY()

	// Order Information (fully right-aligned)
	pdf.SetXY(120, y)
	pdf.SetFont("Amiri", "", 12)
	arabicLabelLtrValueCell(70, 8, "رقم الطلب: ", orderDetail.Order.OrderNumber)
	arabicLabelLtrValueCell(70, 7, "تاريخ الإصدار: ", orderDetail.Order.IssueDate.Format("2006-01-02"))
	if orderDetail.Order.DueDate != nil {
		arabicLabelLtrValueCell(70, 7, "تاريخ الاستحقاق: ", orderDetail.Order.DueDate.Format("2006-01-02"))
	}
	arabicLabelLtrValueCell(70, 7, "الحالة: ", orderDetail.Order.Status)
	yOrder := pdf.GetY()

	if yClient > yOrder {
		pdf.SetY(yClient)
	} else {
		pdf.SetY(yOrder)
	}
	pdf.Ln(5)

	// Table headers (RTL)
	pdf.SetFont("Amiri", "", 10)
	pdf.SetFillColor(240, 240, 240)
	arabicCell(30, 8, "الإجمالي", "1", 0, true, 0)
	arabicCell(30, 8, "الخصم", "1", 0, true, 0)
	arabicCell(30, 8, "سعر الوحدة", "1", 0, true, 0)
	arabicCell(20, 8, "الكمية", "1", 0, true, 0)
	arabicCell(60, 8, "الوصف", "1", 1, true, 0)

	// Items (RTL)
	pdf.SetFont("Amiri", "", 9)
	pdf.SetFillColor(255, 255, 255)
	for _, item := range orderDetail.Items {
		discountAmount := (item.TotalCents * int64(item.DiscountPercent)) / 100
		totalAfterDiscount := item.TotalCents - discountAmount
		ltrCell(30, 7, db.FormatCurrency(totalAfterDiscount, item.Currency), "1", 0, false, 0)
		ltrCell(30, 7, fmt.Sprintf("%d%%", item.DiscountPercent), "1", 0, false, 0)
		ltrCell(30, 7, db.FormatCurrency(item.UnitPriceCents, item.Currency), "1", 0, false, 0)
		ltrCell(20, 7, fmt.Sprintf("%d", item.Qty), "1", 0, false, 0)
		arabicCell(60, 7, item.NameSnapshot, "1", 1, false, 0)
	}

	// Totals (RTL)
	pdf.Ln(5)
	pdf.SetFont("Amiri", "", 10)
	subtotal, discount, _, total := db.CalcOrderTotals(orderDetail.Items, 0, 0)
	ltrCell(35, 7, db.FormatCurrency(subtotal, "USD"), "1", 0, false, 0)
	arabicCell(135, 7, "المجموع:", "", 1, false, 0)
	if discount > 0 {
		ltrCell(35, 7, fmt.Sprintf("-%s", db.FormatCurrency(discount, "USD")), "1", 0, false, 0)
		arabicCell(135, 7, "الخصم:", "", 1, false, 0)
	}
	pdf.SetFont("Amiri", "", 12)
	ltrCell(35, 8, db.FormatCurrency(total, "USD"), "1", 0, false, 0)
	arabicCell(135, 8, "الإجمالي:", "", 1, false, 0)

	// Notes (RTL)
	if orderDetail.Order.Notes != nil && *orderDetail.Order.Notes != "" {
		pdf.Ln(10)
		pdf.SetFont("Amiri", "", 10)
		arabicCell(40, 6, "ملاحظات:", "", 1, false, 0)
		pdf.SetFont("Amiri", "", 9)
		arabicCell(170, 5, *orderDetail.Order.Notes, "", 1, false, 0)
	}

	// Footer: label RTL, date LTR
	pdf.Ln(15)
	pdf.SetFont("Amiri", "", 8)
	arabicLabelLtrValueCell(170, 5, "تم الإنشاء في: ", time.Now().Format("02/01/2006 15:04"))

	// Return PDF bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
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
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	// Register Arabic-supporting font
	fontPaths := []string{
		"frontend/src/assets/fonts/frontendsrcassetsfontsAmiri-Regular.ttf",    // Dev path
		"../frontend/src/assets/fonts/frontendsrcassetsfontsAmiri-Regular.ttf", // Build path
		"./frontendsrcassetsfontsAmiri-Regular.ttf",                            // Direct path
	}

	var fontLoaded bool
	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			pdf.AddUTF8Font("Amiri", "", path)
			fontLoaded = true
			break
		}
	}

	if !fontLoaded {
		return nil, fmt.Errorf("could not find font file in any of these locations: %v", fontPaths)
	}

	pdf.SetFont("Amiri", "", 16)
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
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}
	return buf.Bytes(), nil
}
