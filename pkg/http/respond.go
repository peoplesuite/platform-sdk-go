package http

import (
	"encoding/json"
	"net/http"

	pkgerr "github.com/peoplesuite/platform-sdk-go/pkg/errors"
)

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// OK writes a 200 JSON response.
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

// Created writes a 201 JSON response.
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 response with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ErrorResponse is the standard error body.
type ErrorResponse struct {
	Error     string `json:"error"`
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorJSON writes a JSON error response.
func ErrorJSON(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Error: message})
}

// RespondError maps a pkg/errors.Error to the appropriate HTTP response.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	status := pkgerr.HTTPStatus(err)
	msg := pkgerr.HTTPMessage(err)
	reqID := GetRequestID(r.Context())

	JSON(w, status, ErrorResponse{
		Error:     msg,
		Code:      pkgerr.GetKind(err).String(),
		RequestID: reqID,
	})
}
