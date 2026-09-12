package input

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"invoice-generator/internal/invoice"
)

const (
	invoiceSheetName = "Invoice"
	itemsSheetName   = "Items"

	dateLayout = "2006-01-02"
)

// ParseInvoiceExcel opens an .xlsx workbook and converts the "Invoice" and
// "Items" sheets into a domain Invoice (roadmap step 13).
//
// The returned invoice contains parsed data only: FinalUnitPrice and Price
// on items are not computed here.
func ParseInvoiceExcel(path string) (invoice.Invoice, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if !containsSheet(sheets, invoiceSheetName) {
		return invoice.Invoice{}, fmt.Errorf("missing sheet: %s", invoiceSheetName)
	}
	if !containsSheet(sheets, itemsSheetName) {
		return invoice.Invoice{}, fmt.Errorf("missing sheet: %s", itemsSheetName)
	}

	inv, err := parseInvoiceSheet(f)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("sheet %q: %w", invoiceSheetName, err)
	}

	items, err := parseItemsExcelSheet(f)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("sheet %q: %w", itemsSheetName, err)
	}

	inv.Items = items
	return inv, nil
}

func containsSheet(sheets []string, name string) bool {
	for _, s := range sheets {
		if strings.EqualFold(strings.TrimSpace(s), name) {
			return true
		}
	}
	return false
}

// parseInvoiceSheet reads the key/value metadata sheet.
func parseInvoiceSheet(f *excelize.File) (invoice.Invoice, error) {
	rows, err := f.GetRows(invoiceSheetName)
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("failed to read rows: %w", err)
	}

	var (
		inv     invoice.Invoice
		seen    = make(map[string]bool)
		missing = []string{"invoice_number", "date", "customer"}
	)

	for i, row := range rows {
		rowNumber := i + 1
		if isBlankRow(row) {
			continue
		}
		if len(row) < 2 {
			return invoice.Invoice{}, fmt.Errorf("row %d: expected two columns (Field, Value), got %d", rowNumber, len(row))
		}

		key := normalizeKey(row[0])
		value := strings.TrimSpace(row[1])

		switch key {
		case "invoice_number":
			inv.Number = value
		case "date":
			if value == "" {
				continue
			}
			date, err := time.Parse(dateLayout, value)
			if err != nil {
				return invoice.Invoice{}, fmt.Errorf("invalid date %q on row %d: must be %s", value, rowNumber, dateLayout)
			}
			inv.Date = date
		case "customer":
			inv.Customer = value
		case "description":
			inv.Description = value
		default:
			return invoice.Invoice{}, fmt.Errorf("unknown field %q on row %d", row[0], rowNumber)
		}

		if value != "" {
			seen[key] = true
		}
	}

	var stillMissing []string
	for _, want := range missing {
		if !seen[want] {
			stillMissing = append(stillMissing, want)
		}
	}
	if len(stillMissing) > 0 {
		return invoice.Invoice{}, fmt.Errorf("missing field: %s", strings.Join(stillMissing, ", "))
	}
	return inv, nil
}

// parseItemsExcelSheet reads the "Items" sheet into domain items.
func parseItemsExcelSheet(f *excelize.File) ([]invoice.InvoiceItem, error) {
	rows, err := f.GetRows(itemsSheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("missing header row")
	}

	headers := rows[0]
	if err := ensureItemsHeaders(headers); err != nil {
		return nil, err
	}

	columns := make(map[string]int, len(headers))
	for i, h := range headers {
		if key := normalizeKey(h); key != "" {
			columns[key] = i
		}
	}

	items := make([]invoice.InvoiceItem, 0, len(rows)-1)
	for i, row := range rows[1:] {
		rowNumber := i + 2 // header is row 1
		if isBlankRow(row) {
			continue
		}
		cells := make(map[string]string, len(columns))
		for key, col := range columns {
			if col < len(row) {
				cells[key] = strings.TrimSpace(row[col])
			}
		}
		item, err := parseItemRow(cells, rowNumber)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}