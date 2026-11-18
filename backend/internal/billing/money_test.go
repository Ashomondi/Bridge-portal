package billing

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestMoneyExactParsingAndJSON(t *testing.T) {
	for _, tc := range []struct{ input, want string }{{"0.01", "0.01"}, {"1.2", "1.20"}, {"12500", "12500.00"}, {"-1.01", "-1.01"}, {"9999999999.99", "9999999999.99"}, {"-9999999999.99", "-9999999999.99"}, {"-0", "0.00"}} {
		value, err := ParseMoney(tc.input)
		if err != nil || value.String() != tc.want {
			t.Fatalf("%s: %s %v", tc.input, value, err)
		}
		for _, input := range []string{tc.input, `"` + tc.input + `"`} {
			var decoded Money
			if err := json.Unmarshal([]byte(input), &decoded); err != nil || decoded != value {
				t.Fatalf("JSON %s: %s %v", input, decoded, err)
			}
		}
		encoded, err := json.Marshal(value)
		if err != nil || string(encoded) != `"`+tc.want+`"` {
			t.Fatalf("serialization: %s %v", encoded, err)
		}
	}
	for _, input := range []string{"", "1.001", "1e2", "+1", " 1", "1.", ".1", "10000000000.00", "NaN", "-", "1..2", "１.００"} {
		if _, err := ParseMoney(input); err == nil {
			t.Fatalf("invalid money accepted: %q", input)
		}
	}
	if _, err := addMoney(MaxMoney, 1); err == nil {
		t.Fatal("overflow accepted")
	}
}

func TestBalancesConserveMoney(t *testing.T) {
	for _, tc := range []struct{ open, billed, paid, current, credit Money }{
		{0, 10000, 5000, 5000, 0}, {0, 10000, 10000, 0, 0}, {0, 10000, 12500, 0, 2500}, {2000, 10000, 15000, 0, 3000}, {0, 0, 1, 0, 1},
	} {
		current, credit, err := balances(tc.open, tc.billed, tc.paid)
		if err != nil || current != tc.current || credit != tc.credit || current-credit != tc.open+tc.billed-tc.paid {
			t.Fatalf("balance %+v: %s %s %v", tc, current, credit, err)
		}
	}
}

func TestMoneyNumericRejectsRoundingAndSpecialValues(t *testing.T) {
	for _, input := range []pgtype.Numeric{{}, {NaN: true, Valid: true}, {Int: big.NewInt(1001), Exp: -3, Valid: true}, {Int: big.NewInt(1), Exp: 20, Valid: true}} {
		var money Money
		if err := money.ScanNumeric(input); err == nil {
			t.Fatal("inexact/invalid numeric accepted")
		}
	}
}

func TestMoneyPostgresRoundTrip(t *testing.T) {
	db := testenv.New(t, false)
	for _, expected := range []Money{0, 1, -1, 123456789, MaxMoney, -MaxMoney} {
		var got Money
		if err := db.Admin.QueryRow(context.Background(), "SELECT $1::numeric(12,2)", expected).Scan(&got); err != nil || got != expected {
			t.Fatalf("NUMERIC round trip: want %s got %s: %v", expected, got, err)
		}
	}
}
