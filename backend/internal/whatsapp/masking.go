package whatsapp

import (
	"strings"
	"unicode"
)

// MaskPhoneNumber masks a Kenyan phone number into standard privacy format (e.g. "07•• ••• 482").
// Preserves the leading provider digits ("07" or "01") and the trailing 3 digits while obscuring the rest.
func MaskPhoneNumber(raw string) string {
	var digits []rune
	for _, r := range raw {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		}
	}

	dStr := string(digits)
	// Normalize international "+254" or "254" to local "0"
	if strings.HasPrefix(dStr, "254") && len(dStr) >= 12 {
		dStr = "0" + dStr[3:]
	}

	if len(dStr) < 7 {
		return "07•• ••• •••"
	}

	prefix := dStr[:2]
	if prefix != "07" && prefix != "01" {
		prefix = "07"
	}
	suffix := dStr[len(dStr)-3:]

	return prefix + "•• ••• " + suffix
}
