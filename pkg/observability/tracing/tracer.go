package tracing

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Config configures the Datadog tracer.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
}

// Init initializes Datadog tracer.
func Init(cfg Config) func() {

	tracer.Start(
		tracer.WithService(cfg.ServiceName),
		tracer.WithServiceVersion(cfg.ServiceVersion),
		tracer.WithEnv(cfg.Environment),
		tracer.WithRuntimeMetrics(),
	)

	return func() {
		tracer.Stop()
	}
}
