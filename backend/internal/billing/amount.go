package billing

import (
	"errors"
	"strconv"
	"strings"
)

func paymentAmountCents(amount float64) (int64, error) {
	parts := strings.Split(strconv.FormatFloat(amount, 'f', -1, 64), ".")
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || amount <= 0 || whole > 9999999999 {
		return 0, errors.New("payment amount must be positive and fit NUMERIC(12,2)")
	}
	var fraction int64
	if len(parts) == 2 {
		if len(parts[1]) > 2 {
			return 0, errors.New("payment amount must have at most two decimal places")
		}
		fraction, err = strconv.ParseInt(parts[1]+strings.Repeat("0", 2-len(parts[1])), 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return whole*100 + fraction, nil
}
