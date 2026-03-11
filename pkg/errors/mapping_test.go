package errors

import (
	stderrors "errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHTTPStatusAndMessageAndGetKind(t *testing.T) {
	base := NotFound("missing")

	if got := HTTPStatus(base); got != 404 {
		t.Fatalf("HTTPStatus() = %d, want %d", got, 404)
	}
	if got := HTTPMessage(base); got != "missing" {
		t.Fatalf("HTTPMessage() = %q, want %q", got, "missing")
	}
	if got := GetKind(base); got != KindNotFound {
		t.Fatalf("GetKind() = %v, want %v", got, KindNotFound)
	}

	plain := stderrors.New("plain")
	if got := HTTPStatus(plain); got != 500 {
		t.Fatalf("HTTPStatus(plain) = %d, want 500", got)
	}
	if got := HTTPMessage(plain); got != "plain" {
		t.Fatalf("HTTPMessage(plain) = %q, want %q", got, "plain")
	}
	if got := GetKind(plain); got != KindInternal {
		t.Fatalf("GetKind(plain) = %v, want %v", got, KindInternal)
	}

	if got := HTTPStatus(nil); got != 500 {
		t.Fatalf("HTTPStatus(nil) = %d, want 500", got)
	}
	if got := HTTPMessage(nil); got != "internal error" {
		t.Fatalf("HTTPMessage(nil) = %q, want %q", got, "internal error")
	}
	if got := GetKind(nil); got != KindInternal {
		t.Fatalf("GetKind(nil) = %v, want %v", got, KindInternal)
	}
}

func TestHTTPStatus_AllKinds(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{InvalidArgument("x"), 400},
		{Unauthenticated("x"), 401},
		{PermissionDenied("x"), 403},
		{NotFound("x"), 404},
		{AlreadyExists("x"), 409},
		{Conflict("x"), 409},
		{PreconditionFailed("x"), 412},
		{RateLimited("x"), 429},
		{Unavailable("x"), 503},
		{Timeout("x"), 504},
	}
	for _, tt := range tests {
		if got := HTTPStatus(tt.err); got != tt.status {
			t.Errorf("HTTPStatus(%v) = %d, want %d", tt.err, got, tt.status)
		}
	}
	// default/unknown kind -> 500
	if got := HTTPStatus(stderrors.New("plain")); got != 500 {
		t.Errorf("HTTPStatus(plain) = %d, want 500", got)
	}
}

func TestToGRPC_NilAndKinds(t *testing.T) {
	if err := ToGRPC(nil); err != nil {
		t.Fatalf("ToGRPC(nil) = %v, want nil", err)
	}

	tests := []struct {
		err  error
		code codes.Code
	}{
		{InvalidArgument("bad"), codes.InvalidArgument},
		{NotFound("missing"), codes.NotFound},
		{AlreadyExists("dup"), codes.AlreadyExists},
		{PermissionDenied("nope"), codes.PermissionDenied},
		{Unauthenticated("auth"), codes.Unauthenticated},
		{Conflict("conflict"), codes.Aborted},
		{PreconditionFailed("pre"), codes.FailedPrecondition},
		{Unavailable("down"), codes.Unavailable},
		{Timeout("slow"), codes.DeadlineExceeded},
		{RateLimited("limit"), codes.ResourceExhausted},
	}

	for _, tt := range tests {
		grpcErr := ToGRPC(tt.err)
		if grpcErr == nil {
			t.Fatalf("ToGRPC returned nil for %v", tt.err)
		}
		st, ok := status.FromError(grpcErr)
		if !ok {
			t.Fatalf("status.FromError failed for %v", grpcErr)
		}
		if st.Code() != tt.code {
			t.Fatalf("code = %v, want %v", st.Code(), tt.code)
		}
		if st.Message() != HTTPMessage(tt.err) {
			t.Fatalf("message = %q, want %q", st.Message(), HTTPMessage(tt.err))
		}
	}
}

