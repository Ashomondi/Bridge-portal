package billing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ashomondi/ikonex-academy/backend/internal/billing"
	"github.com/Ashomondi/ikonex-academy/backend/internal/billing/testenv"
	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

func TestFeeStructureValidationAndLoading(t *testing.T) {
	db := testenv.New(t, true)
	schoolA := db.SeedSchool(t, "01")
	schoolB := db.SeedSchool(t, "02")
	ctx := context.Background()

	var otherYearID string
	if err := db.Admin.QueryRow(ctx, `INSERT INTO academic_years(school_id,year,is_current) VALUES($1,2027,false) RETURNING id`, schoolA.ID).Scan(&otherYearID); err != nil {
		t.Fatal(err)
	}

	svc := &billing.Service{DB: db}
	adminPrincipal := contracts.Principal{
		UserID:       schoolA.UserID,
		SchoolID:     schoolA.ID,
		MembershipID: schoolA.MembershipID,
		Role:         "ikonex_admin",
	}

	rawAmt := func(s string) json.RawMessage {
		return json.RawMessage(`"` + s + `"`)
	}

	// Money edge cases required by BILL-02
	if m, err := billing.ParseMoney("0.1"); err != nil || m != 10 {
		t.Fatalf("expected 0.1 -> 10 cents, got %d (%v)", m, err)
	}
	for _, bad := range []string{"10.005", "99999999999.99"} {
		if _, err := billing.ParseMoney(bad); err == nil {
			t.Fatalf("expected %q to be rejected", bad)
		}
	}

	missingUUID := "99999999-9999-4999-8999-999999999999"

	cases := []struct {
		name      string
		role      string
		input     billing.LoadFeeStructureInput
		wantCode  int
		wantError string
		wantField string
	}{
		{
			name: "forbidden non-admin role",
			role: "parent",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10000.00")}},
			},
			wantCode:  http.StatusForbidden,
			wantError: "FORBIDDEN",
		},
		{
			name: "missing academic year",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: missingUUID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10000.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_ACADEMIC_YEAR",
			wantField: "academic_year_id",
		},
		{
			name: "academic year from another school",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolB.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10000.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_ACADEMIC_YEAR",
			wantField: "academic_year_id",
		},
		{
			name: "missing term",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         missingUUID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10000.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_TERM",
			wantField: "term_id",
		},
		{
			name: "term from another school",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolB.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10000.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_TERM",
			wantField: "term_id",
		},
		{
			name: "term/year mismatch",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: otherYearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10000.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "TERM_ACADEMIC_YEAR_MISMATCH",
			wantField: "term_id",
		},
		{
			name: "missing grade",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   missingUUID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10000.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_GRADE_LEVEL",
			wantField: "grade_level_id",
		},
		{
			name: "grade from another school",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolB.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10000.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_GRADE_LEVEL",
			wantField: "grade_level_id",
		},
		{
			name: "no fee items",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          nil,
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "FEE_ITEMS_REQUIRED",
			wantField: "items",
		},
		{
			name: "empty item name",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "   ", Amount: rawAmt("100.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_FEE_ITEM_NAME",
			wantField: "items[0].title",
		},
		{
			name: "duplicate item names",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items: []billing.FeeItemInput{
					{Title: "Tuition", Amount: rawAmt("100.00")},
					{Title: "Tuition", Amount: rawAmt("200.00")},
				},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "DUPLICATE_FEE_ITEM_NAME",
			wantField: "items[1].title",
		},
		{
			name: "case-insensitive duplicate item names",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items: []billing.FeeItemInput{
					{Title: "Tuition", Amount: rawAmt("100.00")},
					{Title: "  tuition ", Amount: rawAmt("200.00")},
				},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "DUPLICATE_FEE_ITEM_NAME",
			wantField: "items[1].title",
		},
		{
			name: "zero amount",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("0.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_FEE_ITEM_AMOUNT",
			wantField: "items[0].amount",
		},
		{
			name: "negative amount",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("-10.00")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_FEE_ITEM_AMOUNT",
			wantField: "items[0].amount",
		},
		{
			name: "more than two decimal places",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("10.005")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_FEE_ITEM_AMOUNT",
			wantField: "items[0].amount",
		},
		{
			name: "individual amount overflow",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("99999999999.99")}},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "INVALID_FEE_ITEM_AMOUNT",
			wantField: "items[0].amount",
		},
		{
			name: "total amount overflow",
			role: "ikonex_admin",
			input: billing.LoadFeeStructureInput{
				AcademicYearID: schoolA.YearID,
				TermID:         schoolA.TermID,
				GradeLevelID:   schoolA.GradeID,
				Items: []billing.FeeItemInput{
					{Title: "Item A", Amount: rawAmt("6000000000.00")},
					{Title: "Item B", Amount: rawAmt("5000000000.00")},
				},
			},
			wantCode:  http.StatusUnprocessableEntity,
			wantError: "FEE_STRUCTURE_TOTAL_OVERFLOW",
			wantField: "items",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := adminPrincipal
			p.Role = tc.role
			_, err := svc.LoadFeeStructure(ctx, p, tc.input)
			var apiErr *contracts.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected APIError, got %v", err)
			}
			if apiErr.Status != tc.wantCode || apiErr.Code != tc.wantError {
				t.Fatalf("got (%d, %s), want (%d, %s)", apiErr.Status, apiErr.Code, tc.wantCode, tc.wantError)
			}
			if tc.wantField != "" && apiErr.Details["field"] != tc.wantField {
				t.Fatalf("got field %v, want %q", apiErr.Details["field"], tc.wantField)
			}
		})
	}

	// Valid structure with unsorted items -> deterministic ordering, period upsert, identical replay, and conflict
	validInput := billing.LoadFeeStructureInput{
		AcademicYearID: schoolA.YearID,
		TermID:         schoolA.TermID,
		GradeLevelID:   schoolA.GradeID,
		Items: []billing.FeeItemInput{
			{Title: "Transport", Amount: rawAmt("5000.00")},
			{Title: "Activity Fee", Amount: rawAmt("1500.5")},
			{Title: "Tuition", Amount: rawAmt("20000.00")},
		},
	}
	created, err := svc.LoadFeeStructure(ctx, adminPrincipal, validInput)
	if err != nil {
		t.Fatalf("LoadFeeStructure valid: %v", err)
	}
	if created.TotalAmount.String() != "26500.50" || created.DueDate != "2026-01-05" || created.BoardingStatus != "all" {
		t.Fatalf("unexpected structure metadata: %+v", created)
	}
	if len(created.Items) != 3 || created.Items[0].Title != "Activity Fee" || created.Items[1].Title != "Transport" || created.Items[2].Title != "Tuition" {
		t.Fatalf("items not deterministically ordered: %+v", created.Items)
	}

	// Identical replay (even with reversed item input order) returns existing structure idempotently
	replayInput := billing.LoadFeeStructureInput{
		AcademicYearID: schoolA.YearID,
		TermID:         schoolA.TermID,
		GradeLevelID:   schoolA.GradeID,
		BoardingStatus: "ALL",
		Items: []billing.FeeItemInput{
			{Title: "Tuition", Amount: rawAmt("20000.00")},
			{Title: "Activity Fee", Amount: rawAmt("1500.50")},
			{Title: "Transport", Amount: rawAmt("5000")},
		},
	}
	replayed, err := svc.LoadFeeStructure(ctx, adminPrincipal, replayInput)
	if err != nil {
		t.Fatalf("identical replay failed: %v", err)
	}
	if replayed.ID != created.ID || len(replayed.Items) != 3 {
		t.Fatalf("replay mismatch: got %+v, want ID %s", replayed, created.ID)
	}

	// Conflicting replay returns 409 FEE_STRUCTURE_CONFLICT
	conflictInput := validInput
	conflictInput.Items = []billing.FeeItemInput{
		{Title: "Transport", Amount: rawAmt("6000.00")},
		{Title: "Activity Fee", Amount: rawAmt("1500.50")},
		{Title: "Tuition", Amount: rawAmt("20000.00")},
	}
	_, err = svc.LoadFeeStructure(ctx, adminPrincipal, conflictInput)
	var apiErr *contracts.APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusConflict || apiErr.Code != "FEE_STRUCTURE_CONFLICT" {
		t.Fatalf("expected 409 FEE_STRUCTURE_CONFLICT, got %v", err)
	}

	// Rollback on mid-transaction failure (audit failure rolls back structure, items, and fee_period)
	failingSvc := &billing.Service{
		DB:    db,
		Audit: failingAuditLogger{},
	}
	_, err = failingSvc.LoadFeeStructure(ctx, contracts.Principal{
		UserID:       schoolB.UserID,
		SchoolID:     schoolB.ID,
		MembershipID: schoolB.MembershipID,
		Role:         "ikonex_admin",
	}, billing.LoadFeeStructureInput{
		AcademicYearID: schoolB.YearID,
		TermID:         schoolB.TermID,
		GradeLevelID:   schoolB.GradeID,
		Items:          []billing.FeeItemInput{{Title: "Tuition", Amount: rawAmt("12000.00")}},
	})
	if err == nil {
		t.Fatal("expected rollback error on audit failure")
	}
	var count int
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM fee_structures WHERE school_id = $1`, schoolB.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("expected 0 structures in schoolB after rollback, got %d (%v)", count, err)
	}
	if err := db.Admin.QueryRow(ctx, `SELECT count(*) FROM fee_periods WHERE school_id = $1`, schoolB.ID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("expected 0 fee_periods in schoolB after rollback, got %d (%v)", count, err)
	}

	// Verify HTTP route and rejects school_id in JSON body (DisallowUnknownFields)
	mux := http.NewServeMux()
	billing.RegisterRoutes(mux, svc, func(*http.Request) (contracts.Principal, error) {
		return adminPrincipal, nil
	})
	badBody := `{"school_id":"` + schoolB.ID + `","academic_year_id":"` + schoolA.YearID + `","term_id":"` + schoolA.TermID + `","grade_level_id":"` + schoolA.GradeID + `","items":[{"title":"Tuition","amount":"100.00"}]}`
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/billing/fee-structures", bytes.NewBufferString(badBody)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when untrusted school_id supplied in body, got %d: %s", rec.Code, rec.Body.String())
	}
}

type failingAuditLogger struct{}

func (failingAuditLogger) Log(context.Context, pgx.Tx, contracts.AuditEvent) error {
	return errors.New("simulated audit failure")
}
