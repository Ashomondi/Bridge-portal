package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

type FeeItemInput struct {
	Title        string          `json:"title"`
	Name         string          `json:"name"`
	Amount       json.RawMessage `json:"amount"`
	IsCompulsory *bool           `json:"is_compulsory"`
}

type LoadFeeStructureInput struct {
	AcademicYearID string         `json:"academic_year_id"`
	TermID         string         `json:"term_id"`
	GradeLevelID   string         `json:"grade_level_id"`
	BoardingStatus string         `json:"boarding_status"`
	Items          []FeeItemInput `json:"items"`
}

type FeeStructureItem struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Amount       Money  `json:"amount"`
	IsCompulsory bool   `json:"is_compulsory"`
}

type FeeStructure struct {
	ID             string             `json:"id"`
	SchoolID       string             `json:"school_id"`
	AcademicYearID string             `json:"academic_year_id"`
	TermID         string             `json:"term_id"`
	GradeLevelID   string             `json:"grade_level_id"`
	FeePeriodID    string             `json:"fee_period_id"`
	BoardingStatus string             `json:"boarding_status"`
	DueDate        string             `json:"due_date"`
	TotalAmount    Money              `json:"total_amount"`
	Items          []FeeStructureItem `json:"items"`
	CreatedAt      time.Time          `json:"created_at"`
}

type ChargeCurrentResult struct {
	StructuresProcessed int `json:"structures_processed"`
	ChargesCreated      int `json:"charges_created"`
}

