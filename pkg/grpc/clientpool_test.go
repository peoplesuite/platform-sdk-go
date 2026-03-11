package grpc

import (
	"net"
	"testing"

	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
)

func TestNewClientPool(t *testing.T) {
	pool := NewClientPool(ClientConfig{Logger: zaptest.NewLogger(t)})
	if pool == nil {
		t.Fatal("NewClientPool returned nil")
	}
}

func TestClientPool_GetSameAddressReturnsSameConn(t *testing.T) {
	srv := grpc.NewServer()
	RegisterHealth(srv)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer func() { _ = lis.Close() }()
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	addr := lis.Addr().String()
	pool := NewClientPool(ClientConfig{Address: addr, Logger: zaptest.NewLogger(t)})
	defer func() { _ = pool.Close() }()

	conn1, err := pool.Get(addr)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	conn2, err := pool.Get(addr)
	if err != nil {
		t.Fatalf("Get second: %v", err)
	}
	if conn1 != conn2 {
		t.Error("Get(same addr) should return same conn")
	}
}

func TestClientPool_Close(t *testing.T) {
	pool := NewClientPool(ClientConfig{Logger: zaptest.NewLogger(t)})
	if err := pool.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}
