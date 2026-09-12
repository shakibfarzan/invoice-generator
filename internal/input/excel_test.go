package input

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"invoice-generator/internal/invoice"
)

// workbookBuilder assembles small .xlsx fixtures on disk for tests.
type workbookBuilder struct {
	t                *testing.T
	invoiceRows      [][]any
	itemsRows        [][]any
	invoiceSheet     bool
	itemsSheet       bool
	invoiceSheetName string
	itemsSheetName   string
}

func newWorkbook(t *testing.T) *workbookBuilder {
	t.Helper()
	return &workbookBuilder{
		t:                t,
		invoiceSheet:     true,
		itemsSheet:       true,
		invoiceSheetName: "Invoice",
		itemsSheetName:   "Items",
	}
}

func (b *workbookBuilder) withInvoiceRows(rows [][]any) *workbookBuilder {
	b.invoiceRows = rows
	return b
}

func (b *workbookBuilder) withItemsRows(rows [][]any) *workbookBuilder {
	b.itemsRows = rows
	return b
}

func (b *workbookBuilder) withoutInvoiceSheet() *workbookBuilder {
	b.invoiceSheet = false
	return b
}

func (b *workbookBuilder) withoutItemsSheet() *workbookBuilder {
	b.itemsSheet = false
	return b
}

func (b *workbookBuilder) withInvoiceSheetName(name string) *workbookBuilder {
	b.invoiceSheetName = name
	return b
}

func (b *workbookBuilder) save() string {
	b.t.Helper()

	f := excelize.NewFile()
	defer f.Close()

	defaultSheet := f.GetSheetName(f.GetActiveSheetIndex())
	if b.invoiceSheet {
		if defaultSheet != b.invoiceSheetName {
			b.mustSetSheetName(f, defaultSheet, b.invoiceSheetName)
		}
		if b.itemsSheet {
			if _, err := f.NewSheet(b.itemsSheetName); err != nil {
				b.t.Fatalf("failed to create sheet %s: %v", b.itemsSheetName, err)
			}
		}
	} else if b.itemsSheet {
		if defaultSheet != b.itemsSheetName {
			b.mustSetSheetName(f, defaultSheet, b.itemsSheetName)
		}
	}

	if b.invoiceSheet {
		writeSheet(b.t, f, b.invoiceSheetName, b.invoiceRows)
	}
	if b.itemsSheet {
		writeSheet(b.t, f, b.itemsSheetName, b.itemsRows)
	}

	path := filepath.Join(b.t.TempDir(), "invoice.xlsx")
	if err := f.SaveAs(path); err != nil {
		b.t.Fatalf("failed to save workbook: %v", err)
	}
	return path
}

func (b *workbookBuilder) mustSetSheetName(f *excelize.File, oldName, newName string) {
	if err := f.SetSheetName(oldName, newName); err != nil {
		b.t.Fatalf("failed to rename sheet %q to %q: %v", oldName, newName, err)
	}
}

func writeSheet(t *testing.T, f *excelize.File, sheet string, rows [][]any) {
	t.Helper()
	for i, row := range rows {
		for j, cell := range row {
			if cell == nil {
				continue
			}
			axis, err := excelize.CoordinatesToCellName(j+1, i+1)
			if err != nil {
				t.Fatalf("failed to build cell name: %v", err)
			}
			if err := f.SetCellValue(sheet, axis, cell); err != nil {
				t.Fatalf("failed to set cell %s: %v", axis, err)
			}
		}
	}
}

func defaultInvoiceRows() [][]any {
	return [][]any{
		{"Field", "Value"},
		{"invoice_number", "10023"},
		{"date", "2026-09-12"},
		{"customer", "Company ABC"},
		{"description", "Website development services"},
	}
}

func defaultItemsRows() [][]any {
	return [][]any{
		{"technical_code", "description", "quantity", "unit_price", "discount_percent"},
		{"PRD-001", "Product A", 2, "500000", nil},
		{nil, "Product B", 3, "250000", "10"},
		{nil, nil, nil, nil, nil}, // blank row must be skipped
		{"PRD-003", "Product C", 1, "1000000", nil},
	}
}

