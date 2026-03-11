package errors

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HTTPStatus(err error) int {
	var e *Error

	if err != nil && errors.As(err, &e) {
		return e.status
	}

	return 500
}

func HTTPMessage(err error) string {
	var e *Error

	if err != nil && errors.As(err, &e) {
		return e.message
	}

	if err != nil {
		return err.Error()
	}

	return "internal error"
}

func GetKind(err error) Kind {
	var e *Error

	if err != nil && errors.As(err, &e) {
		return e.kind
	}

	return KindInternal
}

func ToGRPC(err error) error {
	if err == nil {
		return nil
	}

	code := kindToGRPCCode(GetKind(err))

	return status.Error(code, HTTPMessage(err))
}

func FromGRPC(err error) *Error {

	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return nil
	}

	kind := codeToKind(st.Code())

	return mk(
		HTTPStatusFromCode(st.Code()),
		kind,
		st.Message(),
		nil,
	)
}

func kindToStatus(kind Kind) int {
	switch kind {
	case KindInvalidArgument:
		return 400
	case KindUnauthenticated:
		return 401
	case KindPermissionDenied:
		return 403
	case KindNotFound:
		return 404
	case KindAlreadyExists, KindConflict:
		return 409
	case KindPreconditionFailed:
		return 412
	case KindRateLimited:
		return 429
	case KindUnavailable:
		return 503
	case KindTimeout:
		return 504
	default:
		return 500
	}
}

func kindToGRPCCode(kind Kind) codes.Code {
	switch kind {
	case KindInvalidArgument:
		return codes.InvalidArgument
	case KindNotFound:
		return codes.NotFound
	case KindAlreadyExists:
		return codes.AlreadyExists
	case KindPermissionDenied:
		return codes.PermissionDenied
	case KindUnauthenticated:
		return codes.Unauthenticated
	case KindConflict:
		return codes.Aborted
	case KindPreconditionFailed:
		return codes.FailedPrecondition
	case KindUnavailable:
		return codes.Unavailable
	case KindTimeout:
		return codes.DeadlineExceeded
	case KindRateLimited:
		return codes.ResourceExhausted
	default:
		return codes.Internal
	}
}

func codeToKind(code codes.Code) Kind {
	switch code {
	case codes.InvalidArgument:
		return KindInvalidArgument
	case codes.NotFound:
		return KindNotFound
	case codes.AlreadyExists:
		return KindAlreadyExists
	case codes.PermissionDenied:
		return KindPermissionDenied
	case codes.Unauthenticated:
		return KindUnauthenticated
	case codes.Aborted:
		return KindConflict
	case codes.FailedPrecondition:
		return KindPreconditionFailed
	case codes.Unavailable:
		return KindUnavailable
	case codes.DeadlineExceeded:
		return KindTimeout
	case codes.ResourceExhausted:
		return KindRateLimited
	default:
		return KindInternal
	}
}

// HTTPStatusFromCode returns the HTTP status code for a gRPC code.
func HTTPStatusFromCode(code codes.Code) int {
	return kindToStatus(codeToKind(code))
}
