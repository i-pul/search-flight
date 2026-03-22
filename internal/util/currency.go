package util

import (
	"fmt"
	"math"
	"strings"

	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var idPrinter = message.NewPrinter(language.Indonesian)

// FormatPrice returns a human-readable price string using locale-aware formatting.
// Numbers are formatted with the Indonesian locale (dots as thousands separators,
// comma as decimal separator). The currency symbol is resolved from the ISO 4217
// code (e.g. "Rp" for IDR, "US$" for USD).
// Unknown currency codes fall back to "CODE amount".
func FormatPrice(amount float64, cur string) string {
	unit, err := currency.ParseISO(strings.ToUpper(cur))
	if err != nil {
		return fmt.Sprintf("%s%.2f", strings.ToUpper(cur), amount)
	}

	sym := currencySymbol(unit)

	// Use decimal places only when the amount has a meaningful fractional part.
	abs := math.Abs(amount)
	var numStr string
	if abs-math.Trunc(abs) > 0.005 {
		numStr = idPrinter.Sprintf("%.2f", amount)
	} else {
		numStr = idPrinter.Sprintf("%.0f", amount)
	}

	return sym + "" + numStr
}

// currencySymbol returns the local symbol for a currency unit in the Indonesian
// locale (e.g. "Rp" for IDR, "US$" for USD) by formatting a unit amount and
// stripping the number part.
func currencySymbol(unit currency.Unit) string {
	s := idPrinter.Sprintf("%u", currency.Symbol(unit.Amount(1)))
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return s
	}
	return strings.Join(parts[:len(parts)-1], " ")
}
