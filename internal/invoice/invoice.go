package invoice

import "time"

type InvoiceItem struct {
	ID              uint
	TechnicalCode   string
	Description     string
	Quantity        uint
	UnitPrice       int64
	DiscountPercent float32
	FinalUnitPrice  int64
	Price           int64
}

type Invoice struct {
	Number string
	Date   time.Time
	Customer string
	Items []InvoiceItem
	Description string
}