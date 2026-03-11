package health

import (
	"encoding/json"
	"net/http"
	"strings"
)

// handleLivez handles the /livez endpoint.
// Always returns 200 OK if the process is running.
// This endpoint performs no checks and should be used for Kubernetes livenessProbe.
func (m *Manager) handleLivez(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// handleReadyz handles the /readyz endpoint.
// Returns 200 OK only if all readiness conditions are met.
// Returns 503 Service Unavailable with reasons if not ready.
// This endpoint should be used for Kubernetes readinessProbe.
func (m *Manager) handleReadyz(w http.ResponseWriter, r *http.Request) {
	status := m.state.Status(m.opts.StaleThreshold)

	if status.Ready {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
		return
	}

	// Not ready - return 503 with reasons
	w.WriteHeader(http.StatusServiceUnavailable)
	reasons := strings.Join(status.Reasons, ", ")
	_, _ = w.Write([]byte("not ready: " + reasons))
}

// handleHealthy handles the /healthy endpoint.
// Always returns JSON with full health status including readiness state.
// This endpoint is for debugging and operational visibility, not for probes.
func (m *Manager) handleHealthy(w http.ResponseWriter, r *http.Request) {
	status := m.state.Status(m.opts.StaleThreshold)

	w.Header().Set("Content-Type", "application/json")

	// Set HTTP status based on readiness
	if status.Ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Always return JSON body regardless of status
	if err := json.NewEncoder(w).Encode(status); err != nil {
		// If JSON encoding fails, we've already written the header
		// Nothing we can do at this point
		return
	}
}
