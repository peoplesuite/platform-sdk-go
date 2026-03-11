package interceptors

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func TracingUnary() grpc.UnaryServerInterceptor {

	tracer := otel.Tracer("pkg/grpc")

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		ctx, span := tracer.Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
			),
		)

		defer span.End()

		resp, err := handler(ctx, req)

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(
				attribute.String("rpc.grpc.status_code", status.Code(err).String()),
			)
		} else {
			span.SetAttributes(attribute.String("rpc.grpc.status_code", "OK"))
		}

		return resp, err
	}
}

func TracingStream() grpc.StreamServerInterceptor {

	tracer := otel.Tracer("pkg/grpc")

	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {

		ctx := ss.Context()

		ctx, span := tracer.Start(
			ctx,
			info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
			),
		)

		defer span.End()

		wrapped := &streamCtx{
			ServerStream: ss,
			ctx:          ctx,
		}

		err := handler(srv, wrapped)

		if err != nil {
			span.RecordError(err)
			span.SetAttributes(
				attribute.String("rpc.grpc.status_code", status.Code(err).String()),
			)
		} else {
			span.SetAttributes(attribute.String("rpc.grpc.status_code", "OK"))
		}

		return err
	}
}

type streamCtx struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *streamCtx) Context() context.Context {
	return s.ctx
}
