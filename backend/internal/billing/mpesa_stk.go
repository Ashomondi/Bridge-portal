package billing

import (
	"context"
	"net/http"
)

// STKPushRequest payload
type STKPushRequest struct {
	SchoolID        string  `json:"school_id"`
	StudentID       string  `json:"student_id"`
	Amount          float64 `json:"amount"`
	PhoneNumber     string  `json:"phone_number"`
	Source          string  `json:"source"` // "portal" or "chat"
	ParentConfirmed bool    `json:"parent_confirmed"`
}

type STKPushResponse struct {
	CheckoutRequestID string `json:"checkout_request_id"`
	ResponseCode      string `json:"response_code"`
	CustomerMessage   string `json:"customer_message"`
}

// InitiateSTKPush handles initiating STK push
func (s *Service) InitiateSTKPush(ctx context.Context, req STKPushRequest) (*STKPushResponse, error) {
	return nil, nil
}

func (s *Service) handleSTKPush(w http.ResponseWriter, r *http.Request) {}
func (s *Service) handleSTKCallbackHTTP(w http.ResponseWriter, r *http.Request) {}
