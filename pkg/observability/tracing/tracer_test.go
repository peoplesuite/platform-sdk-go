package tracing

import (
	"testing"
)

func TestInit_NoPanic(t *testing.T) {
	cfg := Config{
		ServiceName:    "test-svc",
		ServiceVersion: "0.0.0",
		Environment:    "test",
	}
	stop := Init(cfg)
	if stop == nil {
		t.Fatal("Init returned nil stop func")
	}
	// Stop should not panic
	stop()
}

func TestInit_StopIdempotent(t *testing.T) {
	cfg := Config{ServiceName: "x", ServiceVersion: "0", Environment: "test"}
	stop := Init(cfg)
	stop()
	stop() // second call should not panic
}
