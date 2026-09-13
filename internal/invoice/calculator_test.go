package invoice

import "testing"

func TestComputeItemPrices(t *testing.T) {
	tests := []struct {
		name              string
		items             []InvoiceItem
		wantFinalUnitPrice []int64
		wantPrice         []int64
	}{
		{
			name:              "empty items",
			items:             nil,
			wantFinalUnitPrice: nil,
			wantPrice:         nil,
		},
		{
			name: "no discount",
			items: []InvoiceItem{
				{Description: "A", Quantity: 2, UnitPrice: 500000, DiscountPercent: 0},
			},
			wantFinalUnitPrice: []int64{500000},
			wantPrice:         []int64{1000000},
		},
		{
			name: "with 10 percent discount",
			items: []InvoiceItem{
				{Description: "A", Quantity: 3, UnitPrice: 250000, DiscountPercent: 10},
			},
			wantFinalUnitPrice: []int64{225000},
			wantPrice:         []int64{675000},
		},
		{
			name: "full 100 percent discount",
			items: []InvoiceItem{
				{Description: "Free", Quantity: 5, UnitPrice: 100000, DiscountPercent: 100},
			},
			wantFinalUnitPrice: []int64{0},
			wantPrice:         []int64{0},
		},
		{
			name: "multiple items mixed",
			items: []InvoiceItem{
				{Description: "A", Quantity: 2, UnitPrice: 500000, DiscountPercent: 0},
				{Description: "B", Quantity: 3, UnitPrice: 250000, DiscountPercent: 10},
				{Description: "C", Quantity: 1, UnitPrice: 1000000, DiscountPercent: 0},
			},
			wantFinalUnitPrice: []int64{500000, 225000, 1000000},
			wantPrice:         []int64{1000000, 675000, 1000000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ComputeItemPrices(tt.items)

			for i, item := range tt.items {
				if item.FinalUnitPrice != tt.wantFinalUnitPrice[i] {
					t.Errorf("item %d: FinalUnitPrice = %d, want %d",
						i, item.FinalUnitPrice, tt.wantFinalUnitPrice[i])
				}
				if item.Price != tt.wantPrice[i] {
					t.Errorf("item %d: Price = %d, want %d",
						i, item.Price, tt.wantPrice[i])
				}
			}
		})
	}
}

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
