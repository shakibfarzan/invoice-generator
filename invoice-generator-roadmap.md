# Invoice Generator Service

A small, local-first Go service/CLI that takes invoice data from CSV/Excel or CLI arguments and generates a professional PDF invoice.

The project is intentionally small. It should **not** start as a full accounting application.

---

## 0. Project Goal

The final workflow should look like:

```text
CSV / Excel / CLI
       |
       v
   Input Parser
       |
       v
 Validation + Normalization
       |
       v
 Invoice Domain Model
       |
       v
 Calculations
       |
       v
 HTML Template
       |
       v
 Headless Chromium
       |
       v
      PDF
```

Example:

```bash
invoice generate invoice.xlsx
```

Output:

```text
output/
└── invoice-10023.pdf
```

The project should eventually support:

```bash
invoice generate invoice.xlsx
invoice validate invoice.xlsx
invoice preview invoice.xlsx
invoice batch ./invoices/
```

---

# 1. Scope

## 1.1 Version 1

Build only:

- Go CLI
- CSV input
- Excel input
- Invoice validation
- Invoice calculations
- HTML invoice template
- PDF generation
- Configurable output directory
- Persian/English text support
- RTL invoice support
- Basic logging and useful errors
- Unit tests for business logic

## 1.2 Explicitly NOT in Version 1

Do not build:

- User accounts
- Authentication
- Web UI
- REST API
- Database
- Cloud deployment
- Payments
- Inventory
- Accounting ledger
- Expense tracking
- Reporting dashboard
- Multi-user support
- Microservices

The point is to make a useful tool quickly.

---

# 2. Recommended Stack

| Purpose | Technology |
|---|---|
| Language | Go |
| CLI | Cobra |
| Excel | `excelize` |
| CSV | Go standard library |
| HTML | `html/template` |
| PDF | Headless Chromium |
| Browser automation | `chromedp` |
| Testing | Go `testing` |
| Formatting | `gofmt` |
| Static checks | `go vet` |
| Dependency management | Go Modules |

Suggested libraries:

```bash
go get github.com/spf13/cobra
go get github.com/xuri/excelize/v2
go get github.com/chromedp/chromedp
```

Use the standard library whenever it is sufficient.

---

# 3. Prerequisites

Install:

- Go
- Git
- Chromium or Google Chrome

Verify:

```bash
go version
git --version
```

Check Chromium/Chrome is available.

On Windows, you may need to configure the executable path if Chromium is not discoverable automatically.

---

# 4. Create the Project

Create the directory:

```bash
mkdir invoice-generator
cd invoice-generator
```

Initialize Go:

```bash
go mod init invoice-generator
```

Initialize Git:

```bash
git init
```

Create the initial structure:

```text
invoice-generator/
├── cmd/
├── internal/
├── templates/
├── examples/
├── output/
├── tests/
├── README.md
├── .gitignore
└── go.mod
```

Do not create every file immediately.

Create files as each feature is implemented.

---

# 5. Architecture

Keep the application divided into clear layers.

```text
cmd/
    CLI commands

internal/
    invoice/
        Domain model
        Validation
        Calculation

    input/
        CSV parser
        Excel parser

    pdf/
        HTML rendering
        PDF generation

templates/
    Invoice HTML/CSS

examples/
    Example input files

output/
    Generated PDFs
```

Dependency direction:

```text
CLI
 |
 +----> input
 |
 +----> invoice
 |
 +----> pdf
```

The `invoice` package should not know anything about Excel, CSV, CLI or Chromium.

This is important because the invoice domain may later move into your accounting application.

---

# 6. Step 1: Build the Invoice Domain Model

Create:

```text
internal/invoice/invoice.go
```

Start with the core entities.

Conceptually:

```go
type Invoice struct {
    Number       string
    Date         time.Time
    DueDate      *time.Time
    Customer     Customer
    Items        []InvoiceItem
    Discount     Money
    TaxRate      decimal/rational representation
    Notes        string
    Currency     string
}
```

Customer:

```go
type Customer struct {
    Name    string
    Phone   string
    Email   string
    Address string
    TaxID   string
}
```

Invoice item:

