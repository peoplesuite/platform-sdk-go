package errors

import (
	stderrors "errors"
	"testing"
)

func TestConstructors_BasicProperties(t *testing.T) {
	tests := []struct {
		name    string
		makeErr func() error
		kind    Kind
		status  int
		msg     string
	}{
		{"Internal", func() error { return Internal("msg") }, KindInternal, 500, "msg"},
		{"InvalidArgument", func() error { return InvalidArgument("msg") }, KindInvalidArgument, 400, "msg"},
		{"NotFound", func() error { return NotFound("msg") }, KindNotFound, 404, "msg"},
		{"AlreadyExists", func() error { return AlreadyExists("msg") }, KindAlreadyExists, 409, "msg"},
		{"PermissionDenied", func() error { return PermissionDenied("msg") }, KindPermissionDenied, 403, "msg"},
		{"Unauthenticated", func() error { return Unauthenticated("msg") }, KindUnauthenticated, 401, "msg"},
		{"Conflict", func() error { return Conflict("msg") }, KindConflict, 409, "msg"},
		{"PreconditionFailed", func() error { return PreconditionFailed("msg") }, KindPreconditionFailed, 412, "msg"},
		{"Unavailable", func() error { return Unavailable("msg") }, KindUnavailable, 503, "msg"},
		{"Timeout", func() error { return Timeout("msg") }, KindTimeout, 504, "msg"},
		{"RateLimited", func() error { return RateLimited("msg") }, KindRateLimited, 429, "msg"},
	}

	for _, tt := range tests {
		err := tt.makeErr()
		if err == nil {
			t.Fatalf("%s returned nil error", tt.name)
		}

		if got := HTTPStatus(err); got != tt.status {
			t.Fatalf("%s HTTPStatus() = %d, want %d", tt.name, got, tt.status)
		}
		if got := GetKind(err); got != tt.kind {
			t.Fatalf("%s GetKind() = %v, want %v", tt.name, got, tt.kind)
		}
		if got := HTTPMessage(err); got != tt.msg {
			t.Fatalf("%s HTTPMessage() = %q, want %q", tt.name, got, tt.msg)
		}
	}
}

func TestInternalWrap(t *testing.T) {
	cause := stderrors.New("root")
	err := InternalWrap("wrap", cause)

	e, ok := err.(*Error)
	if !ok {
		t.Fatalf("InternalWrap did not return *Error, got %T", err)
	}
	if e.Kind() != KindInternal || e.Status() != 500 {
		t.Fatalf("unexpected kind/status: kind=%v status=%d", e.Kind(), e.Status())
	}
	if unwrapped := e.Unwrap(); unwrapped != cause {
		t.Fatalf("Unwrap() = %#v, want %#v", unwrapped, cause)
	}
	if got := e.Error(); got != "wrap: root" {
		t.Fatalf("Error() = %q, want %q", got, "wrap: root")
	}
}

func TestWrap_NilCauseReturnsNil(t *testing.T) {
	if err := Wrap(KindNotFound, "op", nil); err != nil {
		t.Fatalf("expected nil cause to return nil error, got %v", err)
	}
}

func TestWrap_WithAndWithoutOp(t *testing.T) {
	cause := stderrors.New("boom")

	errNoOp := Wrap(KindNotFound, "", cause)
	if errNoOp == nil {
		t.Fatalf("Wrap without op returned nil")
	}
	eNoOp := errNoOp.(*Error)
	if eNoOp.Message() != "boom" {
		t.Fatalf("message without op = %q, want %q", eNoOp.Message(), "boom")
	}
	if eNoOp.Kind() != KindNotFound {
		t.Fatalf("unexpected kind: %v", eNoOp.Kind())
	}

	errWithOp := Wrap(KindNotFound, "op", cause)
	eWithOp := errWithOp.(*Error)
	if eWithOp.Message() != "op: boom" {
		t.Fatalf("message with op = %q, want %q", eWithOp.Message(), "op: boom")
	}
	if eWithOp.Unwrap() != cause {
		t.Fatalf("unwrap = %#v, want %#v", eWithOp.Unwrap(), cause)
	}
}