func TestFromGRPC_BasicAndNonStatus(t *testing.T) {
	if got := FromGRPC(nil); got != nil {
		t.Fatalf("FromGRPC(nil) = %#v, want nil", got)
	}

	grpcErr := status.Error(codes.NotFound, "missing")
	got := FromGRPC(grpcErr)
	if got == nil {
		t.Fatalf("FromGRPC returned nil for gRPC error")
	}
	if got.Status() != HTTPStatusFromCode(codes.NotFound) {
		t.Fatalf("Status() = %d, want %d", got.Status(), HTTPStatusFromCode(codes.NotFound))
	}
	if got.Kind() != codeToKind(codes.NotFound) {
		t.Fatalf("Kind() = %v, want %v", got.Kind(), codeToKind(codes.NotFound))
	}
	if got.Message() != "missing" {
		t.Fatalf("Message() = %q, want %q", got.Message(), "missing")
	}

	nonStatus := stderrors.New("plain")
	if got := FromGRPC(nonStatus); got != nil {
		t.Fatalf("FromGRPC(non-status) = %#v, want nil", got)
	}
}

func TestFromGRPC_AllCodesAndHTTPStatusFromCode(t *testing.T) {
	tests := []struct {
		code       codes.Code
		kind       Kind
		httpStatus int
	}{
		{codes.InvalidArgument, KindInvalidArgument, 400},
		{codes.NotFound, KindNotFound, 404},
		{codes.AlreadyExists, KindAlreadyExists, 409},
		{codes.PermissionDenied, KindPermissionDenied, 403},
		{codes.Unauthenticated, KindUnauthenticated, 401},
		{codes.Aborted, KindConflict, 409},
		{codes.FailedPrecondition, KindPreconditionFailed, 412},
		{codes.Unavailable, KindUnavailable, 503},
		{codes.DeadlineExceeded, KindTimeout, 504},
		{codes.ResourceExhausted, KindRateLimited, 429},
	}
	for _, tt := range tests {
		grpcErr := status.Error(tt.code, "msg")
		got := FromGRPC(grpcErr)
		if got == nil {
			t.Fatalf("FromGRPC(%v) returned nil", tt.code)
		}
		if got.Kind() != tt.kind {
			t.Errorf("FromGRPC(%v).Kind() = %v, want %v", tt.code, got.Kind(), tt.kind)
		}
		if got.Status() != tt.httpStatus {
			t.Errorf("FromGRPC(%v).Status() = %d, want %d", tt.code, got.Status(), tt.httpStatus)
		}
		if s := HTTPStatusFromCode(tt.code); s != tt.httpStatus {
			t.Errorf("HTTPStatusFromCode(%v) = %d, want %d", tt.code, s, tt.httpStatus)
		}
	}
	// default: unknown gRPC code -> KindInternal, 500
	if got := HTTPStatusFromCode(codes.Unknown); got != 500 {
		t.Errorf("HTTPStatusFromCode(Unknown) = %d, want 500", got)
	}
	unknownErr := status.Error(codes.Unknown, "unknown")
	got := FromGRPC(unknownErr)
	if got == nil {
		t.Fatal("FromGRPC(Unknown) returned nil")
	}
	if got.Kind() != KindInternal {
		t.Errorf("FromGRPC(Unknown).Kind() = %v, want KindInternal", got.Kind())
	}
	if got.Status() != 500 {
		t.Errorf("FromGRPC(Unknown).Status() = %d, want 500", got.Status())
	}
}

func TestGRPCRoundTrip(t *testing.T) {
	src := NotFound("missing")

	grpcErr := ToGRPC(src)
	if grpcErr == nil {
		t.Fatalf("ToGRPC returned nil")
	}

	back := FromGRPC(grpcErr)
	if back == nil {
		t.Fatalf("FromGRPC returned nil")
	}

	if back.Kind() != KindNotFound {
		t.Fatalf("round-trip Kind() = %v, want %v", back.Kind(), KindNotFound)
	}
	if back.Status() != HTTPStatus(src) {
		t.Fatalf("round-trip Status() = %d, want %d", back.Status(), HTTPStatus(src))
	}
	if back.Message() != HTTPMessage(src) {
		t.Fatalf("round-trip Message() = %q, want %q", back.Message(), HTTPMessage(src))
	}
}
