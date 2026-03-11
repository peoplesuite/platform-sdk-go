package metrics

import (
	"net/http"
	"time"
)

func HTTPMiddleware(m *Metrics) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			next.ServeHTTP(w, r)

			duration := time.Since(start)

			tags := []string{
				"method:" + r.Method,
				"path:" + r.URL.Path,
			}

			m.Increment("http.requests", tags)
			m.Timing("http.duration", duration, tags)

		})
	}
}
