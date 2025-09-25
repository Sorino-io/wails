package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"myproject/backend/db"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/01walid/goarabic"
	"github.com/go-pdf/fpdf"
)

//go:embed embedded/fonts/frontendsrcassetsfontsAmiri-Regular.ttf
var amiriFont []byte

// registerAmiriFont registers the Amiri UTF-8 font with fpdf.
// It first tries known file paths (dev/build) and falls back to the embedded bytes.
func registerAmiriFont(pdf *fpdf.Fpdf) error {
	// Try existing file paths (useful in dev)
	fontPaths := []string{
		"frontend/src/assets/fonts/frontendsrcassetsfontsAmiri-Regular.ttf",    // Dev path
		"../frontend/src/assets/fonts/frontendsrcassetsfontsAmiri-Regular.ttf", // Build path
		"./frontendsrcassetsfontsAmiri-Regular.ttf",                            // Direct path
	}

	for _, path := range fontPaths {
		if _, err := os.Stat(path); err == nil {
			pdf.AddUTF8Font("Amiri", "", path)
			return nil
		}
	}

	// Fall back to embedded font bytes when running from a packaged binary
	if len(amiriFont) > 0 {
		tmpFile, err := os.CreateTemp("", "amiri-*.ttf")
		if err != nil {
			return fmt.Errorf("failed to create temp font file: %w", err)
		}
		tmpPath := tmpFile.Name()
		if _, err := tmpFile.Write(amiriFont); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
			return fmt.Errorf("failed to write temp font file: %w", err)
		}
		_ = tmpFile.Close()

		// Ensure cleanup regardless of success
		defer os.Remove(tmpPath)

		pdf.AddUTF8Font("Amiri", "", tmpPath)
		return nil
	}

	return fmt.Errorf("could not load Amiri font from disk or embedded resources")
}

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

	// Register Arabic-supporting font (robust in dev & build)
	if err := registerAmiriFont(pdf); err != nil {
		return nil, err
	}

	pdf.SetFont("Amiri", "", 16)

	// Helper for Arabic text (RTL)
	arabicCell := func(w, h float64, txt string, borderStr string, ln int, fill bool, link int) {
		shapedTxt := goarabic.ToGlyph(txt)
		words := strings.Split(shapedTxt, " ")

		// Reverse the order of words for RTL layout
		for i, j := 0, len(words)-1; i < j; i, j = i+1, j-1 {
			words[i], words[j] = words[j], words[i]
		}

		// Reverse individual words if they are Arabic
		for i, word := range words {
			isArabic := false
			for _, r := range word {
				if unicode.Is(unicode.Arabic, r) {
					isArabic = true
					break
				}
			}

			if isArabic {
				words[i] = goarabic.Reverse(word)
			}
		}

		processedTxt := strings.Join(words, " ")
		pdf.CellFormat(w, h, processedTxt, borderStr, ln, "R", fill, link, "")
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
	arabicCell(70, 10, "البركة لللإنتاج الصناعي للأدوات المنزلية", "", 1, false, 0)

	// Client and Order Information
	y := pdf.GetY()
	if y < 35 {
		y = 35
	}

	// Client Information
	pdf.SetXY(20, y)
	pdf.SetFont("Amiri", "", 12)
	arabicLabelLtrValueCell(70, 7, "تاريخ الإصدار: ", orderDetail.Order.IssueDate.Format("2006-01-02"))
	arabicCell(70, 8, "مطلوب من العميل :", "", 2, false, 0)
	pdf.SetFont("Amiri", "", 10)
	arabicCell(70, 6, "الاسم: "+orderDetail.Client.Name, "", 2, false, 0)
	if orderDetail.Client.Phone != nil {
		arabicLabelLtrValueCell(70, 6, "الهاتف: ", *orderDetail.Client.Phone)
	}
	// if orderDetail.Client.Email != nil {
	// 	arabicLabelLtrValueCell(70, 6, "البريد الإلكتروني: ", *orderDetail.Client.Email)
	// }
	if orderDetail.Client.Address != nil {
		arabicCell(70, 6, "العنوان: "+*orderDetail.Client.Address, "", 2, false, 0)
	}
	yClient := pdf.GetY()

	// Order Information (fully right-aligned)
	pdf.SetXY(120, y)
	pdf.SetFont("Amiri", "", 12)
	arabicLabelLtrValueCell(70, 8, "الهاتف : ", "032 23 19 99")
	arabicLabelLtrValueCell(70, 8, "رقم الطلب: ", orderDetail.Order.OrderNumber)
	arabicLabelLtrValueCell(70, 7, "البريد الإلكتروني : ", "elbarakaaouani@gmail.com")
	arabicLabelLtrValueCell(70, 7, "العنوان والرمز البريدي : قمار ولاية الوادي ص.ب  : ", "39400-331")
	if orderDetail.Order.DueDate != nil {
		arabicLabelLtrValueCell(70, 7, "تاريخ الاستحقاق: ", orderDetail.Order.DueDate.Format("2006-01-02"))
	}
	// arabicLabelLtrValueCell(70, 7, "الحالة: ", orderDetail.Order.Status)
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
	arabicCell(60, 8, "التعيين", "1", 1, true, 0)

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

	// Totals (RTL) - Order of presentation required:
	// 1) Order total after discount
	// 2) Client previous debt (snapshot if available)
	// 3) Combined (order total + previous debt)
	pdf.Ln(5)
	pdf.SetFont("Amiri", "", 10)
	_, discount, _, total := db.CalcOrderTotals(orderDetail.Items, 0, 0)
	if discount > 0 {
		ltrCell(35, 7, fmt.Sprintf("-%s", db.FormatCurrency(discount, "USD")), "1", 0, false, 0)
		arabicCell(135, 7, "الخصم:", "", 1, false, 0)
	}
	// Line 1: Order total after discount
	pdf.SetFont("Amiri", "", 12)
	ltrCell(35, 8, db.FormatCurrency(total, "USD"), "1", 0, false, 0)
	arabicCell(135, 8, "مجموع الطلب:", "", 1, false, 0)

	// Line 2: Previous client debt (snapshot preferred)
	debtToShow := orderDetail.Client.DebtCents
	if orderDetail.Order.ClientDebtSnapshotCents != nil {
		debtToShow = *orderDetail.Order.ClientDebtSnapshotCents
	}
	pdf.SetFont("Amiri", "", 11)
	ltrCell(35, 8, db.FormatCurrency(debtToShow, "USD"), "1", 0, false, 0)
	arabicCell(135, 8, "دين سابق للعميل:", "", 1, false, 0)

	// Line 3: Combined total (order total + previous debt)
	combined := total + debtToShow
	pdf.SetFont("Amiri", "", 12)
	ltrCell(35, 8, db.FormatCurrency(combined, "USD"), "1", 0, false, 0)
	arabicCell(135, 8, "الإجمالي مع الدين:", "", 1, false, 0)

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

	// Register Arabic-supporting font (robust in dev & build)
	if err := registerAmiriFont(pdf); err != nil {
		return nil, err
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

	// Email removed from client model

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
