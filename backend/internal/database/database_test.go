package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func TestWithTenantTxRequiresSchoolID(t *testing.T) {
	var pool sql.DB
	err := WithTenantTx(context.Background(), &pool, "", func(*sql.Tx) error { return nil })
	if !errors.Is(err, ErrMissingSchoolID) {
		t.Fatalf("WithTenantTx error = %v, want ErrMissingSchoolID", err)
	}
}

func TestWithTenantTxNilConnection(t *testing.T) {
	err := WithTenantTx(context.Background(), nil, "school-1", func(*sql.Tx) error { return nil })
	if err == nil {
		t.Fatalf("WithTenantTx accepted a nil connection")
	}
}

func TestWithUserTxRequiresUserID(t *testing.T) {
	var pool sql.DB
	err := WithUserTx(context.Background(), &pool, "", func(*sql.Tx) error { return nil })
	if err == nil {
		t.Fatalf("WithUserTx accepted an empty user id")
	}
}
