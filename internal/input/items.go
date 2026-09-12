package input

import (
	"fmt"
	"strconv"
	"strings"

	"invoice-generator/internal/invoice"
)

var (
	itemsRequiredHeaders = []string{"description", "quantity", "unit_price"}
	itemsOptionalHeaders = []string{"technical_code", "discount_percent"}
)

// normalizeKey lowercases and trims a header or field name so that
// "Unit Price", "unit_price " and "UNIT_PRICE" all match.
func normalizeKey(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ensureItemsHeaders checks the header row of an item table (CSV or Excel).
// All required headers must be present and no unknown ones may appear.
func ensureItemsHeaders(headers []string) error {
	seen := make(map[string]bool, len(headers))
	var unknown []string
	for _, h := range headers {
		key := normalizeKey(h)
		if key == "" {
			continue
		}
		if !isKnownItemHeader(key) {
			unknown = append(unknown, key)
			continue
		}
		seen[key] = true
	}

	var missing []string
	for _, want := range itemsRequiredHeaders {
		if !seen[want] {
			missing = append(missing, want)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing header: %s", strings.Join(missing, ", "))
	}
	if len(unknown) > 0 {
		return fmt.Errorf("unknown column: %s", strings.Join(unknown, ", "))
	}
	return nil
}

func isKnownItemHeader(key string) bool {
	for _, h := range itemsRequiredHeaders {
		if key == h {
			return true
		}
	}
	for _, h := range itemsOptionalHeaders {
		if key == h {
			return true
		}
	}
	return false
}

// parseItemRow converts one item row, given as column-name to raw-cell-value
// map, into a domain item. rowNumber is the 1-based row in the source file
// (including the header row) and is used in error messages.
//
// It parses only; range checks such as quantity > 0 or discount <= 100 are
// the validator's job.
func parseItemRow(cells map[string]string, rowNumber int) (invoice.InvoiceItem, error) {
	item := invoice.InvoiceItem{
		TechnicalCode: strings.TrimSpace(cells["technical_code"]),
		Description:   strings.TrimSpace(cells["description"]),
	}

	rawQuantity := strings.TrimSpace(cells["quantity"])
	quantity, err := strconv.ParseUint(rawQuantity, 10, 64)
	if err != nil {
		return invoice.InvoiceItem{}, fmt.Errorf("row %d: quantity %q is not a valid number", rowNumber, rawQuantity)
	}
	item.Quantity = uint(quantity)

	rawUnitPrice := strings.TrimSpace(cells["unit_price"])
	unitPrice, err := strconv.ParseInt(rawUnitPrice, 10, 64)
	if err != nil {
		return invoice.InvoiceItem{}, fmt.Errorf("row %d: unit_price %q is not a valid number", rowNumber, rawUnitPrice)
	}
	item.UnitPrice = unitPrice

	rawDiscount := strings.TrimSpace(cells["discount_percent"])
	if rawDiscount != "" {
		discount, err := strconv.ParseFloat(rawDiscount, 32)
		if err != nil {
			return invoice.InvoiceItem{}, fmt.Errorf("row %d: discount_percent %q is not a valid number", rowNumber, rawDiscount)
		}
		item.DiscountPercent = float32(discount)
	}

	return item, nil
}

func isBlankRow(cells []string) bool {
	for _, c := range cells {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}