func wantDefaultInvoice(t *testing.T) invoice.Invoice {
	t.Helper()
	return invoice.Invoice{
		Number:      "10023",
		Date:        time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
		Customer:    "Company ABC",
		Description: "Website development services",
		Items: []invoice.InvoiceItem{
			{TechnicalCode: "PRD-001", Description: "Product A", Quantity: 2, UnitPrice: 500000},
			{Description: "Product B", Quantity: 3, UnitPrice: 250000, DiscountPercent: 10},
			{TechnicalCode: "PRD-003", Description: "Product C", Quantity: 1, UnitPrice: 1000000},
		},
	}
}

func assertInvoicesEqual(t *testing.T, got, want invoice.Invoice) {
	t.Helper()
	if got.Number != want.Number {
		t.Errorf("Number = %q, want %q", got.Number, want.Number)
	}
	if !got.Date.Equal(want.Date) {
		t.Errorf("Date = %v, want %v", got.Date, want.Date)
	}
	if got.Customer != want.Customer {
		t.Errorf("Customer = %q, want %q", got.Customer, want.Customer)
	}
	if got.Description != want.Description {
		t.Errorf("Description = %q, want %q", got.Description, want.Description)
	}
	assertItems(t, got.Items, want.Items)
}

func TestParseInvoiceExcelValid(t *testing.T) {
	tests := []struct {
		name string
		file string
		want invoice.Invoice
	}{
		{
			name: "canonical workbook",
			file: newWorkbook(t).
				withInvoiceRows(defaultInvoiceRows()).
				withItemsRows(defaultItemsRows()).
				save(),
			want: wantDefaultInvoice(t),
		},
		{
			name: "sheet names are case-insensitive",
			file: newWorkbook(t).
				withInvoiceSheetName("INVOICE").
				withInvoiceRows(defaultInvoiceRows()).
				withItemsRows(defaultItemsRows()).
				save(),
			want: wantDefaultInvoice(t),
		},
		{
			name: "numeric cells for price",
			file: newWorkbook(t).
				withInvoiceRows(defaultInvoiceRows()).
				withItemsRows([][]any{
					{"description", "quantity", "unit_price"},
					{"Product A", 2, 500000},
				}).
				save(),
			want: func() invoice.Invoice {
				inv := wantDefaultInvoice(t)
				inv.Description = "Website development services"
				inv.Items = []invoice.InvoiceItem{
					{Description: "Product A", Quantity: 2, UnitPrice: 500000},
				}
				return inv
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInvoiceExcel(tt.file)
			if err != nil {
				t.Fatalf("ParseInvoiceExcel() error = %v", err)
			}
			assertInvoicesEqual(t, got, tt.want)
		})
	}
}

