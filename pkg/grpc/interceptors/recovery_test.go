package interceptors

import (
	"context"
	"testing"

	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestRecoveryUnary_PanicReturnsInternal(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := RecoveryUnary(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Bar"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		panic("test panic")
	})

	resp, err := interceptor(context.Background(), nil, info, handler)
	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}
	if err == nil {
		t.Fatal("err is nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("err is not status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("code = %v, want Internal", st.Code())
	}
	if st.Message() != "internal error" {
		t.Errorf("message = %q, want %q", st.Message(), "internal error")
	}
}

func TestRecoveryUnary_NoPanicPassesThrough(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := RecoveryUnary(logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Bar"}
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

func TestRecoveryStream_PanicReturnsInternal(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := RecoveryStream(logger)
	info := &grpc.StreamServerInfo{FullMethod: "/test/Stream"}
	handler := grpc.StreamHandler(func(srv any, ss grpc.ServerStream) error {
		panic("stream panic")
	})

	// Use a minimal stream that implements ServerStream
	ctx := context.Background()
	ss := &mockServerStream{ctx: ctx}
	err := interceptor(nil, ss, info, handler)
	if err == nil {
		t.Fatal("err is nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("err is not status: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("code = %v, want Internal", st.Code())
	}
}

type mockServerStream struct {
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context        { return m.ctx }
func (m *mockServerStream) RecvMsg(msg any) error           { return nil }
func (m *mockServerStream) SendMsg(msg any) error           { return nil }
func (m *mockServerStream) SetHeader(md metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(md metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(md metadata.MD)       {}
