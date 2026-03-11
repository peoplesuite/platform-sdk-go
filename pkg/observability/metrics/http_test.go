package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPMiddleware_CallsNextHandler(t *testing.T) {
	m, err := New(Config{ServiceName: "test", Address: "127.0.0.1:19998"})
	if err != nil {
		t.Skipf("New failed: %v", err)
	}
	defer m.Close()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	wrap := HTTPMiddleware(m)(next)
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	rec := httptest.NewRecorder()
	wrap.ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
