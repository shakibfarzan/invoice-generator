package pdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"invoice-generator/internal/invoice"
)

// A4 page size in inches, which is what Chromium's Page.printToPDF expects.
const (
	a4WidthInches  = 8.27
	a4HeightInches = 11.69
)

// Generate renders the invoice to PDF and writes it to outputPath, creating
// the parent directory if it does not exist.
//
// Items must already be priced (invoice.ComputeItemPrices) and total must
// already be calculated (invoice.CalculateTotal); the PDF layer never
// computes business values.
//
// The caller controls cancellation and timeouts through ctx.
func Generate(ctx context.Context, inv invoice.Invoice, total int64, outputPath string) error {
	html, err := Render(NewView(inv, total))
	if err != nil {
		return err
	}
	return GenerateFromHTML(ctx, html, outputPath)
}

// GenerateFromHTML prints an already rendered HTML document to outputPath.
//
// Roadmap step 16: the first implementation writes the HTML to a temporary
// file, opens it in Chromium, prints it and deletes the file. That round
// trip is simple and robust; optimizing it now would be premature.
func GenerateFromHTML(ctx context.Context, html, outputPath string) error {
	tempPath, err := writeTempHTML(html)
	if err != nil {
		return err
	}
	defer os.Remove(tempPath)

	// Chromium loads local files more predictably from a file:// URL than
	// from stdin, and the URL keeps working on Windows where the temporary
	// path contains backslashes.
	data, err := printToPDF(ctx, "file://"+filepath.ToSlash(tempPath))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory %q: %w", filepath.Dir(outputPath), err)
	}

	if err := os.WriteFile(outputPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write PDF %q: %w", outputPath, err)
	}
	return nil
}

// writeTempHTML writes html to a temporary file and returns its path.
// The caller is responsible for removing the file.
func writeTempHTML(html string) (string, error) {
	f, err := os.CreateTemp("", "invoice-*.html")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary HTML file: %w", err)
	}

	if _, err := f.WriteString(html); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", fmt.Errorf("failed to write temporary HTML file: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(f.Name())
		return "", fmt.Errorf("failed to close temporary HTML file: %w", err)
	}
	return f.Name(), nil
}

// printToPDF opens url in headless Chromium and returns the printed PDF.
//
// Margins are zero because the template's CSS box owns the page padding;
// letting both Chromium and CSS add margins would double them.
func printToPDF(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	var data []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		// Wait for the document to be laid out before printing, otherwise
		// Chromium can capture a half-rendered page.
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(a4WidthInches).
				WithPaperHeight(a4HeightInches).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				Do(ctx)
			if err != nil {
				return err
			}
			data = buf
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to print PDF with headless Chromium: %w", err)
	}
	return data, nil
}