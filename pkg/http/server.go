package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ServerConfig holds options for building an HTTP server.
type ServerConfig struct {
	Handler         http.Handler
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DefaultServerConfig returns sensible defaults.
func DefaultServerConfig(handler http.Handler, port int) ServerConfig {
	return ServerConfig{
		Handler:         handler,
		Port:            port,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
}

// NewServer creates an *http.Server from config.
func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      cfg.Handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

// Serve starts the HTTP server and handles graceful shutdown via context.
func Serve(ctx context.Context, srv *http.Server, logger *zap.Logger) error {
	errCh := make(chan error, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		logger.Info("HTTP server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		<-done
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		logger.Info("HTTP server shutting down")
		if err := srv.Shutdown(shutCtx); err != nil {
			<-done
			return err
		}
		<-done
		return nil
	}
}
