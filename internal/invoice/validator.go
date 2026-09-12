package invoice

import (
	"fmt"
	"strings"
)

// ValidationError holds every problem found while validating an invoice,
// so the caller can report all of them at once instead of one per run.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("invoice validation failed:")
	for _, msg := range e.Errors {
		b.WriteString("\n- ")
		b.WriteString(msg)
	}
	return b.String()
}

// Validate checks an invoice against the business rules and returns nil
// when the invoice is valid. Otherwise it returns a *ValidationError that
// contains every problem found, in a stable order.
//
// An invalid invoice must never reach the PDF generator.
func Validate(invoice Invoice) error {
	var errs []string

	if strings.TrimSpace(invoice.Number) == "" {
		errs = append(errs, "invoice number is required")
	}

	if invoice.Date.IsZero() {
		errs = append(errs, "invoice date is required")
	}

	if strings.TrimSpace(invoice.Customer) == "" {
		errs = append(errs, "customer name is required")
	}

	if len(invoice.Items) == 0 {
		errs = append(errs, "invoice must contain at least one item")
	}

	for i, item := range invoice.Items {
		errs = append(errs, validateItem(i+1, item)...)
	}

	if len(errs) == 0 {
		return nil
	}

	return &ValidationError{Errors: errs}
}

// validateItem checks a single item. index is 1-based so error messages
// match how rows are numbered in spreadsheets and CSV files.
func validateItem(index int, item InvoiceItem) []string {
	var errs []string

	if strings.TrimSpace(item.Description) == "" {
		errs = append(errs, fmt.Sprintf("item %d: description is required", index))
	}

	if item.Quantity == 0 {
		errs = append(errs, fmt.Sprintf("item %d: quantity must be greater than zero", index))
	}

	if item.UnitPrice < 0 {
		errs = append(errs, fmt.Sprintf("item %d: unit price must not be negative", index))
	}

	if item.DiscountPercent < 0 || item.DiscountPercent > 100 {
		errs = append(errs, fmt.Sprintf("item %d: discount percent must be between 0 and 100", index))
	}

	return errs
}