```go
type InvoiceItem struct {
    Description string
    Quantity    float64
    UnitPrice   int64
}
```

For money, prefer an integer representation such as the smallest currency unit rather than `float64` for actual monetary calculations.

For example:

```text
500,000 تومان
```

can internally be represented as:

```text
500000
```

This avoids common floating-point problems.

---

# 7. Step 2: Define Calculated Values

An invoice needs:

```text
Subtotal
Discount
Tax
Total
```

The calculation flow:

```text
Item 1 total = quantity × unit price
Item 2 total = quantity × unit price
...
Subtotal = sum(item totals)

Discounted subtotal =
    subtotal - discount

Tax =
    discounted subtotal × tax rate

Total =
    discounted subtotal + tax
```

Create:

```text
internal/invoice/calculator.go
```

Keep calculations as pure functions whenever possible.

Example concept:

```go
func CalculateSubtotal(items []InvoiceItem) int64
```

Then:

```go
func CalculateTax(base int64, taxRate decimal...) int64
```

Finally:

```go
func CalculateTotal(invoice Invoice) Totals
```

Create a result structure:

```go
type Totals struct {
    Subtotal int64
    Discount int64
    Tax      int64
    Total    int64
}
```

---

# 8. Step 3: Write Calculation Tests

Before building Excel support, test the business logic.

Create:

```text
internal/invoice/calculator_test.go
```

Test cases:

### Case 1: One item

```text
quantity = 2
unit price = 500
subtotal = 1000
```

### Case 2: Multiple items

```text
2 × 500
3 × 200

subtotal = 1600
```

### Case 3: Discount

```text
subtotal = 1600
discount = 100

base = 1500
```

### Case 4: Tax

```text
base = 1500
tax = 10%

tax = 150
```

### Case 5: Final total

```text
1500 + 150 = 1650
```

Run:

```bash
go test ./...
```

Do not move forward until these calculations work.

---

# 9. Step 4: Validation

Create:

```text
internal/invoice/validator.go
```

Validate:

### Invoice

- invoice number exists
- date is valid
- customer name exists
- at least one item exists
- currency exists

### Item

- description exists
- quantity > 0
- unit price >= 0

### Discount

```text
discount >= 0
discount <= subtotal
```

### Tax

```text
tax rate >= 0
tax rate <= reasonable maximum
```

Do not allow invalid invoices to reach the PDF generator.

---

# 10. Step 5: Define the Input Format

Before implementing Excel parsing, define exactly what the user will enter.

Recommended Excel workbook:

```text
Sheet: Invoice
```

Example:

| Field | Value |
|---|---|
| invoice_number | 10023 |
| date | 2026-09-12 |
| due_date | 2026-09-20 |
| customer_name | Company ABC |
| customer_phone | 09120000000 |
| customer_email | info@example.com |
| customer_address | Tehran |
| tax_id | 123456 |
| currency | IRR |
| discount | 100000 |
| tax_rate | 10 |
| notes | Thank you |
```

Then:

```text
Sheet: Items
```

| description | quantity | unit_price |
|---|---:|---:|
| Product A | 2 | 500000 |
| Product B | 3 | 250000 |
| Product C | 1 | 1000000 |

This structure is simple for humans and simple for the parser.

---

# 11. Step 6: Create a Sample Excel File

Create:

```text
examples/invoice.xlsx
```

It should contain the two sheets described above.

Also create a CSV alternative:

```text
examples/items.csv
```

CSV can represent the item list, while invoice metadata can be passed through CLI flags or a separate metadata file.

For Version 1, Excel can be the most complete input format.

---

# 12. Step 7: CSV Parser

Create:

```text
internal/input/csv.go
```

Use the standard library:

```go
encoding/csv
```

Responsibilities:

```text
CSV file
  ↓
read records
  ↓
validate headers
  ↓
parse values
  ↓
create []InvoiceItem
```

Do not put calculation logic here.

The parser only converts external data into domain data.

---

# 13. Step 8: Excel Parser

Create:

```text
internal/input/excel.go
```

Use:

```text
github.com/xuri/excelize/v2
```

Responsibilities:

```text
invoice.xlsx
    |
    +-- Invoice sheet
    |
    +-- Items sheet
          |
          v
       Invoice
