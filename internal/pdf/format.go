package pdf

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

import ptime "github.com/yaa110/go-persian-calendar"


// digitGroupSeparator separates thousands in rendered amounts.
const digitGroupSeparator = ","

// percentSign is the Persian/Arabic percent sign, which reads correctly in
// an RTL document.
const percentSign = "٪"

var faDigits = strings.NewReplacer(
	"0", "۰", "1", "۱", "2", "۲", "3", "۳", "4", "۴",
	"5", "۵", "6", "۶", "7", "۷", "8", "۸", "9", "۹",
)

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

func ToPersianDigits(v any) string {
	var s string
	switch x := v.(type) {
	case string:
		s = x
	case int:
		s = strconv.Itoa(x)
	case int64:
		s = strconv.FormatInt(x, 10)
	case float64:
		s = strconv.FormatFloat(x, 'f', -1, 64)
	default:
		s = fmt.Sprintf("%v", x)
	}
	return faDigits.Replace(s)
}

func ToJalali(t time.Time) string {
	pt := ptime.New(t)
	return ToPersianDigits(fmt.Sprintf("%04d/%02d/%02d", pt.Year(), int(pt.Month()), pt.Day()))
}