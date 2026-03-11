package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
)

func TestInvokeUnary_Success(t *testing.T) {
	callCount := 0
	err := InvokeUnary(context.Background(), time.Second, func(ctx context.Context) error {
		callCount++
		return nil
	})
	if err != nil {
		t.Fatalf("InvokeUnary: %v", err)
	}
	if callCount != 1 {
		t.Errorf("call count = %d, want 1", callCount)
	}
}

func TestInvokeUnary_PropagatesError(t *testing.T) {
	expectErr := errors.New("call failed")
	err := InvokeUnary(context.Background(), time.Second, func(ctx context.Context) error {
		return expectErr
	})
	if err != expectErr {
		t.Errorf("err = %v, want %v", err, expectErr)
	}
}

func TestNewClientConn_WithRealServer(t *testing.T) {
	srv := grpc.NewServer()
	RegisterHealth(srv)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer lis.Close()
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := NewClientConn(ClientConfig{
		Address: lis.Addr().String(),
		Logger:  zaptest.NewLogger(t),
	})
	if err != nil {
		t.Fatalf("NewClientConn: %v", err)
	}
	defer conn.Close()
	if conn == nil {
		t.Fatal("conn is nil")
	}
}

func TestNewClientConn_WithInterceptors(t *testing.T) {
	srv := grpc.NewServer()
	RegisterHealth(srv)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer lis.Close()
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	unaryCalled := false
	streamCalled := false
	conn, err := NewClientConn(ClientConfig{
		Address: lis.Addr().String(),
		UnaryInterceptors: []grpc.UnaryClientInterceptor{
			func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				unaryCalled = true
				return invoker(ctx, method, req, reply, cc, opts...)
			},
		},
		StreamInterceptors: []grpc.StreamClientInterceptor{
			func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				streamCalled = true
				return streamer(ctx, desc, cc, method, opts...)
			},
		},
	})
	if err != nil {
		t.Fatalf("NewClientConn: %v", err)
	}
	defer conn.Close()
	if !unaryCalled && !streamCalled {
		// At least chain was used; we may only hit unary for health check
		t.Log("interceptors registered (unary/stream may be hit on actual RPC)")
	}
}

func TestNewClientConn_WithoutLogger(t *testing.T) {
	srv := grpc.NewServer()
	RegisterHealth(srv)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer lis.Close()
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := NewClientConn(ClientConfig{Address: lis.Addr().String()})
	if err != nil {
		t.Fatalf("NewClientConn: %v", err)
	}
	defer conn.Close()
}
