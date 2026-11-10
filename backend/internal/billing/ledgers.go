package billing

import (
	"context"
	"net/http"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

type DarajaProvider interface{}

type STKPollQueue interface{}

func chargeGradeForTerm(ctx context.Context, tx pgx.Tx, structureID string) error {
	return nil
}

func chargeStudent(ctx context.Context, tx pgx.Tx, studentID, feePeriodID string) error {
	return nil
}

func chargeAdmittedStudent(ctx context.Context, tx pgx.Tx, schoolID, studentID, academicYearID string) error {
	return nil
}

func (s *Service) OnTermStartDateChanged(ctx context.Context, tx pgx.Tx, schoolID, termID string, newStartDate time.Time, actorUserID, note string) error {
	return nil
}

func registerLedgerRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity)    {}
func registerFeeChangeRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {}
func registerMpesaRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity)     {}
