// Package pdf turns an invoice into a PDF document.
//
// The pipeline mirrors roadmap steps 11 and 15:
//
//	Invoice ──► Go HTML template ──► HTML ──► headless Chromium ──► PDF
//
// Rendering (this file) and printing (generator.go) are kept apart so the
// HTML can be tested without launching a browser.
package pdf

import (
	"fmt"
	"html/template"
	"io"
	"strings"

	"invoice-generator/internal/invoice"
	"invoice-generator/templates"
)

const (
	// dateLayout is the rendering format for the invoice date. The domain
	// stores a time.Time; the template only ever sees a string.
	dateLayout = "2006-01-02"

	// templateName names the parsed template for error messages.
	templateName = "invoice"
)

// View is the complete set of data the HTML template may read.
//
// It exists so the template never reaches into the domain for anything that
// is not already computed: totals arrive as a single int64 and dates arrive
// as strings.
type View struct {
	// Invoice carries the metadata (number, customer, notes).
	Invoice invoice.Invoice

	// Items is the line items, with FinalUnitPrice and Price already
	// computed by invoice.ComputeItemPrices.
	Items []invoice.InvoiceItem

	// Total is the grand total, computed by invoice.CalculateTotal.
	Total int64

	// Date is Invoice.Date pre-formatted, because formatting in the
	// template would hide the layout choice deep inside the HTML.
	Date string

	// Seller is the name shown in the "seller" box. It is empty until
	// company information is configurable (roadmap step 23).
	Seller string

	// HasDiscount tells the template whether to render the discount
	// column. Persian invoices omit it entirely when nothing is
	// discounted, which keeps the table narrow.
	HasDiscount bool
}

// NewView builds the template data for an invoice whose items have already
// been priced and whose total has already been calculated.
func NewView(inv invoice.Invoice, total int64) View {
	view := View{
		Invoice: inv,
		Items:   inv.Items,
		Total:   total,
	}

	if !inv.Date.IsZero() {
		view.Date = inv.Date.Format(dateLayout)
	}

	for _, item := range inv.Items {
		if item.DiscountPercent > 0 {
			view.HasDiscount = true
			break
		}
	}

	return view
}

// Render returns the full HTML document for the view.
func Render(v View) (string, error) {
	var b strings.Builder
	if err := RenderTo(&b, v); err != nil {
		return "", err
	}
	return b.String(), nil
}

// RenderTo writes the rendered HTML document to w.
func RenderTo(w io.Writer, v View) error {
	tmpl, err := ParseTemplate()
	if err != nil {
		return err
	}

	if err := tmpl.Execute(w, v); err != nil {
		return fmt.Errorf("failed to render invoice template: %w", err)
	}
	return nil
}

// ParseTemplate parses the embedded invoice template together with the
// functions the template is allowed to call.
func ParseTemplate() (*template.Template, error) {
	// Template functions, in alphabetical order:
	//
	//	money renders an amount with thousands separators
	//	pct   renders a discount percentage
	//	inc   turns a zero-based range index into a row number
	tmpl, err := template.New(templateName).
		Funcs(template.FuncMap{
			"money": FormatMoney,
			"pct":   FormatPercent,
			"inc":   func(i int) int { return i + 1 },
		}).
		Parse(templates.InvoiceHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse invoice template: %w", err)
	}
	return tmpl, nil
}
