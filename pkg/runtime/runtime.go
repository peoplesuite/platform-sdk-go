package runtime

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

type Runtime struct {
	opts   Options
	logger *zap.Logger

	httpServer *httpServer
	grpcServer *grpcServer
}

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
