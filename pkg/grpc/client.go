package grpc

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	Address string
	Logger  *zap.Logger

	Timeout time.Duration

	UnaryInterceptors  []grpc.UnaryClientInterceptor
	StreamInterceptors []grpc.StreamClientInterceptor
}

// NewClientConn creates a gRPC client connection with
// tracing and optional interceptors.
func NewClientConn(cfg ClientConfig) (*grpc.ClientConn, error) {
	unary := append([]grpc.UnaryClientInterceptor{}, cfg.UnaryInterceptors...)
	stream := append([]grpc.StreamClientInterceptor{}, cfg.StreamInterceptors...)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
	if len(unary) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(unary...))
	}
	if len(stream) > 0 {
		opts = append(opts, grpc.WithChainStreamInterceptor(stream...))
	}

	conn, err := grpc.NewClient(cfg.Address, opts...)

	if err != nil {
		return nil, err
	}

	if cfg.Logger != nil {
		cfg.Logger.Info("gRPC client connected",
			zap.String("address", cfg.Address),
		)
	}

	return conn, nil
}

// InvokeUnary is a helper to call a unary RPC with timeout.
func InvokeUnary(
	ctx context.Context,
	timeout time.Duration,
	call func(ctx context.Context) error,
) error {

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return call(ctx)
}
