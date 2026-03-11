package interceptors

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestTracingUnary_PassesThrough(t *testing.T) {
	interceptor := TracingUnary()
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	resp, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
}

func TestTracingStream_PassesThrough(t *testing.T) {
	interceptor := TracingStream()
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
