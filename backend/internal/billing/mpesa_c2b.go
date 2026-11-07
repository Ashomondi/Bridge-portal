package billing

import "net/http"

type C2BConfirmationPayload struct {
	TransactionType   string `json:"TransactionType"`
	TransID           string `json:"TransID"`
	TransTime         string `json:"TransTime"`
	TransAmount       string `json:"TransAmount"`
	BusinessShortCode string `json:"BusinessShortCode"`
	BillRefNumber     string `json:"BillRefNumber"`
	MSISDN            string `json:"MSISDN"`
	FirstName         string `json:"FirstName"`
}

func (s *Service) handleC2BConfirmationHTTP(w http.ResponseWriter, r *http.Request) {}
