package interceptors

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthFunc func(ctx context.Context, token string) (context.Context, error)

func AuthUnary(authFunc AuthFunc, logger *zap.Logger) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		token := md.Get("authorization")

		if len(token) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		ctx, err := authFunc(ctx, token[0])
		if err != nil {
			logger.Debug("auth failed", zap.Error(err))
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		return handler(ctx, req)
	}
}