```

The parser should:

1. Open workbook.
2. Check required sheets exist.
3. Read invoice metadata.
4. Read item headers.
5. Parse every item.
6. Convert values to domain types.
7. Return an `Invoice`.
8. Return useful errors when something is wrong.

Example errors:

```text
missing sheet: Items
missing field: customer_name
invalid quantity on row 5
invalid unit_price on row 8
```

Avoid vague errors like:

```text
something went wrong
```

---

# 14. Step 9: Normalize Input

Different input sources should eventually produce exactly the same domain object.

```text
Excel ─────┐
           │
CSV ───────┼──> Invoice
           │
CLI ───────┘
```

This is a key architectural rule.

The PDF generator should not care whether the invoice came from Excel or CLI.

---

# 15. Step 10: Build the CLI

Create:

```text
cmd/invoice/main.go
```

Use Cobra.

Initial command:

```bash
invoice generate invoice.xlsx
```

Expected flow:

```text
CLI
 ↓
Read file
 ↓
Detect format
 ↓
Parse
 ↓
Validate
 ↓
Calculate
 ↓
Generate PDF
 ↓
Print output path
```

Output:

```text
Invoice generated successfully:
output/invoice-10023.pdf
```

---

# 16. Step 11: Generate HTML

Create:

```text
templates/invoice.html
```

Use Go's:

```go
html/template
```

The template receives something like:

```text
Invoice
Totals
```

The HTML should contain:

```text
Header
Customer information
Invoice information
Items table
Subtotal
Discount
Tax
Total
Notes
Footer
```

Example structure:

```html
<html>
<head>
    ...
</head>

<body>
    <header>
        ...
    </header>

    <section class="invoice-meta">
        ...
    </section>

    <table>
        ...
    </table>

    <section class="totals">
        ...
    </section>
</body>
</html>
```

---

# 17. Step 12: Design the Invoice

Keep the first design professional and boring in a good way.

Use:

- clean typography
- strong hierarchy
- clear table
- generous spacing
- readable totals
- minimal decoration

Avoid:

- gradients
- excessive colors
- dashboard-like UI
- unnecessary icons
- complicated graphics

The PDF is the product.

---

# 18. Step 13: Support Persian / RTL

Because invoices may be Persian, design the template for RTL from the beginning.

HTML:

```html
<html lang="fa" dir="rtl">
```

CSS should support:

```css
direction: rtl;
text-align: right;
```

But numerical fields may need careful alignment.

For example:

```text
تعداد       قیمت واحد       مبلغ
2           500,000         1,000,000
```

Keep numbers readable and consistent.

---

# 19. Step 14: Font Handling

PDF generation depends on fonts.

Do not rely blindly on the user's system fonts.

Choose a font that supports:

- Persian
- Latin
- Arabic numerals
- punctuation

Ideally package the required font with the project or document clearly how it should be installed.

Test:

```text
English
فارسی
123456
۱۲۳۴۵۶
```

---

# 20. Step 15: Generate PDF

Create:

```text
internal/pdf/generator.go
```

Use Chromium through `chromedp`.

Conceptually:

```text
Invoice
   ↓
Go HTML Template
   ↓
HTML document
   ↓
Chromium
   ↓
PDF bytes
   ↓
output/invoice-10023.pdf
```

The PDF generator should expose something conceptually like:

```go
Generate(ctx context.Context, invoice Invoice, outputPath string) error
```

Do not let CLI code know how Chromium works.

---

# 21. Step 16: Temporary HTML

A simple first implementation can:

1. Render template to HTML.
2. Save HTML to a temporary file.
3. Open it with Chromium.
4. Print to PDF.
5. Delete temporary file.

Later you can optimize this.

Do not optimize early.

---

# 22. Step 17: Output Naming

Use:

```text
invoice-{invoice_number}.pdf
```

Example:

```text
invoice-10023.pdf
```

Sanitize invoice numbers so a malicious or accidental value cannot create unexpected filesystem paths.

Do not allow:

```text
../../something
```

to become an output path.

---

# 23. Step 18: Add `validate`

Implement:

```bash
invoice validate invoice.xlsx
```

Expected:

```text
Invoice is valid.
```

Or:

```text
Invoice validation failed:

