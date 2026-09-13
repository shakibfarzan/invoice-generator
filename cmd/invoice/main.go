// Package main is the entry point for the invoice CLI.
//
// Usage:
//
//	invoice generate <file>          Parse, validate, and calculate an invoice
//	invoice help                     Show help
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/spf13/cobra"

	"invoice-generator/internal/input"
	"invoice-generator/internal/invoice"
	"invoice-generator/internal/pdf"
)

// pdfTimeout bounds how long headless Chromium may take to print a single
// invoice. Without it, a hung browser would hang the CLI forever.
const pdfTimeout = 60 * time.Second

// ---------------------------------------------------------------------------
// Flags — variables that CLI flags write into.
// ---------------------------------------------------------------------------

var (
	// CSV files only carry item rows, so metadata must come from flags.
	flagNumber   string
	flagDate     string
	flagCustomer string

	// Where the generated PDF is written.
	flagOutput string
)

// ---------------------------------------------------------------------------
// main — build the command tree and execute it.
// ---------------------------------------------------------------------------

func main() {
	// Root command: "invoice"
	rootCmd := &cobra.Command{
		Use:   "invoice",
		Short: "Invoice Generator — create professional invoices from spreadsheet data",
		Long: `A local-first CLI that reads invoice data from Excel (.xlsx) or CSV files,
validates it, runs calculations, and (eventually) produces a PDF.`,

		// SilenceUsage prevents Cobra from printing the full usage text
		// every time a command returns an error — we want clean output.
		SilenceUsage: true,
	}

	// Subcommand: "invoice generate <file>"
	generateCmd := &cobra.Command{
		Use:   "generate [file]",
		Short: "Parse, validate, and calculate totals for an invoice",
		Long: `Read an invoice from an Excel (.xlsx) or CSV file, validate it against
business rules, compute every item's price, print a summary, and write a PDF
to the output directory as invoice-<number>.pdf.

Excel workbooks must contain two sheets:
  • "Invoice" — key/value metadata (invoice_number, date, customer, ...)
  • "Items"   — one row per line item

CSV files contain only the items table. Supply the missing metadata with
the --number, --date, and --customer flags.`,

		// Exactly one positional argument: the file path.
		Args: cobra.ExactArgs(1),

		// RunE is like Run but lets us return an error. Cobra will print
		// the error message and exit with a non-zero status.
		RunE: runGenerate,
	}

	// Attach flags to the generate subcommand.
	generateCmd.Flags().StringVar(&flagNumber, "number", "",
		"invoice number (required when input is CSV)")
	generateCmd.Flags().StringVar(&flagDate, "date", "",
		"invoice date in YYYY-MM-DD format (required when input is CSV)")
	generateCmd.Flags().StringVar(&flagCustomer, "customer", "",
		"customer name (required when input is CSV)")
	generateCmd.Flags().StringVarP(&flagOutput, "output", "o", "output",
		"directory where the PDF is written")

	// Wire the subcommand into the root.
	rootCmd.AddCommand(generateCmd)

	// Execute parses os.Args and dispatches to the matching command.
	if err := rootCmd.Execute(); err != nil {
		// Cobra already printed the error; we just need a non-zero exit.
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// runGenerate — the heart of "invoice generate <file>".
// ---------------------------------------------------------------------------

func runGenerate(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	// 1. Parse ---------------------------------------------------------------
	fmt.Printf("Reading %s...\n", filePath)

	inv, err := parseInput(filePath)
	if err != nil {
		return err // wrapped with context by the parser
	}

	// 2. Validate ------------------------------------------------------------
	fmt.Println("Validating invoice...")

	if err := invoice.Validate(inv); err != nil {
		return err // *ValidationError formats itself nicely
	}

	// 3. Calculate -----------------------------------------------------------
	fmt.Println("Calculating totals...")

	invoice.ComputeItemPrices(inv.Items)
	total := invoice.CalculateTotal(inv)

	// 4. Print summary -------------------------------------------------------
	printSummary(inv, total)

	// 5. Write the PDF -------------------------------------------------------
	outputPath, err := writePDF(inv, total)
	if err != nil {
		return err
	}

	fmt.Printf("✓ Wrote %s\n", outputPath)

	return nil
}

// ---------------------------------------------------------------------------
// writePDF — render an invoice and save it as output/invoice-<number>.pdf.
// ---------------------------------------------------------------------------

func writePDF(inv invoice.Invoice, total int64) (string, error) {
	// Roadmap step 17: the file name is derived from the invoice number,
	// so that number has to be made safe before it touches the filesystem.
	outputPath := filepath.Join(flagOutput, "invoice-"+sanitizeFilePart(inv.Number)+".pdf")

	ctx, cancel := context.WithTimeout(context.Background(), pdfTimeout)
	defer cancel()

	if err := pdf.Generate(ctx, inv, total, outputPath); err != nil {
		return "", err
	}
	return outputPath, nil
}

// sanitizeFilePart reduces a string to characters that are safe inside a
// single file name.
//
// Everything that could change the meaning of a path — separators, spaces,
// control characters, or a name made only of dots — is replaced or dropped,
// so an invoice number like "../../etc/passwd" can never write outside the
// output directory.
func sanitizeFilePart(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	cleaned := strings.Trim(b.String(), "-._")
	if cleaned == "" {
		return "unknown"
	}
	return cleaned
}

// ---------------------------------------------------------------------------
// parseInput — detect file format and delegate to the right parser.
// ---------------------------------------------------------------------------

func parseInput(path string) (invoice.Invoice, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".xlsx", ".xls":
		inv, err := input.ParseInvoiceExcel(path)
		if err != nil {
			return invoice.Invoice{}, fmt.Errorf("failed to read Excel file: %w", err)
		}
		return inv, nil

	case ".csv":
		items, err := input.ParseItemsCSVFile(path)
		if err != nil {
			return invoice.Invoice{}, fmt.Errorf("failed to read CSV file: %w", err)
		}

		// CSV only has items; build the invoice from flags + parsed items.
		inv := invoice.Invoice{
			Number:   flagNumber,
			Customer: flagCustomer,
			Items:    items,
		}

		if flagDate != "" {
			date, err := time.Parse("2006-01-02", flagDate)
			if err != nil {
				return invoice.Invoice{},
					fmt.Errorf("invalid --date %q: expected YYYY-MM-DD", flagDate)
			}
			inv.Date = date
		}

		return inv, nil

	default:
		return invoice.Invoice{},
			fmt.Errorf("unsupported file format %q — use .xlsx or .csv", ext)
	}
}

// ---------------------------------------------------------------------------
// printSummary — human-readable output (until PDF generation lands).
// ---------------------------------------------------------------------------

func printSummary(inv invoice.Invoice, total int64) {
	fmt.Println()
	fmt.Println("✓ Invoice is valid!")
	fmt.Println()

	// Metadata
	fmt.Printf("  Invoice #:  %s\n", inv.Number)
	fmt.Printf("  Date:       %s\n", inv.Date.Format("2006-01-02"))
	fmt.Printf("  Customer:   %s\n", inv.Customer)
	if inv.Description != "" {
		fmt.Printf("  Notes:      %s\n", inv.Description)
	}

	// Items table
	fmt.Println()
	fmt.Printf("  %-4s  %-30s  %6s  %12s  %16s  %12s\n",
		"No.", "Description", "Qty", "Unit Price", "Discount Percent", "Price")
	fmt.Println("  " + strings.Repeat("-", 72))

	for i, item := range inv.Items {
		desc := item.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		fmt.Printf("  %-4d  %-30s  %6d  %12d  %16f  %12d\n",
			i+1, desc, item.Quantity, item.FinalUnitPrice, item.DiscountPercent, item.Price)
	}

	fmt.Println("  " + strings.Repeat("-", 72))
	fmt.Printf("  %54s  %12d\n", "Total:", total)
	fmt.Println()
}
