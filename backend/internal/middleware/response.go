package middleware

import (
	"encoding/json"
	"net/http"
)

type successEnvelope struct {
	Status string `json:"status"`
	Data   any    `json:"data"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type errorEnvelope struct {
	Status string      `json:"status"`
	Error  errorDetail `json:"error"`
}

// RespondJSON writes a success envelope: {"status":"success","data":...}.
func RespondJSON(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, successEnvelope{Status: "success", Data: data})
}

// RespondError writes an error envelope: {"status":"error","error":{...}}.
func RespondError(w http.ResponseWriter, status int, code, message string) {
	RespondErrorDetails(w, status, code, message, nil)
}

// RespondErrorDetails writes an error envelope with extra details.
func RespondErrorDetails(w http.ResponseWriter, status int, code, message string, details any) {
	writeJSON(w, status, errorEnvelope{
		Status: "error",
		Error:  errorDetail{Code: code, Message: message, Details: details},
	})
}

// DecodeJSON decodes a JSON request body, responding with 400 on failure.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "The request body is not valid JSON.")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