- customer_name is required
- item row 4 has invalid quantity
- tax_rate must not be negative
```

This command should NOT generate a PDF.

---

# 24. Step 19: Add `preview`

Optional command:

```bash
invoice preview invoice.xlsx
```

Version 1 can simply generate an HTML file:

```text
output/invoice-10023.html
```

Then open it in the default browser.

This makes template development much faster than regenerating PDFs every time.

---

# 25. Step 20: Add `batch`

Implement:

```bash
invoice batch ./examples/invoices/
```

Input:

```text
invoices/
├── invoice-1.xlsx
├── invoice-2.xlsx
└── invoice-3.xlsx
```

Output:

```text
output/
├── invoice-1001.pdf
├── invoice-1002.pdf
└── invoice-1003.pdf
```

For each file:

```text
Read
 ↓
Parse
 ↓
Validate
 ↓
Generate
```

If one invoice fails, print the error and continue with the others unless `--fail-fast` is specified.

---

# 26. Step 21: CLI Options

Eventually support:

```bash
invoice generate invoice.xlsx \
    --output ./output \
    --template ./templates/invoice.html
```

Useful flags:

```text
--output
--template
--format
--open
--verbose
```

But don't add all flags immediately.

Start with:

```bash
invoice generate invoice.xlsx
```

---

# 27. Step 22: Configuration

Only introduce configuration when needed.

Possible:

```text
config.yaml
```

Example:

```yaml
company:
  name: My Company
  phone: ...
  email: ...
  address: ...

invoice:
  currency: IRR
  tax_rate: 10
  output_dir: ./output
```

This allows company information to stay out of every Excel file.

A future workflow:

```bash
invoice generate invoice.xlsx
```

uses:

```text
invoice.xlsx
+
config.yaml
```

to generate the final invoice.

---

# 28. Step 23: Company Information

Add an optional company configuration:

```text
Company
├── Name
├── Logo
├── Phone
├── Email
├── Address
├── Website
└── Tax ID
```

This belongs to configuration, not each individual invoice.

The final document can have:

```text
┌────────────────────────────────────┐
│ COMPANY NAME                 LOGO  │
│ Phone | Email | Address           │
│                                    │
│             INVOICE                │
│                                    │
│ Customer: ...                      │
└────────────────────────────────────┘
```

---

# 29. Step 24: Logo Support

Allow:

```yaml
company:
  logo: ./assets/logo.png
```

The template should render it if provided.

If no logo exists:

```text
do not render empty logo space
```

Keep the invoice valid without a logo.

---

# 30. Step 25: Error Handling

Errors should be actionable.

Bad:

```text
failed
```

Good:

```text
failed to parse Excel file:
examples/invoice.xlsx

sheet "Items":
row 5:
quantity "abc" is not a valid number
```

Use wrapped errors in Go:

```go
fmt.Errorf("failed to parse items: %w", err)
```

Do not panic for normal user input errors.

---

# 31. Step 26: Logging

Keep CLI output clean.

Normal mode:

```text
Reading invoice.xlsx...
Validating invoice...
Generating PDF...
Done: output/invoice-10023.pdf
```

Verbose mode:

```bash
invoice generate invoice.xlsx --verbose
```

can show more technical information.

---

# 32. Step 27: Testing Strategy

Test the parts that contain logic.

## Unit tests

Test:

- calculations
- validation
- CSV parsing
- Excel parsing
- filename sanitization

## Integration tests

Test:

```text
Excel
 ↓
Invoice
 ↓
