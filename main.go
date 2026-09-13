// This file is a placeholder. The real CLI entry point lives in cmd/invoice/.
//
// To build and run:
//
//	go build -o invoice ./cmd/invoice
//	./invoice generate examples/invoice.xlsx
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("The invoice CLI entry point is in cmd/invoice/.")
	fmt.Println()
	fmt.Println("Build it with:")
	fmt.Println("  go build -o invoice ./cmd/invoice")
	fmt.Println()
	fmt.Println("Then run:")
	fmt.Println("  ./invoice generate examples/invoice.xlsx")
	os.Exit(1)
}
