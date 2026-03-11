package runtime

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"

	grpcpkg "peoplesuite/platform-sdk-go/pkg/grpc"
)

func TestNew_MinimalOptions(t *testing.T) {
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if rt == nil {
		t.Fatal("New returned nil Runtime")
	}
}

func TestRun_ContextCancelled_ReturnsNil(t *testing.T) {
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = rt.Run(ctx)
	if err != nil {
		t.Errorf("Run() = %v, want nil", err)
	}
}

func TestRun_StartHooksError_ReturnsError(t *testing.T) {
	wantErr := errors.New("start hook failed")
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
		StartHooks:  []func(context.Context) error{func(context.Context) error { return wantErr }},
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got := rt.Run(context.Background())
	if got != wantErr {
		t.Errorf("Run() = %v, want %v", got, wantErr)
	}
}

func TestRun_WorkerRunsAndRunExitsOnContextCancel(t *testing.T) {
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
		Workers: []Worker{
			func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
		},
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var runErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runErr = rt.Run(ctx)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()
	if runErr != nil {
		t.Errorf("Run() = %v, want nil", runErr)
	}
}

func TestRun_WorkerReturnsError_NoPanic(t *testing.T) {
	workerRan := make(chan struct{})
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
		Workers: []Worker{
			func(ctx context.Context) error {
				close(workerRan)
				return errors.New("worker err")
			},
		},
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		_ = rt.Run(ctx)
		close(done)
	}()
	select {
	case <-workerRan:
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not run")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not exit")
	}
}

func TestRun_StopHooksCalledOnShutdown(t *testing.T) {
	var stopCalled bool
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
		StopHooks: []func(context.Context) error{
			func(context.Context) error {
				stopCalled = true
				return nil
			},
		},
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = rt.Run(ctx)
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done
	if !stopCalled {
		t.Error("StopHooks were not called")
	}
}

func TestRun_StopHookError_ReturnedFromRun(t *testing.T) {
	wantErr := errors.New("stop hook failed")
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
		StopHooks: []func(context.Context) error{
			func(context.Context) error {
				return wantErr
			},
		},
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- rt.Run(ctx)
	}()
	time.Sleep(50 * time.Millisecond)
	cancel()
	got := <-done
	if got != wantErr {
		t.Errorf("Run() = %v, want %v", got, wantErr)
	}
}

func TestRun_GRPCListenFails_ReturnsError(t *testing.T) {
	grpcSrv := grpc.NewServer()
	grpcpkg.RegisterHealth(grpcSrv)
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
		GRPCPort:    999999, // invalid port (out of range) so Listen fails
		GRPCServer:  grpcSrv,
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	got := rt.Run(ctx)
	if got == nil {
		t.Fatal("Run() expected error when gRPC listen fails, got nil")
	}
}

func TestRun_HTTPListenFails_ReturnsError(t *testing.T) {
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
		HTTPPort:    999999, // invalid port so Listen fails
		HTTPHandler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	got := rt.Run(ctx)
	if got == nil {
		t.Fatal("Run() expected error when HTTP listen fails, got nil")
	}
}

func TestRuntime_initHealth_ReturnsManager(t *testing.T) {
	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	mgr := rt.initHealth(ctx)
	if mgr == nil {
		t.Fatal("initHealth returned nil manager")
	}
}

func TestRun_WithHTTPAndGRPC_GracefulShutdown(t *testing.T) {
	// Use distinct ports to avoid conflicts
	httpPort := 18580
	grpcPort := 18581

	httpHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	grpcSrv := grpc.NewServer()
	grpcpkg.RegisterHealth(grpcSrv)

	opts := Options{
		ServiceName: "test",
		Version:     "0",
		Environment: "test",
		HTTPPort:    httpPort,
		GRPCPort:    grpcPort,
		HTTPHandler: httpHandler,
		GRPCServer:  grpcSrv,
	}
	rt, err := New(opts)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = rt.Run(ctx)
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:18580/")
	if err != nil {
		t.Logf("HTTP GET (may be timing): %v", err)
	} else {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("GET / = %d, want 200", resp.StatusCode)
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not exit after cancel")
	}
}
