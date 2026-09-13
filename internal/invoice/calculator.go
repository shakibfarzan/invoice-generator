package invoice

// ComputeItemPrices fills in FinalUnitPrice and Price for every item based
// on its UnitPrice, DiscountPercent, and Quantity.
//
//   FinalUnitPrice = UnitPrice − (UnitPrice × DiscountPercent / 100)
//   Price          = FinalUnitPrice × Quantity
//
// The slice is modified in place.
func ComputeItemPrices(items []InvoiceItem) {
	for i := range items {
		if items[i].DiscountPercent > 0 {
			discount := int64(float64(items[i].UnitPrice) * float64(items[i].DiscountPercent) / 100)
			items[i].FinalUnitPrice = items[i].UnitPrice - discount
		} else {
			items[i].FinalUnitPrice = items[i].UnitPrice
		}
		items[i].Price = items[i].FinalUnitPrice * int64(items[i].Quantity)
	}
}

// CalculateTotal returns the sum of every item's Price.
// Call ComputeItemPrices first so that Price is populated.
func CalculateTotal(invoice Invoice) int64 {
	sum := int64(0)
	for _, item := range invoice.Items {
		sum += item.Price
	}
	return sum
}
