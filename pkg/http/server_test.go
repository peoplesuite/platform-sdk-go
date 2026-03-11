package http

import (
	"context"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestDefaultServerConfigAndNewServer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	cfg := DefaultServerConfig(handler, 8080)

	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.Handler == nil {
		t.Fatalf("Handler is nil")
	}

	srv := NewServer(cfg)
	if srv.Addr != ":8080" {
		t.Fatalf("Addr = %q, want %q", srv.Addr, ":8080")
	}
	if srv.Handler == nil {
		t.Fatalf("Handler not wired correctly")
	}
}

func TestServe_ShutdownOnContextCancel(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{
		Handler: handler,
		Addr:    ":0",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	logger := zaptest.NewLogger(t)
	go func() {
		done <- Serve(ctx, srv, logger)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	err := <-done
	if err != nil && err != context.Canceled {
		t.Fatalf("Serve returned unexpected error: %v", err)
	}
}

func TestServe_ListenError(t *testing.T) {
	srv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		Addr:    ":999999", // invalid port (out of 1-65535 range) so Listen fails
	}
	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	err := Serve(ctx, srv, logger)
	if err == nil {
		t.Fatal("expected error when ListenAndServe fails, got nil")
	}
}
