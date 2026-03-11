package errors

import "testing"

func TestKind_String_AllDefinedKinds(t *testing.T) {
	tests := []struct {
		kind Kind
		want string
	}{
		{KindInvalidArgument, "InvalidArgument"},
		{KindNotFound, "NotFound"},
		{KindAlreadyExists, "AlreadyExists"},
		{KindPermissionDenied, "PermissionDenied"},
		{KindUnauthenticated, "Unauthenticated"},
		{KindConflict, "Conflict"},
		{KindPreconditionFailed, "PreconditionFailed"},
		{KindUnavailable, "Unavailable"},
		{KindTimeout, "Timeout"},
		{KindRateLimited, "RateLimited"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Fatalf("Kind(%v).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestKind_String_DefaultCase(t *testing.T) {
	unknown := Kind(-1)
	if got := unknown.String(); got != "Internal" {
		t.Fatalf("Kind(-1).String() = %q, want %q", got, "Internal")
	}
}
