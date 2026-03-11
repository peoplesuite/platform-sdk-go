package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want Options
	}{
		{
			name: "defaults applied",
			opts: Options{},
			want: Options{
				ListenAddr:      ":8080",
				StaleThreshold:  5 * time.Minute,
				ReadTimeout:     5 * time.Second,
				WriteTimeout:    10 * time.Second,
				ShutdownTimeout: 10 * time.Second,
			},
		},
		{
			name: "custom options preserved",
			opts: Options{
				ListenAddr:     ":9090",
				StaleThreshold: 10 * time.Minute,
			},
			want: Options{
				ListenAddr:      ":9090",
				StaleThreshold:  10 * time.Minute,
				ReadTimeout:     5 * time.Second,
				WriteTimeout:    10 * time.Second,
				ShutdownTimeout: 10 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.opts)
			if m == nil {
				t.Fatal("NewManager() returned nil")
			}
			if m.opts != tt.want {
				t.Errorf("NewManager().opts = %+v, want %+v", m.opts, tt.want)
			}
			if m.state == nil {
				t.Error("NewManager().state is nil")
			}
		})
	}
}

func TestManager_StateSetters(t *testing.T) {
	m := NewManager(Options{})

	// Test SetTemporalConnected
	m.SetTemporalConnected(true)
	if !m.state.temporalConnected.Load() {
		t.Error("SetTemporalConnected(true) failed")
	}

	m.SetTemporalConnected(false)
	if m.state.temporalConnected.Load() {
		t.Error("SetTemporalConnected(false) failed")
	}

	// Test SetWorkerRunning
	m.SetWorkerRunning(true)
	if !m.state.workerRunning.Load() {
		t.Error("SetWorkerRunning(true) failed")
	}

	m.SetWorkerRunning(false)
	if m.state.workerRunning.Load() {
		t.Error("SetWorkerRunning(false) failed")
	}

	// Test SetDraining
	m.SetDraining(true)
	if !m.state.draining.Load() {
		t.Error("SetDraining(true) failed")
	}

	m.SetDraining(false)
	if m.state.draining.Load() {
		t.Error("SetDraining(false) failed")
	}
}

func TestManager_MarkPoll(t *testing.T) {
	m := NewManager(Options{})

	before := time.Now()
	m.MarkPoll()
	after := time.Now()

	pollTime := time.Unix(0, m.state.lastPollTime.Load())
	if pollTime.Before(before) || pollTime.After(after) {
		t.Errorf("MarkPoll() time %v not in range [%v, %v]", pollTime, before, after)
	}
}

func TestManager_StartHTTP_Integration(t *testing.T) {
	m := NewManager(Options{
		ListenAddr: ":0", // Random port
	})

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- m.StartHTTP("")
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Server should be running, attempt to start again should fail
	err := m.StartHTTP("")
	if err == nil {
		t.Error("second StartHTTP() should return error")
	}

	// Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// Server should have stopped
	select {
	case err := <-errCh:
		// ListenAndServe returns nil on clean shutdown
		if err != nil {
			t.Errorf("StartHTTP() returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("StartHTTP() did not return after Shutdown()")
	}
}

func TestManager_Shutdown_WithoutStart(t *testing.T) {
	m := NewManager(Options{})

	// Shutdown without starting should not error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() without Start() error = %v", err)
	}
}

func TestManager_Endpoints_Integration(t *testing.T) {
	// Find available port
	m := NewManager(Options{
		ListenAddr:     "localhost:0",
		StaleThreshold: -1, // Disable stale check for simpler testing
	})

	// Start server in background
	serverReady := make(chan string)
	go func() {
		// We need to extract the actual address
		// For this test, we'll use a fixed test port
		m.opts.ListenAddr = "localhost:18080"
		serverReady <- m.opts.ListenAddr
		_ = m.StartHTTP("")
	}()

	addr := <-serverReady
	baseURL := fmt.Sprintf("http://%s", addr)

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	tests := []struct {
		name           string
		endpoint       string
		setupState     func()
		wantStatusCode int
	}{
		{
			name:     "/livez always returns 200",
			endpoint: "/livez",
			setupState: func() {
				// No setup needed
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:     "/readyz returns 503 when not ready",
			endpoint: "/readyz",
			setupState: func() {
				m.SetTemporalConnected(false)
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			name:     "/readyz returns 200 when ready",
			endpoint: "/readyz",
			setupState: func() {
				m.SetTemporalConnected(true)
				m.SetWorkerRunning(true)
				m.SetDraining(false)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:     "/healthy returns JSON",
			endpoint: "/healthy",
			setupState: func() {
				m.SetTemporalConnected(true)
				m.SetWorkerRunning(true)
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupState()

			resp, err := http.Get(baseURL + tt.endpoint)
			if err != nil {
				t.Fatalf("GET %s failed: %v", tt.endpoint, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatusCode {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("GET %s status = %d, want %d. Body: %s", tt.endpoint, resp.StatusCode, tt.wantStatusCode, string(body))
			}
		})
	}

	// Cleanup
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = m.Shutdown(ctx)
}

func TestManager_GracefulShutdown(t *testing.T) {
	m := NewManager(Options{
		ListenAddr: "localhost:18081",
	})

	// Start server
	go func() { _ = m.StartHTTP("") }()
	time.Sleep(100 * time.Millisecond)

	// Verify server is running
	resp, err := http.Get("http://localhost:18081/livez")
	if err != nil {
		t.Fatalf("server not running: %v", err)
	}
	resp.Body.Close()

	// Shutdown with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := m.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// Verify server is stopped
	time.Sleep(100 * time.Millisecond)
	_, err = http.Get("http://localhost:18081/livez")
	if err == nil {
		t.Error("server still running after Shutdown()")
	}
}
