package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds Prometheus metrics for HTTP and gRPC server observability.
// Used by pkg/http middleware and pkg/grpc interceptors.
type Metrics struct {
	ActiveRequests  *prometheus.GaugeVec
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	ErrorsTotal     *prometheus.CounterVec
}

// NewServerMetrics creates Metrics with standard labels: "route" (method+path or gRPC method), "code" (HTTP status or gRPC code).
func NewServerMetrics(namespace string) *Metrics {
	return &Metrics{
		ActiveRequests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Namespace: namespace, Name: "active_requests"},
			[]string{"route"},
		),
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Namespace: namespace, Name: "requests_total"},
			[]string{"route", "code"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{Namespace: namespace, Name: "request_duration_seconds"},
			[]string{"route"},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Namespace: namespace, Name: "errors_total"},
			[]string{"route", "code"},
		),
	}
}

// MustRegister registers all metrics with the default registry. Panics on error.
func (m *Metrics) MustRegister(reg prometheus.Registerer) {
	reg.MustRegister(m.ActiveRequests, m.RequestsTotal, m.RequestDuration, m.ErrorsTotal)
}
