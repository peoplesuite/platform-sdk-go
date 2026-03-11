package httpclient

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Options configures the HTTP client (base URL, timeout, TLS, retry, tracing).
type Options struct {
	BaseURL   string
	Timeout   time.Duration
	VerifyTLS bool

	// Retry: exponential backoff on 5xx and connection errors. Zero RetryMax disables retry.
	RetryMax     int           // max attempts including first (default: 3)
	RetryWait    time.Duration // initial backoff (default: 1s)
	RetryMaxWait time.Duration // max backoff (default: 30s)

	// Tracer: if set, each request gets an OTel span (http.method, http.url, http.status_code).
	Tracer trace.Tracer

	// Connection pool settings (optional, with defaults)
	MaxIdleConns        int           // default: 100
	MaxIdleConnsPerHost int           // default: 10
	IdleConnTimeout     time.Duration // default: 90s
}
