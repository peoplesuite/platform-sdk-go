package errors

// Internal returns a 500 Internal error.
func Internal(msg string) error {
	return mk(500, KindInternal, msg, nil)
}

// InternalWrap returns a 500 Internal error wrapping cause.
func InternalWrap(msg string, cause error) error {
	return mk(500, KindInternal, msg, cause)
}

// InvalidArgument returns a 400 InvalidArgument error.
func InvalidArgument(msg string) error {
	return mk(400, KindInvalidArgument, msg, nil)
}

// NotFound returns a 404 NotFound error.
func NotFound(msg string) error {
	return mk(404, KindNotFound, msg, nil)
}

// AlreadyExists returns a 409 AlreadyExists error.
func AlreadyExists(msg string) error {
	return mk(409, KindAlreadyExists, msg, nil)
}

// PermissionDenied returns a 403 PermissionDenied error.
func PermissionDenied(msg string) error {
	return mk(403, KindPermissionDenied, msg, nil)
}

// Unauthenticated returns a 401 Unauthenticated error.
func Unauthenticated(msg string) error {
	return mk(401, KindUnauthenticated, msg, nil)
}

// Conflict returns a 409 Conflict error.
func Conflict(msg string) error {
	return mk(409, KindConflict, msg, nil)
}

// PreconditionFailed returns a 412 PreconditionFailed error.
func PreconditionFailed(msg string) error {
	return mk(412, KindPreconditionFailed, msg, nil)
}

// Unavailable returns a 503 Unavailable error.
func Unavailable(msg string) error {
	return mk(503, KindUnavailable, msg, nil)
}

// Timeout returns a 504 Timeout error.
func Timeout(msg string) error {
	return mk(504, KindTimeout, msg, nil)
}

// RateLimited returns a 429 RateLimited error.
func RateLimited(msg string) error {
	return mk(429, KindRateLimited, msg, nil)
}

// Wrap wraps cause with the given kind and operation.
func Wrap(kind Kind, op string, cause error) error {
	if cause == nil {
		return nil
	}

	msg := cause.Error()

	if op != "" {
		msg = op + ": " + msg
	}

	return mk(kindToStatus(kind), kind, msg, cause)
}
