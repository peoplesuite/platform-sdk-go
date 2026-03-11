package health

import (
	"sync"
	"testing"
	"time"
)

func TestState_SettersAndGetters(t *testing.T) {
	s := newState()

	// Initial state should be all false
	if s.IsReady() {
		t.Error("expected IsReady() to be false initially")
	}

	// Set individual states
	s.SetTemporalConnected(true)
	if s.temporalConnected.Load() != true {
		t.Error("SetTemporalConnected(true) failed")
	}

	s.SetWorkerRunning(true)
	if s.workerRunning.Load() != true {
		t.Error("SetWorkerRunning(true) failed")
	}

	s.SetDraining(false)
	if s.draining.Load() != false {
		t.Error("SetDraining(false) failed")
	}

	// Now should be ready (all conditions met)
	if !s.IsReady() {
		t.Error("expected IsReady() to be true")
	}

	// Set draining to true
	s.SetDraining(true)
	if s.IsReady() {
		t.Error("expected IsReady() to be false when draining")
	}
}

func TestState_IsReady(t *testing.T) {
	tests := []struct {
		name              string
		temporalConnected bool
		workerRunning     bool
		draining          bool
		want              bool
	}{
		{
			name:              "all conditions met",
			temporalConnected: true,
			workerRunning:     true,
			draining:          false,
			want:              true,
		},
		{
			name:              "temporal not connected",
			temporalConnected: false,
			workerRunning:     true,
			draining:          false,
			want:              false,
		},
		{
			name:              "worker not running",
			temporalConnected: true,
			workerRunning:     false,
			draining:          false,
			want:              false,
		},
		{
			name:              "draining",
			temporalConnected: true,
			workerRunning:     true,
			draining:          true,
			want:              false,
		},
		{
			name:              "all false",
			temporalConnected: false,
			workerRunning:     false,
			draining:          false,
			want:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newState()
			s.SetTemporalConnected(tt.temporalConnected)
			s.SetWorkerRunning(tt.workerRunning)
			s.SetDraining(tt.draining)

			if got := s.IsReady(); got != tt.want {
				t.Errorf("IsReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestState_MarkPoll(t *testing.T) {
	s := newState()

	// Initially, last poll time should be 0
	if s.lastPollTime.Load() != 0 {
		t.Error("expected lastPollTime to be 0 initially")
	}

	// Mark poll
	before := time.Now()
	s.MarkPoll()
	after := time.Now()

	pollTime := time.Unix(0, s.lastPollTime.Load())
	if pollTime.Before(before) || pollTime.After(after) {
		t.Errorf("MarkPoll() time %v not in range [%v, %v]", pollTime, before, after)
	}

	// Mark poll again
	time.Sleep(10 * time.Millisecond)
	s.MarkPoll()
	newPollTime := time.Unix(0, s.lastPollTime.Load())
	if !newPollTime.After(pollTime) {
		t.Error("second MarkPoll() should be after first")
	}
}

func TestState_IsReadyWithStaleCheck(t *testing.T) {
	tests := []struct {
		name           string
		setupState     func(*State)
		staleThreshold time.Duration
		want           bool
	}{
		{
			name: "ready with fresh poll",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
				s.MarkPoll()
			},
			staleThreshold: 5 * time.Minute,
			want:           true,
		},
		{
			name: "ready but never polled",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
			},
			staleThreshold: 5 * time.Minute,
			want:           false,
		},
		{
			name: "ready with stale check disabled",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
				// Never poll
			},
			staleThreshold: -1, // Disabled (negative value)
			want:           true,
		},
		{
			name: "not ready due to draining",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(true)
				s.MarkPoll()
			},
			staleThreshold: 5 * time.Minute,
			want:           false,
		},
		{
			name: "ready with old poll but high threshold",
			setupState: func(s *State) {
				s.SetTemporalConnected(true)
				s.SetWorkerRunning(true)
				s.SetDraining(false)
				s.MarkPoll()
				time.Sleep(50 * time.Millisecond)
			},
			staleThreshold: 1 * time.Hour,
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newState()
			tt.setupState(s)

			if got := s.IsReadyWithStaleCheck(tt.staleThreshold); got != tt.want {
				t.Errorf("IsReadyWithStaleCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestState_Status(t *testing.T) {
	tests := []struct {
		name              string
		temporalConnected bool
		workerRunning     bool
		draining          bool
		markPoll          bool
		staleThreshold    time.Duration
		wantReady         bool
		wantReasons       []string
	}{
		{
			name:              "all healthy",
			temporalConnected: true,
			workerRunning:     true,
			draining:          false,
			markPoll:          true,
			staleThreshold:    5 * time.Minute,
			wantReady:         true,
			wantReasons:       []string{},
		},
		{
			name:              "temporal not connected",
			temporalConnected: false,
			workerRunning:     true,
			draining:          false,
			markPoll:          true,
			staleThreshold:    5 * time.Minute,
			wantReady:         false,
			wantReasons:       []string{"temporal not connected"},
		},
		{
			name:              "worker not running",
			temporalConnected: true,
			workerRunning:     false,
			draining:          false,
			markPoll:          true,
			staleThreshold:    5 * time.Minute,
			wantReady:         false,
			wantReasons:       []string{"worker not running"},
		},
		{
			name:              "draining",
			temporalConnected: true,
			workerRunning:     true,
			draining:          true,
			markPoll:          true,
			staleThreshold:    5 * time.Minute,
			wantReady:         false,
			wantReasons:       []string{"draining"},
		},
		{
			name:              "never polled",
			temporalConnected: true,
			workerRunning:     true,
			draining:          false,
			markPoll:          false,
			staleThreshold:    5 * time.Minute,
			wantReady:         false,
			wantReasons:       []string{"never polled"},
		},
		{
			name:              "multiple issues",
			temporalConnected: false,
			workerRunning:     false,
			draining:          true,
			markPoll:          false,
			staleThreshold:    5 * time.Minute,
			wantReady:         false,
			wantReasons:       []string{"temporal not connected", "worker not running", "draining", "never polled"},
		},
		{
			name:              "healthy with stale check disabled",
			temporalConnected: true,
			workerRunning:     true,
			draining:          false,
			markPoll:          false,
			staleThreshold:    -1, // Disabled (negative value)
			wantReady:         true,
			wantReasons:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newState()
			s.SetTemporalConnected(tt.temporalConnected)
			s.SetWorkerRunning(tt.workerRunning)
			s.SetDraining(tt.draining)
			if tt.markPoll {
				s.MarkPoll()
			}

			status := s.Status(tt.staleThreshold)

			if status.Ready != tt.wantReady {
				t.Errorf("Status().Ready = %v, want %v", status.Ready, tt.wantReady)
			}

			if status.TemporalConnected != tt.temporalConnected {
				t.Errorf("Status().TemporalConnected = %v, want %v", status.TemporalConnected, tt.temporalConnected)
			}

			if status.WorkerRunning != tt.workerRunning {
				t.Errorf("Status().WorkerRunning = %v, want %v", status.WorkerRunning, tt.workerRunning)
			}

			if status.Draining != tt.draining {
				t.Errorf("Status().Draining = %v, want %v", status.Draining, tt.draining)
			}

			if len(status.Reasons) != len(tt.wantReasons) {
				t.Errorf("Status().Reasons length = %d, want %d. Got: %v", len(status.Reasons), len(tt.wantReasons), status.Reasons)
			} else {
				for i, reason := range tt.wantReasons {
					if status.Reasons[i] != reason {
						t.Errorf("Status().Reasons[%d] = %v, want %v", i, status.Reasons[i], reason)
					}
				}
			}

			if tt.markPoll {
				if status.LastPollAgoSeconds < 0 {
					t.Error("Status().LastPollAgoSeconds should be >= 0 when poll was marked")
				}
			} else {
				if status.LastPollAgoSeconds != -1 {
					t.Errorf("Status().LastPollAgoSeconds = %v, want -1 when never polled", status.LastPollAgoSeconds)
				}
			}
		})
	}
}

func TestState_ThreadSafety(t *testing.T) {
	s := newState()
	var wg sync.WaitGroup

	// Run concurrent operations
	for i := 0; i < 100; i++ {
		wg.Add(4)

		go func() {
			defer wg.Done()
			s.SetTemporalConnected(true)
			s.SetTemporalConnected(false)
		}()

		go func() {
			defer wg.Done()
			s.SetWorkerRunning(true)
			s.SetWorkerRunning(false)
		}()

		go func() {
			defer wg.Done()
			s.SetDraining(true)
			s.SetDraining(false)
		}()

		go func() {
			defer wg.Done()
			s.MarkPoll()
			s.IsReady()
			s.Status(5 * time.Minute)
		}()
	}

	wg.Wait()
	// If we get here without data races, the test passes
}
