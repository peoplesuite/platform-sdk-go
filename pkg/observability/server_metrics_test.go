package observability

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewServerMetrics(t *testing.T) {
	m := NewServerMetrics("test")
	if m == nil {
		t.Fatal("NewServerMetrics returned nil")
	}
	if m.ActiveRequests == nil {
		t.Error("ActiveRequests is nil")
	}
	if m.RequestsTotal == nil {
		t.Error("RequestsTotal is nil")
	}
	if m.RequestDuration == nil {
		t.Error("RequestDuration is nil")
	}
	if m.ErrorsTotal == nil {
		t.Error("ErrorsTotal is nil")
	}

	// Exercise label vectors (route, code where applicable)
	m.ActiveRequests.WithLabelValues("/foo").Inc()
	m.RequestsTotal.WithLabelValues("/foo", "OK").Inc()
	m.RequestDuration.WithLabelValues("/foo").Observe(0.1)
	m.ErrorsTotal.WithLabelValues("/foo", "Internal").Inc()
}

func TestMetrics_MustRegister(t *testing.T) {
	m := NewServerMetrics("mustreg")
	reg := prometheus.NewRegistry()
	// Must not panic
	m.MustRegister(reg)

	// Add a sample so Gather returns something
	m.ActiveRequests.WithLabelValues("x").Set(1)
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if len(metrics) == 0 {
		t.Error("expected at least one metric family after MustRegister")
	}
}