func TestParseInvoiceExcelErrors(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		wantError string
	}{
		{
			name:      "missing Items sheet",
			file:      newWorkbook(t).withInvoiceRows(defaultInvoiceRows()).withoutItemsSheet().save(),
			wantError: "missing sheet: Items",
		},
		{
			name:      "missing Invoice sheet",
			file:      newWorkbook(t).withItemsRows(defaultItemsRows()).withoutInvoiceSheet().save(),
			wantError: "missing sheet: Invoice",
		},
		{
			name: "missing field",
			file: newWorkbook(t).
				withInvoiceRows([][]any{
					{"Field", "Value"},
					{"invoice_number", "10023"},
					{"date", "2026-09-12"},
				}).
				withItemsRows(defaultItemsRows()).
				save(),
			wantError: "missing field: customer",
		},
		{
			name: "several missing fields reported together",
			file: newWorkbook(t).
				withInvoiceRows([][]any{{"Field", "Value"}}).
				withItemsRows(defaultItemsRows()).
				save(),
			wantError: "missing field: invoice_number, date, customer",
		},
		{
			name: "blank value counts as missing field",
			file: newWorkbook(t).
				withInvoiceRows([][]any{
					{"Field", "Value"},
					{"invoice_number", ""},
					{"date", "2026-09-12"},
					{"customer", "Company ABC"},
				}).
				withItemsRows(defaultItemsRows()).
				save(),
			wantError: "missing field: invoice_number",
		},
		{
			name: "unknown field catches typos",
			file: newWorkbook(t).
				withInvoiceRows([][]any{
					{"Field", "Value"},
					{"invoice_number", "10023"},
					{"custmer", "Company ABC"},
					{"date", "2026-09-12"},
				}).
				withItemsRows(defaultItemsRows()).
				save(),
			wantError: `unknown field "custmer" on row 3`,
		},
		{
			name: "invalid date format",
			file: newWorkbook(t).
				withInvoiceRows([][]any{
					{"Field", "Value"},
					{"invoice_number", "10023"},
					{"date", "12/09/2026"},
					{"customer", "Company ABC"},
				}).
				withItemsRows(defaultItemsRows()).
				save(),
			wantError: `invalid date "12/09/2026" on row 3: must be 2006-01-02`,
		},
		{
			name: "single column row in Invoice sheet",
			file: newWorkbook(t).
				withInvoiceRows([][]any{
					{"Field"},
					{"invoice_number"},
				}).
				withItemsRows(defaultItemsRows()).
				save(),
			wantError: "expected two columns (Field, Value), got 1",
		},
		{
			name: "invalid quantity reports sheet and row",
			file: newWorkbook(t).
				withInvoiceRows(defaultInvoiceRows()).
				withItemsRows([][]any{
					{"description", "quantity", "unit_price"},
					{"Product A", 2, "500000"},
					{"Product B", "abc", "250000"},
				}).
				save(),
			wantError: `sheet "Items": row 3: quantity "abc" is not a valid number`,
		},
		{
			name: "missing item headers",
			file: newWorkbook(t).
				withInvoiceRows(defaultInvoiceRows()).
				withItemsRows([][]any{{"description", "quantity"}}).
				save(),
			wantError: `sheet "Items": missing header: unit_price`,
		},
		{
			name: "empty Items sheet",
			file: newWorkbook(t).
				withInvoiceRows(defaultInvoiceRows()).
				withItemsRows(nil).
				save(),
			wantError: `sheet "Items": missing header row`,
		},
		{
			name: "unknown item column",
			file: newWorkbook(t).
				withInvoiceRows(defaultInvoiceRows()).
				withItemsRows([][]any{
					{"description", "quantity", "unit_price", "prize"},
					{"Product A", 2, "500000", "5"},
				}).
				save(),
			wantError: `sheet "Items": unknown column: prize`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseInvoiceExcel(tt.file)
			if err == nil {
				t.Fatalf("ParseInvoiceExcel() = nil error, want %q", tt.wantError)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("ParseInvoiceExcel() error = %q, want it to contain %q", err.Error(), tt.wantError)
			}
		})
	}
}

func TestParseInvoiceExcelFileMissing(t *testing.T) {
	_, err := ParseInvoiceExcel(filepath.Join(t.TempDir(), "nope.xlsx"))
	if err == nil {
		t.Fatal("ParseInvoiceExcel() = nil error for missing file")
	}
	if !strings.Contains(err.Error(), "failed to open Excel file") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "failed to open Excel file")
	}
}

// TestExampleFiles parses the example files shipped in the repository, so
// the examples can never silently rot.
func TestExampleFiles(t *testing.T) {
	inv, err := ParseInvoiceExcel(filepath.Join("..", "..", "examples", "invoice.xlsx"))
	if err != nil {
		t.Fatalf("ParseInvoiceExcel(examples/invoice.xlsx) error = %v", err)
	}
	assertInvoicesEqual(t, inv, wantDefaultInvoice(t))

	items, err := ParseItemsCSVFile(filepath.Join("..", "..", "examples", "items.csv"))
	if err != nil {
		t.Fatalf("ParseItemsCSVFile(examples/items.csv) error = %v", err)
	}
	want := wantDefaultInvoice(t)
	assertItems(t, items, want.Items)
}

func TestExampleFilesExist(t *testing.T) {
	for _, name := range []string{"invoice.xlsx", "items.csv"} {
		path := filepath.Join("..", "..", "examples", name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("example file %s is missing", path)
		}
	}
}
