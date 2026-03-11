package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"peoplesuite/platform-sdk-go/pkg/observability"
)

func MetricsUnary(m *observability.Metrics) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		m.ActiveRequests.WithLabelValues(info.FullMethod).Inc()
		defer m.ActiveRequests.WithLabelValues(info.FullMethod).Dec()

		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		code := status.Code(err)

		m.RequestsTotal.WithLabelValues(info.FullMethod, code.String()).Inc()
		m.RequestDuration.WithLabelValues(info.FullMethod).Observe(duration.Seconds())

		if err != nil {
			m.ErrorsTotal.WithLabelValues(info.FullMethod, code.String()).Inc()
		}

		return resp, err
	}
}

func MetricsStream(m *observability.Metrics) grpc.StreamServerInterceptor {

	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		m.ActiveRequests.WithLabelValues(info.FullMethod).Inc()
		defer m.ActiveRequests.WithLabelValues(info.FullMethod).Dec()

		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start)
		code := status.Code(err)

		m.RequestsTotal.WithLabelValues(info.FullMethod, code.String()).Inc()
		m.RequestDuration.WithLabelValues(info.FullMethod).Observe(duration.Seconds())

		if err != nil {
			m.ErrorsTotal.WithLabelValues(info.FullMethod, code.String()).Inc()
		}

		return err
	}
}
