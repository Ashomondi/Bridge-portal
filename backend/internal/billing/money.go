package billing

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

// Money stores KES minor units with the exact range of NUMERIC(12,2).
type Money int64

const MaxMoney Money = 999999999999

var ErrInvalidMoney = errors.New("amount must be an exact decimal within NUMERIC(12,2)")

func ParseMoney(value string) (Money, error) {
	if value == "" || len(value) > 14 {
		return 0, ErrInvalidMoney
	}
	negative := value[0] == '-'
	if negative {
		value = value[1:]
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, ErrInvalidMoney
	}
	fraction := "00"
	if len(parts) == 2 {
		if len(parts[1]) < 1 || len(parts[1]) > 2 {
			return 0, ErrInvalidMoney
		}
		fraction = parts[1] + strings.Repeat("0", 2-len(parts[1]))
	}
	digits := parts[0] + fraction
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return 0, ErrInvalidMoney
		}
	}
	n, err := strconv.ParseInt(digits, 10, 64)
	if err != nil || n > int64(MaxMoney) {
		return 0, ErrInvalidMoney
	}
	if negative {
		n = -n
	}
	return Money(n), nil
}

func (m Money) String() string {
	n := int64(m)
	sign := ""
	if n < 0 {
		sign, n = "-", -n
	}
	return fmt.Sprintf("%s%d.%02d", sign, n/100, n%100)
}

func (m Money) MarshalJSON() ([]byte, error) {
	if m < -MaxMoney || m > MaxMoney {
		return nil, ErrInvalidMoney
	}
	return json.Marshal(m.String())
}

func (m *Money) UnmarshalJSON(data []byte) error {
	value := string(data)
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &value); err != nil {
			return ErrInvalidMoney
		}
	}
	parsed, err := ParseMoney(value)
	if err == nil {
		*m = parsed
	}
	return err
}

func (m Money) NumericValue() (pgtype.Numeric, error) {
	if m < -MaxMoney || m > MaxMoney {
		return pgtype.Numeric{}, ErrInvalidMoney
	}
	return pgtype.Numeric{Int: big.NewInt(int64(m)), Exp: -2, Valid: true}, nil
}

func (m *Money) ScanNumeric(value pgtype.Numeric) error {
	if !value.Valid || value.NaN || value.InfinityModifier != pgtype.Finite || value.Int == nil {
		return ErrInvalidMoney
	}
	amount := new(big.Int).Set(value.Int)
	exponent := int64(value.Exp) + 2
	if amount.Sign() != 0 {
		if exponent > 12 || exponent < -12 {
			return ErrInvalidMoney
		}
		power := new(big.Int).Exp(big.NewInt(10), big.NewInt(abs(exponent)), nil)
		if exponent >= 0 {
			amount.Mul(amount, power)
		} else {
			var remainder big.Int
			amount.QuoRem(amount, power, &remainder)
			if remainder.Sign() != 0 {
				return ErrInvalidMoney
			}
		}
	}
	if !amount.IsInt64() || amount.Int64() > int64(MaxMoney) || amount.Int64() < -int64(MaxMoney) {
		return ErrInvalidMoney
	}
	*m = Money(amount.Int64())
	return nil
}

func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

func addMoney(a, b Money) (Money, error) {
	if a < -MaxMoney || a > MaxMoney || b < -MaxMoney || b > MaxMoney || a+b < -MaxMoney || a+b > MaxMoney {
		return 0, ErrInvalidMoney
	}
	return a + b, nil
}

func balances(opening, billed, paid Money) (current, credit Money, err error) {
	if opening < 0 || billed < 0 || paid < 0 {
		return 0, 0, ErrInvalidMoney
	}
	total, err := addMoney(opening, billed)
	if err != nil {
		return 0, 0, err
	}
	net, err := addMoney(total, -paid)
	if err != nil {
		return 0, 0, err
	}
	if net < 0 {
		return 0, -net, nil
	}
	return net, 0, nil
}
