package httpclient

import "fmt"

// HTTPError represents a non-2xx response (StatusCode and Body). Use errors.As to detect and map to pkg/errors in handlers.
type HTTPError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http %d: %s", e.StatusCode, string(e.Body))
}

func NewHTTPError(statusCode int, body []byte) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Body:       body,
	}
}
