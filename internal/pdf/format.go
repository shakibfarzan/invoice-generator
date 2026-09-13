package pdf

import (
	"strconv"
	"strings"
)

// digitGroupSeparator separates thousands in rendered amounts.
const digitGroupSeparator = ","

// percentSign is the Persian/Arabic percent sign, which reads correctly in
// an RTL document.
const percentSign = "٪"

// FormatMoney renders an amount as digits grouped in threes.
//
// The domain stores money as a plain integer (the smallest currency unit),
// so formatting belongs to the presentation layer: the template asks for a
// formatted string and never does arithmetic.
//
//	FormatMoney(0)         // "0"
//	FormatMoney(1234500)   // "1,234,500"
//	FormatMoney(-1234500)  // "-1,234,500"
func FormatMoney(amount int64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	grouped := groupDigits(strconv.FormatInt(amount, 10))
	if negative {
		return "-" + grouped
	}
	return grouped
}

// FormatPercent renders a discount percentage without a trailing ".0".
//
//	FormatPercent(0)     // "0٪"
//	FormatPercent(10)    // "10٪"
//	FormatPercent(12.5)  // "12.5٪"
func FormatPercent(percent float32) string {
	return strconv.FormatFloat(float64(percent), 'f', -1, 32) + percentSign
}

// groupDigits inserts a separator every three digits, counted from the right.
func groupDigits(digits string) string {
	if len(digits) <= 3 {
		return digits
	}

	lead := len(digits) % 3
	if lead == 0 {
		lead = 3
	}

	var b strings.Builder
	b.Grow(len(digits) + len(digits)/3)
	b.WriteString(digits[:lead])
	for i := lead; i < len(digits); i += 3 {
		b.WriteString(digitGroupSeparator)
		b.WriteString(digits[i : i+3])
	}
	return b.String()
}
