package invoice

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func validInvoice() Invoice {
	return Invoice{
		Number:   "10023",
		Date:     time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC),
		Customer: "Company ABC",
		Items: []InvoiceItem{
			{
				Description:     "Product A",
				Quantity:        2,
				UnitPrice:       500000,
				DiscountPercent: 0,
			},
		},
	}
}

func TestValidateValidInvoice(t *testing.T) {
	tests := []struct {
		name    string
		invoice Invoice
	}{
		{
			name:    "complete invoice",
			invoice: validInvoice(),
		},
		{
			name: "multiple items",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items = append(invoice.Items,
					InvoiceItem{
						Description:     "Product B",
						Quantity:        3,
						UnitPrice:       250000,
						DiscountPercent: 10,
					},
					InvoiceItem{
						Description: "Product C",
						Quantity:    1,
						UnitPrice:   1000000,
					},
				)
				return invoice
			}(),
		},
		{
			name: "zero unit price is allowed (free item)",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items[0].UnitPrice = 0
				return invoice
			}(),
		},
		{
			name: "full 100 percent discount is allowed",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items[0].DiscountPercent = 100
				return invoice
			}(),
		},
		{
			name: "description with surrounding whitespace is accepted",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Customer = "  Company ABC  "
				invoice.Items[0].Description = " Product A "
				return invoice
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.invoice); err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestValidateInvalidInvoice(t *testing.T) {
	tests := []struct {
		name    string
		invoice Invoice
		want    []string
	}{
		{
			name:    "empty invoice",
			invoice: Invoice{},
			want: []string{
				"invoice number is required",
				"invoice date is required",
				"customer name is required",
				"invoice must contain at least one item",
			},
		},
		{
			name: "missing invoice number",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Number = ""
				return invoice
			}(),
			want: []string{"invoice number is required"},
		},
		{
			name: "whitespace-only invoice number",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Number = "   "
				return invoice
			}(),
			want: []string{"invoice number is required"},
		},
		{
			name: "zero date",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Date = time.Time{}
				return invoice
			}(),
			want: []string{"invoice date is required"},
		},
		{
			name: "missing customer",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Customer = ""
				return invoice
			}(),
			want: []string{"customer name is required"},
		},
		{
			name: "whitespace-only customer",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Customer = "\t "
				return invoice
			}(),
			want: []string{"customer name is required"},
		},
		{
			name: "nil items",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items = nil
				return invoice
			}(),
			want: []string{"invoice must contain at least one item"},
		},
		{
			name: "item with empty description",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items[0].Description = ""
				return invoice
			}(),
			want: []string{"item 1: description is required"},
		},
		{
			name: "item with zero quantity",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items[0].Quantity = 0
				return invoice
			}(),
			want: []string{"item 1: quantity must be greater than zero"},
		},
		{
			name: "item with negative unit price",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items[0].UnitPrice = -1
				return invoice
			}(),
			want: []string{"item 1: unit price must not be negative"},
		},
		{
			name: "item with negative discount percent",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items[0].DiscountPercent = -0.5
				return invoice
			}(),
			want: []string{"item 1: discount percent must be between 0 and 100"},
		},
		{
			name: "item with discount percent above 100",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items[0].DiscountPercent = 100.5
				return invoice
			}(),
			want: []string{"item 1: discount percent must be between 0 and 100"},
		},
		{
			name: "second item errors report item 2",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items = append(invoice.Items,
					InvoiceItem{Description: "", Quantity: 0, UnitPrice: -5},
				)
				return invoice
			}(),
			want: []string{
				"item 2: description is required",
				"item 2: quantity must be greater than zero",
				"item 2: unit price must not be negative",
			},
		},
		{
			name: "errors from several items are all reported in order",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Items[0].Quantity = 0
				invoice.Items = append(invoice.Items,
					InvoiceItem{Description: "B", Quantity: 1, UnitPrice: 100},
					InvoiceItem{Description: "", Quantity: 2, UnitPrice: 200},
				)
				return invoice
			}(),
			want: []string{
				"item 1: quantity must be greater than zero",
				"item 3: description is required",
			},
		},
		{
			name: "multiple problems on invoice and items are aggregated",
			invoice: func() Invoice {
				invoice := validInvoice()
				invoice.Number = ""
				invoice.Customer = ""
				invoice.Items[0].Quantity = 0
				return invoice
			}(),
			want: []string{
				"invoice number is required",
				"customer name is required",
				"item 1: quantity must be greater than zero",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.invoice)
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %v", tt.want)
			}

			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("Validate() error is %T, want *ValidationError", err)
			}

			if got := validationErr.Errors; !equalStrings(got, tt.want) {
				t.Errorf("Validate() errors = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationErrorMessages(t *testing.T) {
	err := &ValidationError{Errors: []string{
		"customer name is required",
		"item 1: quantity must be greater than zero",
	}}

	msg := err.Error()

	if !strings.HasPrefix(msg, "invoice validation failed") {
		t.Errorf("Error() = %q, want prefix %q", msg, "invoice validation failed")
	}
	for _, want := range err.Errors {
		if !strings.Contains(msg, want) {
			t.Errorf("Error() = %q, want it to contain %q", msg, want)
		}
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