type ValidationError struct {
	Field   string
	Code    string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func validationFail(field, code, message string) *contracts.APIError {
	err := contracts.Fail(http.StatusUnprocessableEntity, code, message)
	err.Details["field"] = field
	return err
}

type validatedItem struct {
	Title        string
	Amount       Money
	IsCompulsory bool
}

func normalizeAndValidateStructureInput(in LoadFeeStructureInput) (string, []validatedItem, Money, error) {
	if !contracts.ValidID(in.AcademicYearID) {
		return "", nil, 0, validationFail("academic_year_id", "INVALID_ACADEMIC_YEAR", "A valid academic_year_id is required.")
	}
	if !contracts.ValidID(in.TermID) {
		return "", nil, 0, validationFail("term_id", "INVALID_TERM", "A valid term_id is required.")
	}
	if !contracts.ValidID(in.GradeLevelID) {
		return "", nil, 0, validationFail("grade_level_id", "INVALID_GRADE_LEVEL", "A valid grade_level_id is required.")
	}
	boarding := strings.ToLower(strings.TrimSpace(in.BoardingStatus))
	if boarding == "" {
		boarding = "all"
	}
	switch boarding {
	case "all":
	case "boarder", "day":
		return "", nil, 0, validationFail("boarding_status", "UNSUPPORTED_BOARDING_VARIANT", "Release 1 supports only the 'all' boarding_status fee structure variant.")
	default:
		return "", nil, 0, validationFail("boarding_status", "INVALID_BOARDING_STATUS", "boarding_status must be 'all' in Release 1.")
	}
	if len(in.Items) == 0 {
		return "", nil, 0, validationFail("items", "FEE_ITEMS_REQUIRED", "At least one fee item is required.")
	}
	seen := make(map[string]struct{}, len(in.Items))
	items := make([]validatedItem, 0, len(in.Items))
	var total Money
	for idx, raw := range in.Items {
		title := strings.TrimSpace(raw.Title)
		if title == "" {
			title = strings.TrimSpace(raw.Name)
		}
		fieldPrefix := fmt.Sprintf("items[%d]", idx)
		if title == "" || utf8.RuneCountInString(title) > 255 {
			return "", nil, 0, validationFail(fieldPrefix+".title", "INVALID_FEE_ITEM_NAME", "Fee item title must be between 1 and 255 characters.")
		}
		lowerKey := strings.ToLower(title)
		if _, dup := seen[lowerKey]; dup {
			return "", nil, 0, validationFail(fieldPrefix+".title", "DUPLICATE_FEE_ITEM_NAME", "Fee item names must be unique case-insensitively.")
		}
		seen[lowerKey] = struct{}{}

		if len(raw.Amount) == 0 {
			return "", nil, 0, validationFail(fieldPrefix+".amount", "INVALID_FEE_ITEM_AMOUNT", "Fee item amount is required.")
		}
		var amount Money
		if err := json.Unmarshal(raw.Amount, &amount); err != nil || amount <= 0 {
			return "", nil, 0, validationFail(fieldPrefix+".amount", "INVALID_FEE_ITEM_AMOUNT", "Fee item amount must be a positive decimal within NUMERIC(12,2).")
		}
		nextTotal, err := addMoney(total, amount)
		if err != nil {
			return "", nil, 0, validationFail("items", "FEE_STRUCTURE_TOTAL_OVERFLOW", "Total fee structure amount exceeds NUMERIC(12,2).")
		}
		total = nextTotal
		compulsory := true
		if raw.IsCompulsory != nil {
			compulsory = *raw.IsCompulsory
		}
		items = append(items, validatedItem{
			Title:        title,
			Amount:       amount,
			IsCompulsory: compulsory,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		li := strings.ToLower(items[i].Title)
		lj := strings.ToLower(items[j].Title)
		if li != lj {
			return li < lj
		}
		return items[i].Title < items[j].Title
	})
	return boarding, items, total, nil
}

// LoadFeeStructure validates and transactionally loads a fee structure, its items,
// and the corresponding generic term fee_period. Requires an active school
// membership holding the billing.fee_structures.manage badge permission and a
// configured AuditLogger. If an identical structure already exists for
// (school, year, grade, term, 'all'), it returns the existing record idempotently
// and still triggers term charging when the academic year or term is current.
// If a structure exists with different items or amounts, it returns 409 FEE_STRUCTURE_CONFLICT.
func (s *Service) LoadFeeStructure(ctx context.Context, p contracts.Principal, in LoadFeeStructureInput) (FeeStructure, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) || p.Role == "" {
		return FeeStructure{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if s == nil || s.DB == nil {
		return FeeStructure{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return FeeStructure{}, err
	}
	var out FeeStructure
	err := s.DB.WithTenantTx(ctx, p, func(tx pgx.Tx) error {
		if err := requireSchoolPermissionTx(ctx, tx, p, PermFeeStructuresManage, "Only authorized fee-structure administrators may load fee structures."); err != nil {
			return err
		}
		var err error
		out, err = s.LoadFeeStructureTx(ctx, tx, p.SchoolID, p.UserID, in)
		return err
	})
	return out, err
}

func (s *Service) LoadFeeStructureTx(ctx context.Context, tx pgx.Tx, schoolID, actorUserID string, in LoadFeeStructureInput) (FeeStructure, error) {
	if !contracts.ValidID(schoolID) {
		return FeeStructure{}, contracts.Fail(http.StatusBadRequest, "INVALID_SCHOOL", "A valid school is required.")
	}
	if !contracts.ValidID(actorUserID) {
		return FeeStructure{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "A valid actor user_id is required.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return FeeStructure{}, err
	}
	boarding, items, total, err := normalizeAndValidateStructureInput(in)
	if err != nil {
		return FeeStructure{}, err
	}

	var yearExists, yearCurrent bool
	err = tx.QueryRow(ctx, `SELECT true, is_current FROM academic_years WHERE school_id = $1 AND id = $2`, schoolID, in.AcademicYearID).Scan(&yearExists, &yearCurrent)
	if errors.Is(err, pgx.ErrNoRows) || !yearExists {
		return FeeStructure{}, validationFail("academic_year_id", "INVALID_ACADEMIC_YEAR", "Academic year does not exist in this school.")
	}
	if err != nil {
		return FeeStructure{}, err
	}

	var termYearID string
	var termNumber int
	var termStartDate time.Time
	var termCurrent bool
	err = tx.QueryRow(ctx, `SELECT academic_year_id, term_number, start_date, is_current FROM terms WHERE school_id = $1 AND id = $2`, schoolID, in.TermID).Scan(&termYearID, &termNumber, &termStartDate, &termCurrent)
	if errors.Is(err, pgx.ErrNoRows) {
		return FeeStructure{}, validationFail("term_id", "INVALID_TERM", "Term does not exist in this school.")
	}
	if err != nil {
		return FeeStructure{}, err
	}
	if termYearID != in.AcademicYearID {
		return FeeStructure{}, validationFail("term_id", "TERM_ACADEMIC_YEAR_MISMATCH", "Term does not belong to the specified academic year.")
	}

	var gradeExists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM grade_levels WHERE school_id = $1 AND id = $2)`, schoolID, in.GradeLevelID).Scan(&gradeExists)
	if err != nil {
		return FeeStructure{}, err
	}
	if !gradeExists {
		return FeeStructure{}, validationFail("grade_level_id", "INVALID_GRADE_LEVEL", "Grade level does not exist in this school.")
	}

	dueDateStr := termStartDate.Format("2006-01-02")
	periodName := fmt.Sprintf("Term %d", termNumber)

	var feePeriodID string
	var periodDueDate time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO fee_periods (school_id, academic_year_id, term_id, period_type, name, due_date)
		VALUES ($1, $2, $3, 'term', $4, $5::date)
		ON CONFLICT (school_id, academic_year_id, term_id)
		DO UPDATE SET name = fee_periods.name
		RETURNING id, due_date`,
		schoolID, in.AcademicYearID, in.TermID, periodName, dueDateStr,
	).Scan(&feePeriodID, &periodDueDate)
	if err != nil {
		return FeeStructure{}, err
	}

	totalNumeric, err := total.NumericValue()
	if err != nil {
		return FeeStructure{}, err
	}

	var structureID string
	var createdAt time.Time
	var loadedItems []FeeStructureItem
	inserted := true
	err = tx.QueryRow(ctx, `
		INSERT INTO fee_structures (school_id, academic_year_id, grade_level_id, term_id, boarding_status, total_amount)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (school_id, academic_year_id, grade_level_id, term_id, boarding_status) DO NOTHING
		RETURNING id, created_at`,
		schoolID, in.AcademicYearID, in.GradeLevelID, in.TermID, boarding, totalNumeric,
	).Scan(&structureID, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		inserted = false
		var existingTotal Money
		err = tx.QueryRow(ctx, `
			SELECT id, total_amount, created_at
			FROM fee_structures
			WHERE school_id = $1 AND academic_year_id = $2 AND grade_level_id = $3 AND term_id = $4 AND boarding_status = $5
			FOR UPDATE`,
			schoolID, in.AcademicYearID, in.GradeLevelID, in.TermID, boarding,
		).Scan(&structureID, &existingTotal, &createdAt)
		if err != nil {
			return FeeStructure{}, err
		}
		loadedItems, err = loadStructureItemsDeterministic(ctx, tx, schoolID, structureID)
		if err != nil {
			return FeeStructure{}, err
		}
		if existingTotal != total || !sameStructureItems(loadedItems, items) {
			return FeeStructure{}, contracts.Fail(http.StatusConflict, "FEE_STRUCTURE_CONFLICT", "A fee structure with different items or amounts already exists for this grade and term.")
		}
		total = existingTotal
	} else if err != nil {
		return FeeStructure{}, err
	} else {
		for _, item := range items {
			itemNum, err := item.Amount.NumericValue()
			if err != nil {
				return FeeStructure{}, err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO fee_structure_items (fee_structure_id, title, amount, is_compulsory)
				VALUES ($1, $2, $3, $4)`,
				structureID, item.Title, itemNum, item.IsCompulsory,
			); err != nil {
				return FeeStructure{}, err
			}
		}
		loadedItems, err = loadStructureItemsDeterministic(ctx, tx, schoolID, structureID)
		if err != nil {
			return FeeStructure{}, err
		}
	}

	eligibleToCharge := yearCurrent || termCurrent
	if eligibleToCharge {
		if _, err := s.chargeGradeForTermCount(ctx, tx, structureID, actorUserID); err != nil {
			return FeeStructure{}, err
		}
	}

	if err := s.logAuditPgx(ctx, tx, contracts.AuditEvent{
		SchoolID:     schoolID,
		UserID:       actorUserID,
		Action:       "billing.fee_structure.load",
		ResourceType: "fee_structure",
		ResourceID:   structureID,
		Details: map[string]any{
			"academic_year_id": in.AcademicYearID,
			"term_id":          in.TermID,
			"grade_level_id":   in.GradeLevelID,
			"boarding_status":  boarding,
			"total_amount":     total.String(),
			"items_count":      len(loadedItems),
			"replayed":         !inserted,
			"charged_on_load":  eligibleToCharge,
		},
	}); err != nil {
		return FeeStructure{}, err
	}

	return FeeStructure{
		ID:             structureID,
		SchoolID:       schoolID,
		AcademicYearID: in.AcademicYearID,
		TermID:         in.TermID,
		GradeLevelID:   in.GradeLevelID,
		FeePeriodID:    feePeriodID,
		BoardingStatus: boarding,
		DueDate:        periodDueDate.Format("2006-01-02"),
		TotalAmount:    total,
		Items:          loadedItems,
		CreatedAt:      createdAt.UTC(),
	}, nil
}

// ChargeCurrentTermStructures idempotently charges all preloaded 'all'-variant
// fee structures in the caller's school whose academic year or term is now
// current. Designed for both admin activation triggers and retry-safe scheduled
// background jobs.
func (s *Service) ChargeCurrentTermStructures(ctx context.Context, p contracts.Principal) (ChargeCurrentResult, error) {
	if !contracts.ValidID(p.UserID) || !contracts.ValidID(p.SchoolID) || !contracts.ValidID(p.MembershipID) || p.Role == "" {
		return ChargeCurrentResult{}, contracts.Fail(http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	}
	if s == nil || s.DB == nil {
		return ChargeCurrentResult{}, contracts.Fail(http.StatusInternalServerError, "INTERNAL_ERROR", "The request could not be completed.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return ChargeCurrentResult{}, err
	}
	var out ChargeCurrentResult
	err := s.DB.WithTenantTx(ctx, p, func(tx pgx.Tx) error {
		if err := requireSchoolPermissionTx(ctx, tx, p, PermFeeStructuresManage, "Only authorized fee-structure administrators may trigger term charging."); err != nil {
			return err
		}
		var err error
		out, err = s.ChargeCurrentTermStructuresTx(ctx, tx, p.SchoolID, p.UserID)
		return err
	})
	return out, err
}

// ChargeCurrentTermStructuresTx idempotently evaluates and charges all current
// academic-year or current-term fee structures in schoolID within tx.
func (s *Service) ChargeCurrentTermStructuresTx(ctx context.Context, tx pgx.Tx, schoolID, actorUserID string) (ChargeCurrentResult, error) {
	if !contracts.ValidID(schoolID) {
		return ChargeCurrentResult{}, contracts.Fail(http.StatusBadRequest, "INVALID_SCHOOL", "A valid school is required.")
	}
	if actorUserID != "" && !contracts.ValidID(actorUserID) {
		return ChargeCurrentResult{}, contracts.Fail(http.StatusBadRequest, "INVALID_ACTOR", "actorUserID must be empty or a valid UUID.")
	}
	if err := s.requireAuditLogger(); err != nil {
		return ChargeCurrentResult{}, err
	}

	rows, err := tx.Query(ctx, `
		SELECT fs.id
		FROM fee_structures fs
		JOIN academic_years ay ON ay.school_id = fs.school_id AND ay.id = fs.academic_year_id
		JOIN terms t ON t.school_id = fs.school_id AND t.id = fs.term_id
		JOIN grade_levels gl ON gl.school_id = fs.school_id AND gl.id = fs.grade_level_id
		WHERE fs.school_id = $1
		  AND fs.boarding_status = 'all'
		  AND (ay.is_current = TRUE OR t.is_current = TRUE)
		ORDER BY ay.year ASC, t.term_number ASC, gl.number ASC, fs.id ASC`,
		schoolID,
	)
	if err != nil {
		return ChargeCurrentResult{}, err
	}
	var structureIDs []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			rows.Close()
			return ChargeCurrentResult{}, err
		}
		structureIDs = append(structureIDs, sid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return ChargeCurrentResult{}, err
	}

	var res ChargeCurrentResult
	for _, sid := range structureIDs {
		charged, err := s.chargeGradeForTermCount(ctx, tx, sid, actorUserID)
		if err != nil {
			return ChargeCurrentResult{}, err
		}
		res.StructuresProcessed++
		res.ChargesCreated += charged
	}

	if err := s.logAuditPgx(ctx, tx, contracts.AuditEvent{
		SchoolID:     schoolID,
		UserID:       actorUserID,
		Action:       "billing.fee_structure.charge_current",
		ResourceType: "school",
		ResourceID:   schoolID,
		Details: map[string]any{
			"structures_processed": res.StructuresProcessed,
			"charges_created":      res.ChargesCreated,
		},
	}); err != nil {
		return ChargeCurrentResult{}, err
	}
	return res, nil
}

func loadStructureItemsDeterministic(ctx context.Context, tx pgx.Tx, schoolID, structureID string) ([]FeeStructureItem, error) {
	rows, err := tx.Query(ctx, `
		SELECT i.id, i.title, i.amount, i.is_compulsory
		FROM fee_structure_items i
		JOIN fee_structures fs ON fs.id = i.fee_structure_id
		WHERE fs.school_id = $1 AND i.fee_structure_id = $2
		ORDER BY lower(i.title) ASC, i.title ASC, i.id ASC`,
		schoolID, structureID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []FeeStructureItem{}
	for rows.Next() {
		var it FeeStructureItem
		if err := rows.Scan(&it.ID, &it.Title, &it.Amount, &it.IsCompulsory); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func sameStructureItems(existing []FeeStructureItem, incoming []validatedItem) bool {
	if len(existing) != len(incoming) {
		return false
	}
	for i := range existing {
		if existing[i].Title != incoming[i].Title ||
			existing[i].Amount != incoming[i].Amount ||
			existing[i].IsCompulsory != incoming[i].IsCompulsory {
			return false
		}
	}
	return true
}
