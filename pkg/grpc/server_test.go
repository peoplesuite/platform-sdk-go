package grpc

import (
	"testing"

	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"

	"github.com/peoplesuite/platform-sdk-go/pkg/observability"
)

func TestNewServer_NonNil(t *testing.T) {
	logger := zaptest.NewLogger(t)
	srv := NewServer(ServerConfig{
		Logger:           logger,
		EnableReflection: false,
	})
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestNewServer_WithReflection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	srv := NewServer(ServerConfig{
		Logger:           logger,
		EnableReflection: true,
	})
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestNewServer_WithMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	metrics := observability.NewServerMetrics("test")
	srv := NewServer(ServerConfig{
		Logger:           logger,
		Metrics:          metrics,
		EnableReflection: false,
	})
	if srv == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServe_ReturnsWhenServerStopped(t *testing.T) {
	logger := zaptest.NewLogger(t)
	srv := grpc.NewServer()
	RegisterHealth(srv)

	done := make(chan error, 1)
	go func() {
		done <- Serve(srv, 0, logger)
	}()
	srv.GracefulStop()
	err := <-done
	// Serve returns nil on GracefulStop
	if err != nil {
		t.Logf("Serve returned: %v", err)
	}
}
