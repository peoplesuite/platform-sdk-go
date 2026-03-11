package logging

import (
	"testing"

	"go.uber.org/zap"
)

func TestMethod(t *testing.T) {
	f := Method("GET")
	if f.Key != "http.method" {
		t.Errorf("Method key = %q, want http.method", f.Key)
	}
}

func TestPath(t *testing.T) {
	f := Path("/foo")
	if f.Key != "http.path" {
		t.Errorf("Path key = %q, want http.path", f.Key)
	}
}

func TestStatus(t *testing.T) {
	f := Status(200)
	if f.Key != "http.status" {
		t.Errorf("Status key = %q, want http.status", f.Key)
	}
}

func TestDuration(t *testing.T) {
	f := Duration("1s")
	if f.Key != "duration" {
		t.Errorf("Duration key = %q, want duration", f.Key)
	}
}

func TestRequestID(t *testing.T) {
	f := RequestID("abc")
	if f.Key != "request_id" {
		t.Errorf("RequestID key = %q, want request_id", f.Key)
	}
}

func TestUserID(t *testing.T) {
	f := UserID("u1")
	if f.Key != "user_id" {
		t.Errorf("UserID key = %q, want user_id", f.Key)
	}
}

// Ensure fields can be used with zap (smoke test)
func TestFields_WithLogger(t *testing.T) {
	logger := zap.NewNop()
	logger.Info("msg", Method("GET"), Path("/"), Status(200), RequestID("id"), UserID("u"))
}
