package pdf

import (
	"strings"
	"testing"
	"time"

	"invoice-generator/internal/invoice"
)

// sampleInvoice mirrors examples/items.csv: one plain item and one that may
// carry a discount.
func sampleInvoice(discount float32) invoice.Invoice {
	items := []invoice.InvoiceItem{
		{
			TechnicalCode: "PRD-001",
			Description:   "Product A",
			Quantity:      2,
			UnitPrice:     500000,
		},
		{
			Description:     "Product B",
			Quantity:        3,
			UnitPrice:       250000,
			DiscountPercent: discount,
		},
	}

	return invoice.Invoice{
		Number:      "10023",
		Date:        time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
		Customer:    "شرکت نمونه",
		Description: "توضیحات نمونه",
		Items:       items,
	}
}

// renderSample prices the sample invoice and renders it to HTML.
func renderSample(t *testing.T, discount float32) string {
	t.Helper()

	inv := sampleInvoice(discount)
	invoice.ComputeItemPrices(inv.Items)
	total := invoice.CalculateTotal(inv)

	html, err := Render(NewView(inv, total))
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	return html
}

func TestRenderContainsInvoiceData(t *testing.T) {
	// Without a discount, item totals are 2 × 500,000 and 3 × 250,000.
	wants := []string{
		"10023",
		"2026-09-13",
		"شرکت نمونه",
		"Product A",
		"PRD-001",
		"1,000,000",
		"750,000",
		"1,750,000",
		"توضیحات نمونه",
	}

	html := renderSample(t, 0)

	for _, want := range wants {
		if !strings.Contains(html, want) {
			t.Errorf("rendered HTML does not contain %q", want)
		}
	}
}

func TestRenderTotalsUseFormattedMoney(t *testing.T) {
	html := renderSample(t, 0)

	// An unformatted amount means the template bypassed the money helper.
	if strings.Contains(html, "1750000") {
		t.Error("rendered HTML contains the unformatted total 1750000")
	}
	if !strings.Contains(html, "جمع کل") {
		t.Error("rendered HTML is missing the grand total label")
	}
}

func TestRenderDiscountColumn(t *testing.T) {
	t.Run("shown when an item is discounted", func(t *testing.T) {
		html := renderSample(t, 10)

		if !strings.Contains(html, "تخفیف") {
			t.Error("expected the discount column for a discounted item")
		}
		if !strings.Contains(html, "10٪") {
			t.Error("expected the discount percentage to be rendered")
		}
	})

	t.Run("hidden when nothing is discounted", func(t *testing.T) {
		html := renderSample(t, 0)

		if strings.Contains(html, "تخفیف") {
			t.Error("did not expect a discount column without any discount")
		}
	})
}

func TestRenderEscapesUntrustedValues(t *testing.T) {
	inv := sampleInvoice(0)
	inv.Customer = "<b>alert</b>"

	html, err := Render(NewView(inv, 1750000))
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if strings.Contains(html, "<b>alert</b>") {
		t.Error("rendered HTML did not escape the customer name")
	}
	if !strings.Contains(html, "&lt;b&gt;alert&lt;/b&gt;") {
		t.Error("expected the escaped customer name in the output")
	}
}

func TestRenderEmptyInvoice(t *testing.T) {
	var inv invoice.Invoice

	html, err := Render(NewView(inv, 0))
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if !strings.Contains(html, "جمع کل") {
		t.Error("expected the totals block even for an empty invoice")
	}
}