PDF
```

You don't need hundreds of tests.

Focus on critical behavior.

---

# 33. Step 28: Golden PDF Tests

PDF files can be difficult to compare byte-for-byte because metadata can change.

Do not initially compare raw PDF bytes.

Instead test:

```text
PDF was generated
PDF is not empty
PDF can be opened
expected invoice number exists
expected customer name exists
```

You can improve this later.

---

# 34. Step 29: Security Considerations

Even a local CLI should avoid obvious problems.

Validate:

- file paths
- invoice numbers
- output paths
- numeric values
- template paths

Do not execute arbitrary shell commands from input files.

Do not allow user-provided HTML to accidentally become executable browser content if the tool later handles untrusted input.

Keep external template loading controlled.

---

# 35. Step 30: Git Workflow

Suggested commits:

```text
init project
add invoice domain model
add invoice calculations
add calculation tests
add invoice validation
add csv parser
add excel parser
add cli generate command
add html invoice template
add pdf generator
add rtl support
add validate command
add preview command
add batch command
add configuration
improve error handling
add integration tests
```

Keep commits small enough that you understand what changed.

---

# 36. Recommended Development Order

Follow this exact order:

```text
1. Create Go project
       ↓
2. Invoice structs
       ↓
3. Money representation
       ↓
4. Calculation functions
       ↓
5. Calculation tests
       ↓
6. Validation
       ↓
7. Excel format
       ↓
8. Excel parser
       ↓
9. CLI
       ↓
10. HTML template
       ↓
11. PDF generation
       ↓
12. RTL/Persian support
       ↓
13. Output handling
       ↓
14. validate command
       ↓
15. preview command
       ↓
16. configuration
       ↓
17. company/logo
       ↓
18. batch generation
       ↓
19. integration tests
       ↓
20. packaging/documentation
```

---

# 37. Milestone 1: Domain

Goal:

```text
Invoice
Customer
InvoiceItem
Totals
```

Definition of done:

```bash
go test ./...
```

passes.

No CLI yet.

---

# 38. Milestone 2: Input

Goal:

```text
invoice.xlsx
    ↓
Invoice struct
```

Definition of done:

A valid Excel file can be converted into an `Invoice`.

Invalid Excel produces useful errors.

---

# 39. Milestone 3: First CLI

Goal:

```bash
invoice validate invoice.xlsx
```

Definition of done:

```text
valid file   → success
invalid file → readable errors
```

---

# 40. Milestone 4: First PDF

Goal:

```bash
invoice generate invoice.xlsx
```

produces:

```text
output/invoice-10023.pdf
```

Definition of done:

- PDF opens
- customer appears
- invoice number appears
- items appear
- totals are correct

---

# 41. Milestone 5: Professional Output

Improve:

- typography
- spacing
- table
- logo
- company information
- Persian RTL
- page margins
- print layout
- footer

Do not change the domain model just to improve styling.

---

# 42. Milestone 6: Usability

Add:

```bash
invoice preview invoice.xlsx
invoice batch ./invoices/
```

Improve:

- errors
- progress messages
- output naming
- configuration

---

# 43. Final CLI

The desired final interface:

```bash
invoice generate invoice.xlsx
```

```bash
invoice validate invoice.xlsx
```

```bash
invoice preview invoice.xlsx
```

```bash
invoice batch ./invoices/
```

Optional:

```bash
invoice version
invoice help
```

---

# 44. Example Final Project

```text
invoice-generator/
│
├── cmd/
│   └── invoice/
│       └── main.go
│
├── internal/
│   ├── invoice/
│   │   ├── invoice.go
│   │   ├── calculator.go
│   │   ├── calculator_test.go
│   │   ├── validator.go
│   │   └── validator_test.go
│   │
│   ├── input/
│   │   ├── csv.go
│   │   ├── csv_test.go
│   │   ├── excel.go
│   │   └── excel_test.go
│   │
│   ├── pdf/
│   │   ├── generator.go
│   │   └── generator_test.go
│   │
│   └── config/
│       └── config.go
│
├── templates/
│   └── invoice.html
│
├── assets/
│   └── logo.png
│
├── examples/
│   ├── invoice.xlsx
│   └── items.csv
│
├── output/
│   └── .gitkeep
│
├── README.md
├── .gitignore
└── go.mod
```

---

# 45. What You Should Learn While Building It

This project is also a practical Go learning project.

Learn each concept exactly when you need it.

## Go basics

- structs
- methods
- pointers
- slices
- maps
- interfaces
- packages
- errors
- `defer`
- file I/O
- JSON
- time
- context

## Go project structure

Learn:

```text
package boundaries
internal/
cmd/
go.mod
```

## CLI

Learn:

```text
Cobra
commands
flags
arguments
exit codes
```

## File processing

Learn:

```text
os
io
bufio
encoding/csv
Excel libraries
```

## Templates

Learn:

```text
html/template
template data
conditional rendering
loops
custom template functions
```

## PDF

Learn:

```text
Chromium
headless browser
HTML → PDF
chromedp
```

## Testing

Learn:

```text
testing
table-driven tests
integration tests
```

This makes the project useful as both a tool and a Go learning exercise.

---

# 46. Future Migration to the Accounting App

Do NOT build the accounting app now.

But design the invoice domain so it can later become:

```text
Accounting Application
│
├── Customer
├── Product
├── Invoice
│   ├── InvoiceItem
│   └── Totals
│
├── Payment
├── Expense
├── Transaction
└── Reports
```

The important reusable part is:

```text
internal/invoice/
```

Later it can become part of a larger Go backend.

For example:

```text
Future Accounting Backend
│
├── customer/
├── product/
├── invoice/
├── payment/
├── expense/
└── transaction/
```

The current project should therefore keep the invoice logic independent from:

- CLI
- Excel
- PDF
- filesystem
- Chromium

---

# 47. Possible Future Architecture

If the accounting application eventually becomes serious:

```text
                  Accounting Web App
                         |
                      REST API
                         |
                    Go Backend
                         |
        ┌────────────────┼────────────────┐
        │                │                │
    Customers        Invoices         Payments
        │                │                │
        └────────────────┼────────────────┘
                         |
                     PostgreSQL
