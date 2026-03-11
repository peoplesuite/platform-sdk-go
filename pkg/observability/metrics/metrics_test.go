package metrics

import (
	"testing"
	"time"
)

func TestNew_InvalidAddress(t *testing.T) {
	_, err := New(Config{
		ServiceName: "test",
		Address:     "invalid-address-no-colon",
	})
	if err == nil {
		t.Error("New with invalid address should return error")
	}
}

func TestNew_EmptyAddress(t *testing.T) {
	// Empty address may fail or use default; behavior is implementation-dependent
	_, err := New(Config{ServiceName: "test", Address: ""})
	// Just ensure we don't panic; accept either nil or non-nil error
	_ = err
}

func TestMetrics_MethodsNoPanic(t *testing.T) {
	// Create with a valid-looking but unreachable address so we can test method calls
	// statsd.New("127.0.0.1:1") might succeed (client created) but send fails
	m, err := New(Config{ServiceName: "test", Address: "127.0.0.1:19999"})
	if err != nil {
		t.Skipf("New failed (no statsd listener): %v", err)
	}
	defer m.Close()

	// Methods should not panic; they ignore send errors
	m.Increment("n", nil)
	m.Gauge("g", 1, nil)
	m.Histogram("h", 1, nil)
	m.Timing("t", time.Millisecond, nil)
}
