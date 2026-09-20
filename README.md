# Invoice Generator

A small, local-first Go CLI that turns invoice data from an Excel/CSV file into
a professional, print-ready **PDF invoice** with Persian (RTL) support.

```text
CSV / Excel / CLI
        │
        ▼
   Input Parser  ──►  Validation  ──►  Calculation  ──►  HTML Template  ──►  Headless Chromium  ──►  PDF
```

The project is intentionally small: it generates invoices and gets out of your
way. There is no database, web server, accounts, or payments — those belong to a
future accounting application (see the roadmap file).

---

## Features

- **Excel (`.xlsx`) and CSV input** — the same domain model regardless of source.
- **Validation** — required fields and sane numeric ranges are checked before
  anything is rendered.
- **Deterministic money math** — amounts are stored as integers (smallest
  currency unit) so there are no floating-point rounding surprises.
- **Professional PDF** — A4, RTL layout, embedded Persian font (Vazirmatn).
- **Company branding** — an optional `config.yaml` supplies the seller's name,
  contact details and logo (roadmap steps 23 & 24).
- **Safe output names** — the PDF is written as `invoice-<number>.pdf` with the
  number sanitized so it can never escape the output directory.

---

## Requirements

- **Go 1.27 or newer** (see `go.mod`).
- **Chromium or Google Chrome** — PDF generation drives a headless browser via
  [`chromedp`](https://github.com/chromedp/chromedp). The browser is discovered
  on your system; on Windows you may need to point Chromium at an executable.

---

## Build

The real entry point lives in `cmd/invoice/` (the root `main.go` is only a
placeholder that prints these instructions).

```bash
go build -o invoice ./cmd/invoice
```

---

## Usage

```bash
./invoice generate <file> [flags]
```

`generate` reads the file, validates it, computes totals, and writes a PDF.

### Input formats

**Excel (`.xlsx`)** — the most complete format. Two sheets:

| Sheet     | Shape                                                     |
|-----------|----------------------------------------------------------|
| `Invoice` | A two-column `Field, Value` table (one metadata row each). |
| `Items`   | One row per line item with a header row.                  |

Recognized `Invoice` fields: `invoice_number`, `date` (YYYY-MM-DD),
`customer`, `description` (optional).

Recognized `Items` columns: `description` (required), `quantity` (required),
`unit_price` (required), `technical_code` (optional), `discount_percent`
(optional, 0–100).

**CSV (`.csv`)** — carries the item table only. Supply the missing invoice
metadata with flags:

```bash
./invoice generate items.csv --number 10023 --date 2026-09-12 --customer "Company ABC"
```

### Flags

| Flag                  | Default        | Description                                                      |
|-----------------------|----------------|------------------------------------------------------------------|
| `--number`            | `""`           | Invoice number (required for CSV input).                         |
| `--date`              | `""`           | Invoice date, `YYYY-MM-DD` (required for CSV input).             |
| `--customer`          | `""`           | Customer name (required for CSV input).                          |
| `--output`, `-o`      | `output`       | Directory where the PDF is written.                              |
| `--config`            | `config.yaml`  | Path to the optional company config file (branding/logo).        |

### Example

```bash
./invoice generate examples/invoice.xlsx
# ▸ output/invoice-10023.pdf
```

If the invoice number is `10023`, the output is `output/invoice-10023.pdf`.

---

## Company branding & logo (`config.yaml`)

Company details belong in configuration, not in every spreadsheet. Drop a
`config.yaml` next to where you run the CLI (or pass `--config <path>`):

```yaml
# All fields are optional. Omit or blank anything you don't want printed.
company:
  name: "شرکت فاکتورساز نمونه"
  logo: ./assets/logo.png          # png, jpg, svg, gif or webp
  phone:   "+98 21 1234 5678"
  email:   info@example.com
  address: "تهران، ایران"
  website: "https://example.com"
  tax_id:  "1234567890"            # national / VAT registration number

# Reserved for a later step (currency, tax rate, output dir). Ignored for now.
invoice:
  currency: IRR
  tax_rate: 10
  output_dir: ./output
```

- The **logo** is embedded into the PDF as an inline data URI, so the document
  is self-contained and renders identically everywhere.
- If the logo file is missing or unreadable, **no logo space is rendered** — the
  invoice stays valid and just omits the image.
- If `config.yaml` does not exist at all, the generator still produces a valid,
  unbranded invoice.

---

## Project layout

```text
invoice-generator/
├── cmd/invoice/main.go        # CLI entry point (the `generate` command)
├── internal/
│   ├── config/                # Company configuration (steps 22–24)
│   ├── input/                 # CSV + Excel parsers
│   ├── invoice/               # Domain model, validation, calculation
│   └── pdf/                   # HTML rendering + Chromium PDF generation
├── templates/
│   └── invoice.html           # The invoice template (RTL, Persian)
├── assets/logo.png            # Sample company logo
├── config.yaml                # Sample company configuration
├── examples/
│   ├── invoice.xlsx           # Reference Excel input
│   └── items.csv              # Reference CSV input (items only)
├── go.mod
└── invoice-generator-roadmap.md
```

The `invoice` domain package deliberately knows nothing about Excel, CSV, the
CLI, or Chromium, so it can later be lifted into a larger accounting backend.

---

## Development

```bash
go test ./...        # run the test suite
go vet ./...         # static checks
gofmt -l .           # list files needing formatting (use `gofmt -w .` to fix)
```

The config file is parsed with a tiny, dependency-free YAML subset tailored to
the fields above, so no third-party YAML library is required.

---

## Status & roadmap

Implemented:

- Invoice domain model, money-safe calculations, validation.
- Excel + CSV input parsers.
- `generate` command and PDF output (RTL / Persian, embedded font).
- Company configuration and logo support (`config.yaml`, `--config`).

Planned (see `invoice-generator-roadmap.md`):

- `validate`, `preview`, and `batch` commands.
- Richer configuration (default currency, tax rate, output directory).
- Verbose logging and a broader testing strategy.
