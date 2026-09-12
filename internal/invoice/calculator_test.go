package invoice

import "testing"

func TestCalculateTotal(t *testing.T) {
	tests := []struct {
		name     string
		invoice  Invoice
		expected int64
	}{
		{
			name:     "empty items",
			invoice:  Invoice{},
			expected: 0,
		},
		{
			name: "single item",
			invoice: Invoice{
				Items: []InvoiceItem{
					{Price: 100},
				},
			},
			expected: 100,
		},
		{
			name: "multiple items",
			invoice: Invoice{
				Items: []InvoiceItem{
					{Price: 100},
					{Price: 250},
					{Price: 650},
				},
			},
			expected: 1000,
		},
		{
			name: "zero priced items",
			invoice: Invoice{
				Items: []InvoiceItem{
					{Price: 0},
					{Price: 42},
				},
			},
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTotal(tt.invoice)
			if result != tt.expected {
				t.Errorf("CalculateTotal() = %d, want %d", result, tt.expected)
			}
		})
	}
}
