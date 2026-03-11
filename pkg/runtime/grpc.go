package runtime

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type grpcServer struct {
	server *grpc.Server
	port   int
	logger *zap.Logger
}

func newGRPCServer(opts Options, logger *zap.Logger) *grpcServer {

	if opts.GRPCServer == nil {
		return nil
	}

	return &grpcServer{
		server: opts.GRPCServer,
		port:   opts.GRPCPort,
		logger: logger,
	}
}

func (r *Runtime) startGRPC(ctx context.Context, errCh chan<- error) {

	if r.grpcServer == nil {
		return
	}

	addr := fmt.Sprintf(":%d", r.grpcServer.port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		errCh <- err
		return
	}

	go func() {

		r.logger.Info("gRPC server starting", zap.Int("port", r.grpcServer.port))

		if err := r.grpcServer.server.Serve(lis); err != nil {
			errCh <- err
		}

	}()

	go func() {
		<-ctx.Done()
		r.grpcServer.server.GracefulStop()
	}()
}
