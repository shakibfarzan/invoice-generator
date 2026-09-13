// Package main is the entry point for the invoice CLI.
//
// Usage:
//
//	invoice generate <file>          Parse, validate, and calculate an invoice
//	invoice help                     Show help
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"invoice-generator/internal/input"
	"invoice-generator/internal/invoice"
)

// ---------------------------------------------------------------------------
// Flags — variables that CLI flags write into.
// ---------------------------------------------------------------------------

var (
	// CSV files only carry item rows, so metadata must come from flags.
	flagNumber   string
	flagDate     string
	flagCustomer string
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
business rules, compute every item's price, and print a summary.

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

	return nil
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
	fmt.Printf("  %-4s  %-30s  %6s  %12s  %12s\n",
		"No.", "Description", "Qty", "Unit Price", "Price")
	fmt.Println("  " + strings.Repeat("-", 72))

	for i, item := range inv.Items {
		desc := item.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		fmt.Printf("  %-4d  %-30s  %6d  %12d  %12d\n",
			i+1, desc, item.Quantity, item.FinalUnitPrice, item.Price)
	}

	fmt.Println("  " + strings.Repeat("-", 72))
	fmt.Printf("  %54s  %12d\n", "Total:", total)
	fmt.Println()
	fmt.Println("  (PDF generation coming in a future step)")
}
