package whatsapp

import (
	"errors"
	"strings"

	"github.com/Ashomondi/ikonex-academy/backend/internal/notifications"
)

var ErrInvalidPhoneNumber = errors.New("invalid phone number")

// MaskPhone normalizes a phone number and returns its masked display representation
// (e.g., "+254712345678" -> "07•• ••• 678"). It never echoes private input in errors.
func MaskPhone(phone string) (string, error) {
	normalized, err := notifications.NormalizePhone(phone)
	if err != nil {
		return "", ErrInvalidPhoneNumber
	}
	if strings.HasPrefix(normalized, "+254") && len(normalized) == 13 {
		localPrefix := "0" + normalized[4:5]
		suffix := normalized[len(normalized)-3:]
		return localPrefix + "•• ••• " + suffix, nil
	}
	if len(normalized) < 8 {
		return "", ErrInvalidPhoneNumber
	}
	prefix := normalized[:3]
	suffix := normalized[len(normalized)-3:]
	return prefix + "•• ••• " + suffix, nil
}
