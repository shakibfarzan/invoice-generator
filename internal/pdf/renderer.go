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
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"invoice-generator/internal/config"
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

	// Company is the seller's branding, loaded from configuration
	// (roadmap steps 23 and 24). It may be zero when no config file is
	// present, in which case the template falls back to a placeholder.
	Company config.Company

	// Logo is the company logo rendered as an inline data: URI, ready to
	// drop straight into an <img src>. It is empty when no logo is
	// configured or the file cannot be read, so the template can skip it
	// without leaving blank space (roadmap step 24).
	Logo string

	// HasDiscount tells the template whether to render the discount
	// column. Persian invoices omit it entirely when nothing is
	// discounted, which keeps the table narrow.
	HasDiscount bool

	FontFaceCSS template.CSS
}

// NewView builds the template data for an invoice whose items have already
// been priced and whose total has already been calculated.
//
// company carries the seller's branding (name, contact, logo). A zero value
// is perfectly valid and simply produces an unbranded invoice.
func NewView(inv invoice.Invoice, total int64, company config.Company) View {
	view := View{
		Invoice:     inv,
		Items:       inv.Items,
		Total:       total,
		Company:     company,
		Logo:        resolveLogo(company.Logo),
		FontFaceCSS: template.CSS(FontFaceCSS()),
	}

	if !inv.Date.IsZero() {
		view.Date = ToJalali(inv.Date)
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
			"money": func(v int64) string { return ToPersianDigits(FormatMoney(v)) },
			"pct":   func(v float32) string { return ToPersianDigits(FormatPercent(v)) },
			"inc":   func(i int) string { return ToPersianDigits(strconv.Itoa(i + 1)) },
			"fa": ToPersianDigits,
		}).
		Parse(templates.InvoiceHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse invoice template: %w", err)
	}
	return tmpl, nil
}

// resolveLogo turns a logo file path into an inline data: URI so the PDF
// renderer does not depend on the filesystem layout Chromium sees. A missing
// or unreadable logo yields an empty string, which the template treats as
// "no logo" (roadmap step 24: the invoice stays valid without a logo).
//
// The path is interpreted relative to the process working directory, so a
// config that sets "logo: ./assets/logo.png" resolves from wherever the CLI
// is run. An absolute path is used as-is.
func resolveLogo(logoPath string) string {
	if logoPath == "" {
		return ""
	}

	data, err := os.ReadFile(logoPath)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("data:%s;base64,%s",
		logoMIME(logoPath), base64.StdEncoding.EncodeToString(data))
}

// logoMIME maps a logo file extension to its image MIME type.
func logoMIME(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}
