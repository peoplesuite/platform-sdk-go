package health

import "time"

// Options configures the health check manager.
type Options struct {
	// ListenAddr is the address to listen on (default ":8080")
	ListenAddr string

	// StaleThreshold is the maximum duration since last poll before considering worker stale.
	// 0 = apply default (5 minutes)
	// Negative value (e.g. -1) = disable stale checking
	StaleThreshold time.Duration

	// ReadTimeout is the maximum duration for reading the entire request (default: 5s)
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response (default: 10s)
	WriteTimeout time.Duration

	// ShutdownTimeout is the maximum duration to wait for graceful shutdown (default: 10s)
	ShutdownTimeout time.Duration
}

// applyDefaults returns a new Options with defaults applied for zero values.
func (o Options) applyDefaults() Options {
	if o.ListenAddr == "" {
		o.ListenAddr = ":8080"
	}
	if o.StaleThreshold == 0 {
		o.StaleThreshold = 5 * time.Minute
	} else if o.StaleThreshold < 0 {
		// Negative value means disabled - set to 0
		o.StaleThreshold = 0
	}
	if o.ReadTimeout == 0 {
		o.ReadTimeout = 5 * time.Second
	}
	if o.WriteTimeout == 0 {
		o.WriteTimeout = 10 * time.Second
	}
	if o.ShutdownTimeout == 0 {
		o.ShutdownTimeout = 10 * time.Second
	}
	return o
}
