package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

// Manager manages health check state and HTTP server lifecycle.
type Manager struct {
	opts   Options
	state  *State
	server *http.Server
	mu     sync.Mutex
}

// NewManager creates a new health check manager with the given options.
// Defaults are applied for any zero-value options.
func NewManager(opts Options) *Manager {
	opts = opts.applyDefaults()
	return &Manager{
		opts:  opts,
		state: newState(),
	}
}

// SetTemporalConnected sets whether the Temporal client is connected.
func (m *Manager) SetTemporalConnected(connected bool) {
	m.state.SetTemporalConnected(connected)
}

// SetWorkerRunning sets whether the Temporal worker is running.
func (m *Manager) SetWorkerRunning(running bool) {
	m.state.SetWorkerRunning(running)
}

// SetDraining sets whether the service is draining (shutting down).
// When draining is true, readiness probes will fail.
func (m *Manager) SetDraining(draining bool) {
	m.state.SetDraining(draining)
}

// MarkPoll records the current time as the last successful poll/task execution.
// This is used to detect stale workers that haven't processed tasks recently.
func (m *Manager) MarkPoll() {
	m.state.MarkPoll()
}

// StartHTTP starts the HTTP health check server on the configured address.
// This is a blocking call that returns when the server stops or encounters an error.
// Use it in a goroutine for background operation.
//
// The addr parameter overrides the ListenAddr from Options if provided.
// Pass empty string to use the configured ListenAddr.
func (m *Manager) StartHTTP(addr string) error {
	m.mu.Lock()
	if m.server != nil {
		m.mu.Unlock()
		return fmt.Errorf("HTTP server already started")
	}

	if addr == "" {
		addr = m.opts.ListenAddr
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/livez", m.handleLivez)
	mux.HandleFunc("/readyz", m.handleReadyz)
	mux.HandleFunc("/healthy", m.handleHealthy)

	m.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  m.opts.ReadTimeout,
		WriteTimeout: m.opts.WriteTimeout,
	}
	m.mu.Unlock()

	// ListenAndServe always returns a non-nil error
	if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("health server failed: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the HTTP health check server.
// It respects the context timeout/cancellation.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	srv := m.server
	m.mu.Unlock()

	if srv == nil {
		return nil // Already shutdown or never started
	}

	return srv.Shutdown(ctx)
}
