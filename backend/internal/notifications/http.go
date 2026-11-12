package notifications

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Ashomondi/ikonex-academy/backend/internal/internalapi/contracts"
	"github.com/jackc/pgx/v5"
)

type Notification struct {
	ID        string          `json:"id"`
	Category  string          `json:"category"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Metadata  json.RawMessage `json:"metadata"`
	IsRead    bool            `json:"is_read"`
	ReadAt    *time.Time      `json:"read_at"`
	CreatedAt time.Time       `json:"created_at"`
}

func (s *Service) List(ctx context.Context, p contracts.Principal, page *contracts.Page) ([]Notification, error) {
	notices := []Notification{}
	err := s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, "SELECT count(*) FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2", p.SchoolID, p.UserID).Scan(&page.Total); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, `SELECT id,category,title,body,metadata,is_read,read_at,created_at
			FROM in_app_notifications WHERE school_id=$1 AND recipient_user_id=$2
			ORDER BY created_at DESC,id LIMIT $3 OFFSET $4`, p.SchoolID, p.UserID, page.PerPage, (page.Page-1)*page.PerPage)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var n Notification
			if err := rows.Scan(&n.ID, &n.Category, &n.Title, &n.Body, &n.Metadata, &n.IsRead, &n.ReadAt, &n.CreatedAt); err != nil {
				return err
			}
			n.CreatedAt = n.CreatedAt.UTC()
			if n.ReadAt != nil {
				readAt := n.ReadAt.UTC()
				n.ReadAt = &readAt
			}
			notices = append(notices, n)
		}
		return rows.Err()
	})
	return notices, err
}

func (s *Service) MarkRead(ctx context.Context, p contracts.Principal, id string) error {
	if !contracts.ValidID(id) {
		return contracts.Fail(400, "INVALID_NOTIFICATION_ID", "A valid notification ID is required.")
	}
	return s.DB.WithTenantTx(ctx, p.SchoolID, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, `UPDATE in_app_notifications SET is_read=true,read_at=coalesce(read_at,now())
			WHERE school_id=$1 AND recipient_user_id=$2 AND id=$3`, p.SchoolID, p.UserID, id)
		if err != nil {
			return err
		}
		if result.RowsAffected() != 1 {
			return contracts.Fail(404, "NOTIFICATION_NOT_FOUND", "Notification not found.")
		}
		return nil
	})
}

func RegisterRoutes(mux *http.ServeMux, service *Service, identity contracts.Identity) {
	mux.HandleFunc("GET /api/v1/notifications", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		page, err := contracts.Pagination(r)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		result, err := service.List(r.Context(), p, &page)
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, 200, result, &page)
	})
	mux.HandleFunc("PATCH /api/v1/notifications/{id}/read", func(w http.ResponseWriter, r *http.Request) {
		p, err := contracts.RequireIdentity(r, identity)
		if err == nil {
			err = service.MarkRead(r.Context(), p, r.PathValue("id"))
		}
		if err != nil {
			contracts.WriteError(w, err)
			return
		}
		contracts.Success(w, 200, map[string]any{"id": r.PathValue("id"), "is_read": true}, nil)
	})
}
