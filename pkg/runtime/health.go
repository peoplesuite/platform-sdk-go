package runtime

import (
	"context"

	"github.com/peoplesuite/platform-sdk-go/pkg/health"
)

// initHealth is reserved for optional health HTTP server (e.g. /livez, /readyz).
// It is not currently called by Runtime; kept for API compatibility or future use.
//
//nolint:unused
func (r *Runtime) initHealth(ctx context.Context) *health.Manager {
	manager := health.NewManager(health.Options{})
	go func() { _ = manager.StartHTTP(":8081") }()
	return manager
}
