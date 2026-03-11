package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgerr "peoplesuite/platform-sdk-go/pkg/errors"
)

func TestJSON_WithBodyAndNil(t *testing.T) {
	rr := httptest.NewRecorder()
	JSON(rr, http.StatusCreated, map[string]string{"ok": "yes"})

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if ct := rr.Header().Get("Content-Type"); ct == "" {
		t.Fatalf("missing Content-Type header")
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["ok"] != "yes" {
		t.Fatalf("body[ok] = %q, want %q", body["ok"], "yes")
	}

	rr2 := httptest.NewRecorder()
	JSON(rr2, http.StatusNoContent, nil)
	if rr2.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr2.Code, http.StatusNoContent)
	}
}

func TestOKCreatedNoContentHelpers(t *testing.T) {
	rr := httptest.NewRecorder()
	OK(rr, map[string]string{"ok": "yes"})
	if rr.Code != http.StatusOK {
		t.Fatalf("OK status = %d, want %d", rr.Code, http.StatusOK)
	}

	rr2 := httptest.NewRecorder()
	Created(rr2, map[string]string{"ok": "yes"})
	if rr2.Code != http.StatusCreated {
		t.Fatalf("Created status = %d, want %d", rr2.Code, http.StatusCreated)
	}

	rr3 := httptest.NewRecorder()
	NoContent(rr3)
	if rr3.Code != http.StatusNoContent {
		t.Fatalf("NoContent status = %d, want %d", rr3.Code, http.StatusNoContent)
	}
}

func TestErrorJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	ErrorJSON(rr, http.StatusBadRequest, "bad")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var body ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.Error != "bad" {
		t.Fatalf("Error field = %q, want %q", body.Error, "bad")
	}
}

func TestRespondError_UsesPlatformErrorAndRequestID(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)

	// Add request ID via middleware helper.
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := pkgerr.NotFound("missing")
		RespondError(w, r, err)
	})).ServeHTTP

	handler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}

	var body ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.Error != "missing" {
		t.Fatalf("Error = %q, want %q", body.Error, "missing")
	}
	if body.Code != pkgerr.KindNotFound.String() {
		t.Fatalf("Code = %q, want %q", body.Code, pkgerr.KindNotFound.String())
	}
	if body.RequestID == "" {
		t.Fatalf("expected non-empty RequestID")
	}
}
