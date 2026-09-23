package helper

import (
	"strconv"
	"strings"
)

func GetFormattedCurrency(currencyCode string, amount int64) string {
	currencyCode = strings.ToUpper(currencyCode)
	switch currencyCode {
	case "IDR":
		return formatIDR(amount)
	}
	return ""
}

func formatIDR(amount int64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	digits := strconv.FormatInt(amount, 10)
	var b strings.Builder
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(d)
	}
	return sign + "Rp " + b.String()
}
