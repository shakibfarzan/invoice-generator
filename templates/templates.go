// Package templates embeds the HTML templates into the binary so the built
// CLI does not need the repository checked out next to it at runtime.
package templates

import _ "embed"

// InvoiceHTML is the raw source of the invoice template.
// It is parsed by internal/pdf, which owns the template functions.
//
//go:embed invoice.html
var InvoiceHTML string
