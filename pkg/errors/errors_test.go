package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

func TestError_ErrorWithoutCause(t *testing.T) {
	e := mk(500, KindInternal, "oops", nil)

	if got := e.Error(); got != "oops" {
		t.Fatalf("Error() = %q, want %q", got, "oops")
	}
}

func TestError_ErrorWithCause(t *testing.T) {
	cause := stderrors.New("root")
	e := mk(500, KindInternal, "wrap", cause)

	got := e.Error()
	want := "wrap: root"
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("cause")
	e := mk(500, KindInternal, "msg", cause)

	if unwrapped := e.Unwrap(); unwrapped != cause {
		t.Fatalf("Unwrap() = %#v, want %#v", unwrapped, cause)
	}

	eNoCause := mk(500, KindInternal, "msg", nil)
	if unwrapped := eNoCause.Unwrap(); unwrapped != nil {
		t.Fatalf("Unwrap() on no-cause error = %#v, want nil", unwrapped)
	}
}

func TestError_StatusKindMessage(t *testing.T) {
	e := mk(404, KindNotFound, "missing", nil)

	if got := e.Status(); got != 404 {
		t.Fatalf("Status() = %d, want %d", got, 404)
	}
	if got := e.Kind(); got != KindNotFound {
		t.Fatalf("Kind() = %v, want %v", got, KindNotFound)
	}
	if got := e.Message(); got != "missing" {
		t.Fatalf("Message() = %q, want %q", got, "missing")
	}
}
