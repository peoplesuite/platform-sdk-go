package runtime

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// Runtime runs HTTP server, gRPC server, and workers until context is cancelled.
type Runtime struct {
	opts   Options
	logger *zap.Logger

	httpServer *httpServer
	grpcServer *grpcServer
}

// New creates a Runtime with the given options.
func New(opts Options) (*Runtime, error) {

	logger, err := initLogger(opts)
	if err != nil {
		return nil, err
	}

	rt := &Runtime{
		opts:   opts,
		logger: logger,
	}

	rt.httpServer = newHTTPServer(opts, logger)
	rt.grpcServer = newGRPCServer(opts, logger)

	return rt, nil
}

// Run starts HTTP, gRPC, and workers; it blocks until ctx is cancelled and then shuts down.
func (r *Runtime) Run(ctx context.Context) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)

	if err := runStartHooks(ctx, r.opts.StartHooks); err != nil {
		return err
	}

	r.startHTTP(ctx, errCh)
	r.startGRPC(ctx, errCh)
	r.startWorkers(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {

	case sig := <-sigCh:
		r.logger.Info("shutdown signal received", zap.String("signal", sig.String()))

	case err := <-errCh:
		return err

	case <-ctx.Done():
	}

	return r.shutdown(context.Background())
}
