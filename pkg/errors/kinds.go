package errors

// Kind identifies the error category.
type Kind int

// Error kinds for HTTP/gRPC mapping.
const (
	KindInternal Kind = iota
	KindInvalidArgument
	KindNotFound
	KindAlreadyExists
	KindPermissionDenied
	KindUnauthenticated
	KindConflict
	KindPreconditionFailed
	KindUnavailable
	KindTimeout
	KindRateLimited
)

func (k Kind) String() string {
	switch k {
	case KindInvalidArgument:
		return "InvalidArgument"
	case KindNotFound:
		return "NotFound"
	case KindAlreadyExists:
		return "AlreadyExists"
	case KindPermissionDenied:
		return "PermissionDenied"
	case KindUnauthenticated:
		return "Unauthenticated"
	case KindConflict:
		return "Conflict"
	case KindPreconditionFailed:
		return "PreconditionFailed"
	case KindUnavailable:
		return "Unavailable"
	case KindTimeout:
		return "Timeout"
	case KindRateLimited:
		return "RateLimited"
	default:
		return "Internal"
	}
}