```

The current invoice generator can evolve into:

```text
Invoice Domain
     |
     +---- CLI input
     |
     +---- Excel input
     |
     +---- API input
     |
     +---- Web UI input
```

This is why separating the domain from infrastructure matters.

---

# 48. Definition of Done

The project is finished when all of these work:

### Input

- [ ] Valid Excel can be read
- [ ] Invalid Excel gives useful errors
- [ ] CSV item input works

### Domain

- [ ] Invoice model exists
- [ ] Customer model exists
- [ ] InvoiceItem model exists
- [ ] Money calculation is deterministic
- [ ] Totals are correct

### Validation

- [ ] Required fields are checked
- [ ] Quantity is validated
- [ ] Price is validated
- [ ] Discount is validated
- [ ] Tax is validated

### PDF

- [ ] HTML template renders
- [ ] PDF is generated
- [ ] PDF opens correctly
- [ ] Persian text works
- [ ] RTL layout works
- [ ] Logo works
- [ ] Totals are correct

### CLI

- [ ] `generate` works
- [ ] `validate` works
- [ ] `preview` works
- [ ] `batch` works

### Quality

- [ ] Unit tests pass
- [ ] Integration test works
- [ ] `go fmt` clean
- [ ] `go vet ./...` clean
- [ ] README explains installation and usage
- [ ] Example Excel file is included

---

# 49. First Session Checklist

Do NOT try to build everything in one sitting.

Start with only this:

```text
[ ] Install/verify Go
[ ] Create invoice-generator
[ ] go mod init
[ ] Initialize Git
[ ] Create internal/invoice
[ ] Create Invoice struct
[ ] Create Customer struct
[ ] Create InvoiceItem struct
[ ] Create Totals struct
[ ] Implement subtotal
[ ] Implement discount
[ ] Implement tax
[ ] Implement total
[ ] Write tests
[ ] Run go test ./...
```

At the end of the first session, you should have:

```text
Go
 ↓
Invoice domain
 ↓
Correct calculations
 ↓
Passing tests
```

Nothing else.

That is the correct starting point.

---

# 50. Development Philosophy

Keep asking:

> "Does this feature help me turn invoice data into a PDF?"

If the answer is no, postpone it.

The first version should feel almost ridiculously small.

The target is:

```text
invoice.xlsx
     ↓
   Go CLI
     ↓
 invoice.pdf
```

Once that works reliably, everything else is refinement.

Do not accidentally turn this into the accounting application before the accounting application exists.
