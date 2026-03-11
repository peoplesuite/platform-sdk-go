package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestManager_handleLivez(t *testing.T) {
	m := NewManager(Options{})

	tests := []struct {
		name           string
		setupState     func(*State)
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "always returns 200",
			setupState:     func(s *State) {},
			wantStatusCode: http.StatusOK,
			wantBody:       "ok",
		},
		{
			name: "returns 200 even when not ready",
			setupState: func(s *State) {
				s.SetTemporalConnected(false)
				s.SetWorkerRunning(false)
				s.SetDraining(true)
			},
			wantStatusCode: http.StatusOK,
			wantBody:       "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupState(m.state)

			req := httptest.NewRequest(http.MethodGet, "/livez", nil)
			w := httptest.NewRecorder()

			m.handleLivez(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("handleLivez() status = %d, want %d", w.Code, tt.wantStatusCode)
			}

			if got := w.Body.String(); got != tt.wantBody {
				t.Errorf("handleLivez() body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestManager_handleReadyz(t *testing.T) {
	tests := []struct {
		name           string
		setupState     func(*State)
		staleThreshold time.Duration
		wantStatusCode int
		wantBodyPrefix string
	}{
		{
			name: "ready",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
			},
			staleThreshold: -1, // Disable stale check
			wantStatusCode: http.StatusOK,
			wantBodyPrefix: "ready",
		},
		{
			name: "not ready - temporal not connected",
			setupState: func(s *State) {
				s.SetTemporalConnected(false)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
			},
			staleThreshold: -1, // Disable stale check
			wantStatusCode: http.StatusServiceUnavailable,
			wantBodyPrefix: "not ready: temporal not connected",
		},
		{
			name: "not ready - worker not running",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(false)
				s.SetDraining(false)
			},
			staleThreshold: -1, // Disable stale check
			wantStatusCode: http.StatusServiceUnavailable,
			wantBodyPrefix: "not ready: worker not running",
		},
		{
			name: "not ready - draining",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(true)
			},
			staleThreshold: -1, // Disable stale check
			wantStatusCode: http.StatusServiceUnavailable,
			wantBodyPrefix: "not ready: draining",
		},
		{
			name: "not ready - multiple reasons",
			setupState: func(s *State) {
				s.SetTemporalConnected(false)
				s.SetWorkerRunning(false)
				s.SetDraining(true)
			},
			staleThreshold: -1, // Disable stale check
			wantStatusCode: http.StatusServiceUnavailable,
			wantBodyPrefix: "not ready:",
		},
		{
			name: "ready with recent poll",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
				s.MarkPoll()
			},
			staleThreshold: 5 * time.Minute,
			wantStatusCode: http.StatusOK,
			wantBodyPrefix: "ready",
		},
		{
			name: "not ready - never polled",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
			},
			staleThreshold: 5 * time.Minute,
			wantStatusCode: http.StatusServiceUnavailable,
			wantBodyPrefix: "not ready: never polled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(Options{StaleThreshold: tt.staleThreshold})
			tt.setupState(m.state)

			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
			w := httptest.NewRecorder()

			m.handleReadyz(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("handleReadyz() status = %d, want %d", w.Code, tt.wantStatusCode)
			}

			body := w.Body.String()
			if len(body) < len(tt.wantBodyPrefix) || body[:len(tt.wantBodyPrefix)] != tt.wantBodyPrefix {
				t.Errorf("handleReadyz() body = %q, want prefix %q", body, tt.wantBodyPrefix)
			}
		})
	}
}

func TestManager_handleHealthy(t *testing.T) {
	tests := []struct {
		name           string
		setupState     func(*State)
		staleThreshold time.Duration
		wantStatusCode int
		wantReady      bool
	}{
		{
			name: "ready",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
			},
			staleThreshold: -1, // Disable stale check
			wantStatusCode: http.StatusOK,
			wantReady:      true,
		},
		{
			name: "not ready",
			setupState: func(s *State) {
				s.SetTemporalConnected(false)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
			},
			staleThreshold: -1, // Disable stale check
			wantStatusCode: http.StatusServiceUnavailable,
			wantReady:      false,
		},
		{
			name: "draining",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(true)
			},
			staleThreshold: -1, // Disable stale check
			wantStatusCode: http.StatusServiceUnavailable,
			wantReady:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(Options{StaleThreshold: tt.staleThreshold})
			tt.setupState(m.state)

			req := httptest.NewRequest(http.MethodGet, "/healthy", nil)
			w := httptest.NewRecorder()

			m.handleHealthy(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("handleHealthy() status = %d, want %d", w.Code, tt.wantStatusCode)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("handleHealthy() Content-Type = %q, want %q", contentType, "application/json")
			}

			var status StatusResponse
			if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if status.Ready != tt.wantReady {
				t.Errorf("handleHealthy() Ready = %v, want %v", status.Ready, tt.wantReady)
			}

			// Verify response includes all expected fields
			if status.Reasons == nil {
				t.Error("handleHealthy() Reasons should not be nil")
			}
		})
	}
}

func TestManager_handleHealthy_JSONFields(t *testing.T) {
	m := NewManager(Options{StaleThreshold: 5 * time.Minute})
	m.state.SetTemporalConnected(true)
	m.state.SetWorkerRunning(false)
	m.state.SetDraining(true)
	m.state.MarkPoll()

	req := httptest.NewRequest(http.MethodGet, "/healthy", nil)
	w := httptest.NewRecorder()

	m.handleHealthy(w, req)

	var status StatusResponse
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if status.TemporalConnected != true {
		t.Error("expected TemporalConnected to be true")
	}
	if status.WorkerRunning != false {
		t.Error("expected WorkerRunning to be false")
	}
	if status.Draining != true {
		t.Error("expected Draining to be true")
	}
	if status.LastPollAgoSeconds < 0 {
		t.Error("expected LastPollAgoSeconds to be >= 0")
	}
	if status.Ready != false {
		t.Error("expected Ready to be false")
	}
	if len(status.Reasons) == 0 {
		t.Error("expected Reasons to have elements")
	}
}
