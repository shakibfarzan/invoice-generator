# Example input files

These files are the reference for the input formats the tool accepts
(documented in full in `internal/input/doc.go`). The test suite parses both
of them, so they always stay in sync with the parsers.

## invoice.xlsx

The most complete input format. Two sheets:

- **Invoice** — metadata as key/value rows: `invoice_number`, `date`
  (YYYY-MM-DD), `customer`, and optional `description`.
- **Items** — one item per row with columns `technical_code` (optional),
  `description`, `quantity`, `unit_price`, `discount_percent` (optional).

## items.csv

The item list only (comma-separated, UTF-8, same columns as the Items
sheet). Invoice metadata is intended to come from CLI flags or a config
file later.