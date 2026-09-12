package invoice

func CalculateTotal(invoice Invoice) int64 {
	sum := int64(0)
	for _, item := range invoice.Items {
		sum += item.Price
	}
	return sum
}
