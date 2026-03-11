package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	"github.com/peoplesuite/platform-sdk-go/pkg/observability"
)

func TestChainOrder(t *testing.T) {
	var calls []string

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "m1-before")
			next.ServeHTTP(w, r)
			calls = append(calls, "m1-after")
		})
	}
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, "m2-before")
			next.ServeHTTP(w, r)
			calls = append(calls, "m2-after")
		})
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, "handler")
	})

	h := Chain(m1, m2)(final)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	want := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	if len(calls) != len(want) {
		t.Fatalf("calls len = %d, want %d", len(calls), len(want))
	}
	for i := range calls {
		if calls[i] != want[i] {
			t.Fatalf("calls[%d] = %q, want %q", i, calls[i], want[i])
		}
	}
}

func TestRequestID_GeneratedAndPreserved(t *testing.T) {
	// No header → generated
	var gotID string
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetRequestID(r.Context())
	})
	h := RequestID(final)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if gotID == "" {
		t.Fatalf("expected generated request ID")
	}
	if hdr := rr.Header().Get(RequestIDHeader); hdr == "" || hdr != gotID {
		t.Fatalf("header %q = %q, want %q", RequestIDHeader, hdr, gotID)
	}

	// Existing header preserved
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set(RequestIDHeader, "fixed-id")
	h.ServeHTTP(rr2, req2)

	if hdr := rr2.Header().Get(RequestIDHeader); hdr != "fixed-id" {
		t.Fatalf("expected fixed-id header, got %q", hdr)
	}
}

func TestCORS_PreflightAndHeaders(t *testing.T) {
	cfg := DefaultCORSConfig()
	h := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Preflight
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want %d", rr.Code, http.StatusNoContent)
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatalf("missing Access-Control-Allow-Methods")
	}

	// Actual request
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Origin", "http://example.com")
	h.ServeHTTP(rr2, req2)
	if rr2.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Fatalf("missing Access-Control-Allow-Origin for actual request")
	}
}

func TestCORS_AllowCredentials(t *testing.T) {
	cfg := DefaultCORSConfig()
	cfg.AllowCredentials = true
	cfg.AllowOrigins = []string{"http://example.com"}
	h := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	h.ServeHTTP(rr, req)
	if rr.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("Access-Control-Allow-Credentials = %q, want true", rr.Header().Get("Access-Control-Allow-Credentials"))
	}
}

func TestLogging_SeverityByStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("x"))
	})

	h := Logging(logger)(handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestLogging_Status500(t *testing.T) {
	logger := zaptest.NewLogger(t)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	h := Logging(logger)(handler)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
}

func TestTracing_RunsRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := Tracing()(handler)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestRecovery_PanicsAreHandled(t *testing.T) {
	logger := zaptest.NewLogger(t)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	h := Recovery(logger)(handler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestMetricsMiddleware_BasicFlow(t *testing.T) {
	m := observability.NewServerMetrics("test")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	h := Metrics(m)(handler)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
