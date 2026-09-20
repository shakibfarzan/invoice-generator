// Package config holds optional seller/company configuration that is shared
// across every invoice (roadmap steps 22, 23 and 24).
//
// It is deliberately separate from the invoice domain: a single company's
// details (name, contact, logo) should not be repeated inside every
// spreadsheet. The CLI loads this once and passes the result to the PDF
// renderer.
//
// Configuration is read from a small YAML subset (see Parse). The file is
// optional — when it is missing the generator simply produces an invoice
// without company branding, which keeps the document valid (roadmap step 24
// requires the invoice to work without a logo, and therefore without a
// company block as well).
package config

import (
	"fmt"
	"os"
	"strings"
)

// Company describes the seller. None of the fields are required; an absent
// company is fine and the invoice stays valid.
type Company struct {
	Name    string // legal / trading name shown on the document
	Logo    string // path to a logo image (png, jpg, svg, ...); optional
	Phone   string
	Email   string
	Address string
	Website string
	TaxID   string // economic / VAT registration number
}

// Config is the top-level configuration document.
type Config struct {
	Company Company
}

// Load reads configuration from path.
//
// A missing file is not an error: it yields a zero Config so callers can
// still generate an invoice, just without company branding. Any other read
// error, or a structurally broken file, is reported.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("failed to read config %q: %w", path, err)
	}
	return Parse(string(data))
}

// Parse parses configuration from its textual form (a limited YAML subset).
//
// Supported grammar (only what the project currently needs):
//
//	# comment
//	section:
//	  key: value
//	  key: "quoted value"   # inline comment
//
// Nested scalars under the recognised "company" (and, reserved for step 22,
// "invoice") sections are read; everything else is ignored. Lists, anchors,
// multiline strings and flow style are intentionally unsupported to avoid
// pulling in a YAML dependency for a handful of fields.
func Parse(s string) (Config, error) {
	var cfg Config
	var section string

	lines := strings.Split(s, "\n")
	for i, raw := range lines {
		lineNo := i + 1

		// Drop a trailing carriage return so Windows line endings work.
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		colon := strings.Index(trimmed, ":")
		if colon < 0 {
			return cfg, fmt.Errorf("config line %d: expected 'key: value', got %q", lineNo, trimmed)
		}

		key := strings.TrimSpace(trimmed[:colon])
		val := strings.TrimSpace(trimmed[colon+1:])

		// Indentation decides whether this line starts a section or is a
		// key nested under the current section. Spaces and tabs both count
		// as indentation so the loader accepts either style.
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		if indent == 0 {
			if val == "" {
				section = key
			} else {
				// A top-level scalar key we do not model yet; stop
				// assigning nested keys until the next section.
				section = ""
			}
			continue
		}

		// Nested key: strip an inline comment, then surrounding quotes.
		val = stripComment(val)
		val = unquote(val)

		switch section {
		case "company":
			setCompanyField(&cfg.Company, key, val)
		default:
			// Unknown sections (e.g. "invoice" from step 22) are ignored
			// until those steps are implemented.
		}
	}

	return cfg, nil
}

// setCompanyField maps a YAML key onto the matching Company field.
func setCompanyField(c *Company, key, val string) {
	switch key {
	case "name":
		c.Name = val
	case "logo":
		c.Logo = val
	case "phone":
		c.Phone = val
	case "email":
		c.Email = val
	case "address":
		c.Address = val
	case "website", "url":
		c.Website = val
	case "tax_id", "taxid", "vat":
		c.TaxID = val
	}
}

// stripComment removes an inline comment (a '#' preceded by whitespace and
// not inside quotes) from a value.
func stripComment(s string) string {
	inQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\'' || c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == '#' && !inQuote {
			if i == 0 || s[i-1] == ' ' || s[i-1] == '\t' {
				return strings.TrimRight(s[:i], " \t")
			}
		}
	}
	return s
}

// unquote removes one layer of surrounding single or double quotes.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
