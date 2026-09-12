package input

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"invoice-generator/internal/invoice"
)

const validCSV = `technical_code,description,quantity,unit_price,discount_percent
PRD-001,Product A,2,500000,
PRD-002,Product B,3,250000,10
,Product C,1,1000000,
`

func TestParseItemsCSVValid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []invoice.InvoiceItem
	}{
		{
			name:  "full rows",
			input: validCSV,
			want: []invoice.InvoiceItem{
				{TechnicalCode: "PRD-001", Description: "Product A", Quantity: 2, UnitPrice: 500000},
				{TechnicalCode: "PRD-002", Description: "Product B", Quantity: 3, UnitPrice: 250000, DiscountPercent: 10},
				{Description: "Product C", Quantity: 1, UnitPrice: 1000000},
			},
		},
		{
			name:  "required columns only",
			input: "description,quantity,unit_price\nWidget,1,42\n",
			want: []invoice.InvoiceItem{
				{Description: "Widget", Quantity: 1, UnitPrice: 42},
			},
		},
		{
			name:  "column order does not matter",
			input: "unit_price,quantity,description\n42,1,Widget\n",
			want: []invoice.InvoiceItem{
				{Description: "Widget", Quantity: 1, UnitPrice: 42},
			},
		},
		{
			name:  "headers are case-insensitive",
			input: "Description,QUANTITY,Unit_Price\nWidget,1,42\n",
			want: []invoice.InvoiceItem{
				{Description: "Widget", Quantity: 1, UnitPrice: 42},
			},
		},
		{
			name:  "surrounding whitespace is trimmed",
			input: "description, quantity ,unit_price\n Widget , 1 , 42 \n",
			want: []invoice.InvoiceItem{
				{Description: "Widget", Quantity: 1, UnitPrice: 42},
			},
		},
		{
			name:  "quoted values and commas inside text",
			input: "description,quantity,unit_price\n\"Widget, large\",1,42\n",
			want: []invoice.InvoiceItem{
				{Description: "Widget, large", Quantity: 1, UnitPrice: 42},
			},
		},
		{
			name:  "fractional discount percent",
			input: "description,quantity,unit_price,discount_percent\nWidget,1,42,12.5\n",
			want: []invoice.InvoiceItem{
				{Description: "Widget", Quantity: 1, UnitPrice: 42, DiscountPercent: 12.5},
			},
		},
		{
			name:  "blank data rows are skipped",
			input: "description,quantity,unit_price\n\nWidget,1,42\n\n , , \nGadget,2,7\n",
			want: []invoice.InvoiceItem{
				{Description: "Widget", Quantity: 1, UnitPrice: 42},
				{Description: "Gadget", Quantity: 2, UnitPrice: 7},
			},
		},
		{
			name:  "header only yields no items",
			input: "description,quantity,unit_price\n",
			want:  []invoice.InvoiceItem{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseItemsCSV(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("ParseItemsCSV() error = %v", err)
			}
			assertItems(t, got, tt.want)
		})
	}
}

func TestParseItemsCSVStripsBOM(t *testing.T) {
	got, err := ParseItemsCSV(strings.NewReader("\xEF\xBB\xBF" + validCSV))
	if err != nil {
		t.Fatalf("ParseItemsCSV() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("ParseItemsCSV() got %d items, want 3", len(got))
	}
}

func TestParseItemsCSVErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError string
	}{
		{
			name:      "empty input",
			input:     "",
			wantError: "missing header row",
		},
		{
			name:      "missing required header",
			input:     "description,quantity\nWidget,1\n",
			wantError: "missing header: unit_price",
		},
		{
			name:      "several missing headers reported together",
			input:     "description\nWidget\n",
			wantError: "missing header: quantity, unit_price",
		},
		{
			name:      "unknown column",
			input:     "description,quantity,unit_price,prize\nWidget,1,42,5\n",
			wantError: `unknown column: prize`,
		},
		{
			name:      "invalid quantity reports row number",
			input:     "description,quantity,unit_price\nWidget,abc,42\n",
			wantError: `row 2: quantity "abc" is not a valid number`,
		},
		{
			name:      "fractional quantity is invalid",
			input:     "description,quantity,unit_price\nWidget,1.5,42\n",
			wantError: `row 2: quantity "1.5" is not a valid number`,
		},
		{
			name:      "negative quantity is invalid",
			input:     "description,quantity,unit_price\nWidget,-2,42\n",
			wantError: `row 2: quantity "-2" is not a valid number`,
		},
		{
			name:      "invalid unit_price reports row number",
			input:     "description,quantity,unit_price\nA,1,42\nB,1,two\n",
			wantError: `row 3: unit_price "two" is not a valid number`,
		},
		{
			name:      "invalid discount_percent reports row number",
			input:     "description,quantity,unit_price,discount_percent\nWidget,1,42,10%\n",
			wantError: `row 2: discount_percent "10%" is not a valid number`,
		},
		{
			name:      "missing quantity value",
			input:     "description,quantity,unit_price\nWidget,,42\n",
			wantError: `row 2: quantity "" is not a valid number`,
		},
		{
			name:      "wrong number of fields",
			input:     "description,quantity,unit_price\nWidget,1\n",
			wantError: "malformed CSV",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseItemsCSV(strings.NewReader(tt.input))
			if err == nil {
				t.Fatalf("ParseItemsCSV() = nil error, want %q", tt.wantError)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("ParseItemsCSV() error = %q, want it to contain %q", err.Error(), tt.wantError)
			}
		})
	}
}

func TestParseItemsCSVFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.csv")
	content := "\xEF\xBB\xBF" + "description,quantity,unit_price\nWidget,1,42\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	got, err := ParseItemsCSVFile(path)
	if err != nil {
		t.Fatalf("ParseItemsCSVFile() error = %v", err)
	}
	if len(got) != 1 || got[0].Description != "Widget" {
		t.Fatalf("ParseItemsCSVFile() = %+v, want one Widget item", got)
	}
}

func TestParseItemsCSVFileMissing(t *testing.T) {
	_, err := ParseItemsCSVFile(filepath.Join(t.TempDir(), "nope.csv"))
	if err == nil {
		t.Fatal("ParseItemsCSVFile() = nil error for missing file")
	}
	if !strings.Contains(err.Error(), "failed to open CSV file") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "failed to open CSV file")
	}
}

func assertItems(t *testing.T, got, want []invoice.InvoiceItem) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d items, want %d: %+v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("item %d = %+v, want %+v", i+1, got[i], want[i])
		}
	}
}
