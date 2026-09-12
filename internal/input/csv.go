package input

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"invoice-generator/internal/invoice"
)

// ParseItemsCSVFile reads a CSV file from disk and parses it into invoice
// items. See ParseItemsCSV for the expected format.
func ParseItemsCSVFile(path string) ([]invoice.InvoiceItem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer f.Close()

	items, err := ParseItemsCSV(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV file %s: %w", path, err)
	}
	return items, nil
}

// ParseItemsCSV parses item rows from CSV data.
//
// The first record must be a header row containing at least description,
// quantity and unit_price (see the package documentation for details).
// Subsequent records become items; fully empty records are skipped.
// A UTF-8 byte order mark is tolerated, because that is what spreadsheets
// commonly emit.
func ParseItemsCSV(r io.Reader) ([]invoice.InvoiceItem, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV data: %w", err)
	}
	data = bytes.TrimPrefix(data, []byte("\xEF\xBB\xBF"))

	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("malformed CSV: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("missing header row")
	}

	headers := records[0]
	if err := ensureItemsHeaders(headers); err != nil {
		return nil, err
	}

	columns := make(map[string]int, len(headers))
	for i, h := range headers {
		if key := normalizeKey(h); key != "" {
			columns[key] = i
		}
	}

	items := make([]invoice.InvoiceItem, 0, len(records)-1)
	for i, record := range records[1:] {
		rowNumber := i + 2 // header is row 1
		if isBlankRow(record) {
			continue
		}
		cells := make(map[string]string, len(columns))
		for key, col := range columns {
			if col < len(record) {
				cells[key] = strings.TrimSpace(record[col])
			}
		}
		item, err := parseItemRow(cells, rowNumber)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}