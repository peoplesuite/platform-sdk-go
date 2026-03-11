package logging

import (
	"go.uber.org/zap"
)

// Method returns a zap field for HTTP method.
func Method(v string) zap.Field {
	return zap.String("http.method", v)
}

// Path returns a zap field for HTTP path.
func Path(v string) zap.Field {
	return zap.String("http.path", v)
}

// Status returns a zap field for HTTP status code.
func Status(v int) zap.Field {
	return zap.Int("http.status", v)
}

// Duration returns a zap field for duration.
func Duration(v string) zap.Field {
	return zap.String("duration", v)
}

// RequestID returns a zap field for request ID.
func RequestID(v string) zap.Field {
	return zap.String("request_id", v)
}

// UserID returns a zap field for user ID.
func UserID(v string) zap.Field {
	return zap.String("user_id", v)
}
