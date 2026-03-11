package metrics

import (
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

type Metrics struct {
	client *statsd.Client
}

// Config defines metrics configuration.
type Config struct {
	ServiceName string
	Address     string
}

func New(cfg Config) (*Metrics, error) {

	client, err := statsd.New(
		cfg.Address,
		statsd.WithNamespace(cfg.ServiceName+"."),
	)

	if err != nil {
		return nil, err
	}

	return &Metrics{
		client: client,
	}, nil
}

func (m *Metrics) Increment(name string, tags []string) {
	_ = m.client.Incr(name, tags, 1)
}

func (m *Metrics) Gauge(name string, value float64, tags []string) {
	_ = m.client.Gauge(name, value, tags, 1)
}

func (m *Metrics) Histogram(name string, value float64, tags []string) {
	_ = m.client.Histogram(name, value, tags, 1)
}

func (m *Metrics) Timing(name string, duration time.Duration, tags []string) {
	_ = m.client.Timing(name, duration, tags, 1)
}

func (m *Metrics) Close() error {
	return m.client.Close()
}
