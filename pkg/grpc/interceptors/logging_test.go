package interceptors

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLoggingUnary_LogsAndPassesThrough(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := LoggingUnary(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	resp, err := interceptor(context.Background(), "req", info, handler)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
}

func TestLoggingUnary_LogsError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := LoggingUnary(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "not found")
	})

	resp, err := interceptor(context.Background(), nil, info, handler)
	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}
	if err == nil {
		t.Fatal("err is nil")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("code = %v, want NotFound", status.Code(err))
	}
}

func TestLoggingStream_LogsAndPassesThrough(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := LoggingStream(logger)
	info := &grpc.StreamServerInfo{FullMethod: "/svc/Stream"}
	handler := grpc.StreamHandler(func(srv any, ss grpc.ServerStream) error {
		return nil
	})
	ss := &mockServerStream{ctx: context.Background()}

	err := interceptor(nil, ss, info, handler)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
}
