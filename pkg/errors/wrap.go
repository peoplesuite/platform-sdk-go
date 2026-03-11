package errors

func Internal(msg string) error {
	return mk(500, KindInternal, msg, nil)
}

func InternalWrap(msg string, cause error) error {
	return mk(500, KindInternal, msg, cause)
}

func InvalidArgument(msg string) error {
	return mk(400, KindInvalidArgument, msg, nil)
}

func NotFound(msg string) error {
	return mk(404, KindNotFound, msg, nil)
}

func AlreadyExists(msg string) error {
	return mk(409, KindAlreadyExists, msg, nil)
}

func PermissionDenied(msg string) error {
	return mk(403, KindPermissionDenied, msg, nil)
}

func Unauthenticated(msg string) error {
	return mk(401, KindUnauthenticated, msg, nil)
}

func Conflict(msg string) error {
	return mk(409, KindConflict, msg, nil)
}

func PreconditionFailed(msg string) error {
	return mk(412, KindPreconditionFailed, msg, nil)
}

func Unavailable(msg string) error {
	return mk(503, KindUnavailable, msg, nil)
}

func Timeout(msg string) error {
	return mk(504, KindTimeout, msg, nil)
}

func RateLimited(msg string) error {
	return mk(429, KindRateLimited, msg, nil)
}

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
