package grpc

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"peoplesuite/platform-sdk-go/pkg/grpc/interceptors"
	"peoplesuite/platform-sdk-go/pkg/observability"
)

type ServerConfig struct {
	Logger           *zap.Logger
	Metrics          *observability.Metrics
	EnableReflection bool
	AuthFunc         interceptors.AuthFunc

	ExtraUnary  []grpc.UnaryServerInterceptor
	ExtraStream []grpc.StreamServerInterceptor
}

func NewServer(cfg ServerConfig) *grpc.Server {

	var unary []grpc.UnaryServerInterceptor
	var stream []grpc.StreamServerInterceptor

	// base interceptors
	unary = append(unary,
		interceptors.RecoveryUnary(cfg.Logger),
		interceptors.TracingUnary(),
	)

	stream = append(stream,
		interceptors.RecoveryStream(cfg.Logger),
		interceptors.TracingStream(),
	)

	// metrics
	if cfg.Metrics != nil {

		unary = append(unary,
			interceptors.MetricsUnary(cfg.Metrics),
		)

		stream = append(stream,
			interceptors.MetricsStream(cfg.Metrics),
		)
	}

	// logging
	unary = append(unary,
		interceptors.LoggingUnary(cfg.Logger),
	)

	stream = append(stream,
		interceptors.LoggingStream(cfg.Logger),
	)

	// authentication
	if cfg.AuthFunc != nil {

		unary = append(unary,
			interceptors.AuthUnary(cfg.AuthFunc, cfg.Logger),
		)
	}

	// user interceptors
	unary = append(unary, cfg.ExtraUnary...)
	stream = append(stream, cfg.ExtraStream...)

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unary...),
		grpc.ChainStreamInterceptor(stream...),
	)

	RegisterHealth(srv)

	if cfg.EnableReflection {
		reflection.Register(srv)
	}

	return srv
}

func Serve(
	srv *grpc.Server,
	port int,
	logger *zap.Logger,
) error {

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}

	logger.Info(
		"gRPC server started",
		zap.Int("port", port),
	)

	return srv.Serve(lis)
}
