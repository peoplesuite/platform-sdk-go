package health

import (
	"sync/atomic"
	"time"
)

// State holds the atomic health state of the service.
type State struct {
	temporalConnected atomic.Bool
	workerRunning     atomic.Bool
	draining          atomic.Bool
	lastPollTime      atomic.Int64 // Unix nanoseconds
}

// newState creates a new State with zero values.
func newState() *State {
	return &State{}
}

// SetTemporalConnected sets whether the Temporal client is connected.
func (s *State) SetTemporalConnected(connected bool) {
	s.temporalConnected.Store(connected)
}

// SetWorkerRunning sets whether the Temporal worker is running.
func (s *State) SetWorkerRunning(running bool) {
	s.workerRunning.Store(running)
}

// SetDraining sets whether the service is draining (shutting down).
func (s *State) SetDraining(draining bool) {
	s.draining.Store(draining)
}

// MarkPoll records the current time as the last successful poll/task execution.
func (s *State) MarkPoll() {
	s.lastPollTime.Store(time.Now().UnixNano())
}

// IsReady returns true if all readiness conditions are met.
// Does not check stale threshold - use IsReadyWithStaleCheck for that.
func (s *State) IsReady() bool {
	return s.temporalConnected.Load() &&
		s.workerRunning.Load() &&
		!s.draining.Load()
}

// IsReadyWithStaleCheck returns true if all readiness conditions are met,
// including checking that the last poll is within the stale threshold.
// If staleThreshold is 0 or negative, stale checking is disabled.
func (s *State) IsReadyWithStaleCheck(staleThreshold time.Duration) bool {
	if !s.IsReady() {
		return false
	}

	// If stale checking is disabled (0 or negative), we're ready
	if staleThreshold <= 0 {
		return true
	}

	lastPoll := s.lastPollTime.Load()
	// If never polled (0), not ready
	if lastPoll == 0 {
		return false
	}

	elapsed := time.Since(time.Unix(0, lastPoll))
	return elapsed <= staleThreshold
}

// StatusResponse represents the full health status for JSON responses.
type StatusResponse struct {
	Ready              bool     `json:"ready"`
	TemporalConnected  bool     `json:"temporalConnected"`
	WorkerRunning      bool     `json:"workerRunning"`
	Draining           bool     `json:"draining"`
	LastPollAgoSeconds float64  `json:"lastPollAgoSeconds"`
	Reasons            []string `json:"reasons"`
}

// Status returns the current health status with all details.
func (s *State) Status(staleThreshold time.Duration) StatusResponse {
	status := StatusResponse{
		TemporalConnected: s.temporalConnected.Load(),
		WorkerRunning:     s.workerRunning.Load(),
		Draining:          s.draining.Load(),
		Reasons:           []string{},
	}

	lastPoll := s.lastPollTime.Load()
	if lastPoll != 0 {
		elapsed := time.Since(time.Unix(0, lastPoll))
		status.LastPollAgoSeconds = elapsed.Seconds()
	} else {
		status.LastPollAgoSeconds = -1 // Never polled
	}

	// Determine readiness and collect reasons if not ready
	ready := true

	if !status.TemporalConnected {
		ready = false
		status.Reasons = append(status.Reasons, "temporal not connected")
	}
	if !status.WorkerRunning {
		ready = false
		status.Reasons = append(status.Reasons, "worker not running")
	}
	if status.Draining {
		ready = false
		status.Reasons = append(status.Reasons, "draining")
	}

	// Check stale threshold if enabled (positive value)
	if staleThreshold > 0 {
		if lastPoll == 0 {
			ready = false
			status.Reasons = append(status.Reasons, "never polled")
		} else {
			elapsed := time.Since(time.Unix(0, lastPoll))
			if elapsed > staleThreshold {
				ready = false
				status.Reasons = append(status.Reasons, "stale poll")
			}
		}
	}
	// If staleThreshold is 0 or negative, stale checking is disabled

	status.Ready = ready
	return status
}
