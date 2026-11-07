package academics

import (
	"context"
	"net/http"
	"strings"
)

// CBC performance levels, highest to lowest (C43).
const (
	LevelEE = "EE" // Exceeding Expectations
	LevelME = "ME" // Meeting Expectations
	LevelAE = "AE" // Approaching Expectations
	LevelBE = "BE" // Below Expectations
)

// levelValue maps a CBC level to the 4..1 value used by the performance graph.
// It is deliberately not a rank or position.
func levelValue(level string) int {
	switch level {
	case LevelEE:
		return 4
	case LevelME:
		return 3
	case LevelAE:
		return 2
	case LevelBE:
		return 1
	default:
		return 0
	}
}

// resolveCBCLevel converts a numeric score into a CBC level using one locked
// grading-scale version. It never consults the currently active scale.
func resolveCBCLevel(scale GradingScale, score float64) string {
	switch {
	case score >= scale.EEMin:
		return LevelEE
	case score >= scale.MEMin:
		return LevelME
	case score >= scale.AEMin:
		return LevelAE
	default:
		return LevelBE
	}
}

// GradingScaleRequest is the payload for PUT /academics/grading-scales.
type GradingScaleRequest struct {
	EEMin float64 `json:"ee_min"`
	MEMin float64 `json:"me_min"`
	AEMin float64 `json:"ae_min"`
	BEMin float64 `json:"be_min"`
}

// GradingScaleList is returned by GET /academics/grading-scales.
type GradingScaleList struct {
	Active *GradingScale  `json:"active,omitempty"`
	Scales []GradingScale `json:"scales"`
}

// ListGradingScales returns every immutable version for the school.
func (s *Service) ListGradingScales(ctx context.Context) (GradingScaleList, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return GradingScaleList{}, err
	}
	var out GradingScaleList
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		scales, err := tx.ListGradingScales(ctx, actor.SchoolID)
		if err != nil {
			return err
		}
		out.Scales = scales
		if active, err := tx.ActiveGradingScale(ctx, actor.SchoolID); err == nil {
			copy := active
			out.Active = &copy
		}
		return nil
	})
	return out, err
}

// CreateGradingScaleVersion inserts a new version and deactivates the previous
// one. Already-released terms keep their locked version, so past CBC levels do
// not change (AC-19).
func (s *Service) CreateGradingScaleVersion(ctx context.Context, req GradingScaleRequest) (GradingScale, error) {
	actor, err := s.actor(ctx)
	if err != nil {
		return GradingScale{}, err
	}
	if !s.canManage(actor, PermGradingScaleManage) {
		return GradingScale{}, errForbidden(CodeForbidden, "Only the Dean of Studies may edit the CBC grading scale.")
	}
	if err := validateGradingScale(req); err != nil {
		return GradingScale{}, err
	}

	var created GradingScale
	err = s.tx(ctx, actor.SchoolID, func(tx Store) error {
		existing, err := tx.ListGradingScales(ctx, actor.SchoolID)
		if err != nil {
			return err
		}
		version := 1
		for _, g := range existing {
			if g.VersionNumber >= version {
				version = g.VersionNumber + 1
			}
		}
		created, err = tx.CreateGradingScale(ctx, actor.SchoolID, GradingScale{
			VersionNumber: version,
			EEMin:         req.EEMin,
			MEMin:         req.MEMin,
			AEMin:         req.AEMin,
			BEMin:         req.BEMin,
			IsActive:      true,
			CreatedBy:     actor.UserID,
		})
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, actor, "grading_scale_version_created", "grading_scale", created.ID, map[string]any{
			"version_number": version,
			"ee_min":         req.EEMin,
			"me_min":         req.MEMin,
			"ae_min":         req.AEMin,
			"be_min":         req.BEMin,
		})
	})
	return created, err
}

func validateGradingScale(req GradingScaleRequest) *APIError {
	bands := []float64{req.BEMin, req.AEMin, req.MEMin, req.EEMin}
	for _, v := range bands {
		if v < 0 || v > 100 {
			return errValidation("Grading scale thresholds must be between 0 and 100.")
		}
	}
	if !(req.EEMin > req.MEMin && req.MEMin > req.AEMin && req.AEMin > req.BEMin) {
		return errValidation("Grading scale thresholds must be strictly descending: ee_min > me_min > ae_min > be_min.")
	}
	if req.EEMin > 100 || req.BEMin < 0 {
		return errValidation("Grading scale thresholds must be between 0 and 100.")
	}
	return nil
}

func (s *Service) handleListGradingScales(w http.ResponseWriter, r *http.Request) {
	result, err := s.ListGradingScales(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (s *Service) handleCreateGradingScale(w http.ResponseWriter, r *http.Request) {
	var req GradingScaleRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, err)
		return
	}
	result, err := s.CreateGradingScaleVersion(r.Context(), req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

// gradingScaleForRelease returns the scale locked to a release, falling back to
// the active scale only for a first release.
func gradingScaleForRelease(ctx context.Context, tx Store, schoolID, versionID string) (GradingScale, error) {
	if strings.TrimSpace(versionID) != "" {
		return tx.GradingScaleByID(ctx, schoolID, versionID)
	}
	return tx.ActiveGradingScale(ctx, schoolID)
}

func init() {
	registerRoute(http.MethodGet, "/api/v1/academics/grading-scales", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleListGradingScales(w, r)
	})
	registerRoute(http.MethodPut, "/api/v1/academics/grading-scales", func(s *Service, w http.ResponseWriter, r *http.Request) {
		s.handleCreateGradingScale(w, r)
	})
}
