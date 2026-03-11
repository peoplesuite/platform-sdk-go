package http

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"peoplesuite/platform-sdk-go/pkg/observability"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in order: first middleware is outermost.
//
//	Chain(RequestID, Logging, Recovery)(handler)
//	→ RequestID wraps Logging wraps Recovery wraps handler
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

const RequestIDHeader = "X-Request-ID"

type reqIDKey struct{}

// RequestID ensures every request has a unique request ID.
// If the incoming request has an X-Request-ID header, it's reused; otherwise a new UUID is generated.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			id = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), reqIDKey{}, id)
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from context.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(reqIDKey{}).(string); ok {
		return v
	}
	return ""
}

// CORSConfig configures Cross-Origin Resource Sharing.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           int // seconds
}

// DefaultCORSConfig returns permissive CORS for development.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS returns a CORS middleware.
func CORS(cfg CORSConfig) Middleware {
	originSet := make(map[string]bool)
	allowAll := false
	for _, o := range cfg.AllowOrigins {
		if o == "*" {
			allowAll = true
		}
		originSet[o] = true
	}

	methods := joinStrings(cfg.AllowMethods, ", ")
	headers := joinStrings(cfg.AllowHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowAll || originSet[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Preflight
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", methods)
				w.Header().Set("Access-Control-Allow-Headers", headers)
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	n, err := sr.ResponseWriter.Write(b)
	sr.bytes += n
	return n, err
}

// Logging logs every request with method, path, status, duration.
func Logging(logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rec.status),
				zap.Int("bytes", rec.bytes),
				zap.Duration("duration", duration),
				zap.String("request_id", GetRequestID(r.Context())),
			}

			if rec.status >= 500 {
				logger.Error("HTTP request", fields...)
			} else if rec.status >= 400 {
				logger.Warn("HTTP request", fields...)
			} else {
				logger.Info("HTTP request", fields...)
			}
		})
	}
}

// Recovery catches panics and returns 500.
func Recovery(logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := string(debug.Stack())
					logger.Error("HTTP panic recovered",
						zap.Any("panic", rec),
						zap.String("path", r.URL.Path),
						zap.String("stack", stack),
					)
					ErrorJSON(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Tracing creates an OTel span for each request.
func Tracing() Middleware {
	tracer := otel.Tracer("pkg/httputil")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.path", r.URL.Path),
				),
			)
			defer span.End()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Metrics records request count, duration, and active requests.
func Metrics(m *observability.Metrics) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			label := r.Method + " " + r.URL.Path

			m.ActiveRequests.WithLabelValues(label).Inc()
			defer m.ActiveRequests.WithLabelValues(label).Dec()

			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			duration := time.Since(start)
			m.RequestsTotal.WithLabelValues(label, fmt.Sprintf("%d", rec.status)).Inc()
			m.RequestDuration.WithLabelValues(label).Observe(duration.Seconds())
		})
	}
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
