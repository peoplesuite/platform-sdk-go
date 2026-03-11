package grpc

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/peoplesuite/platform-sdk-go/pkg/errors"
)

func TestToStatus_Nil(t *testing.T) {
	if err := ToStatus(nil); err != nil {
		t.Errorf("ToStatus(nil) = %v, want nil", err)
	}
}

func TestToStatus_NotFound(t *testing.T) {
	err := ToStatus(errors.NotFound("missing"))
	if err == nil {
		t.Fatal("ToStatus(NotFound) should not return nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("err is not status: %v", err)
	}
	if st.Code() != codes.NotFound {
		t.Errorf("code = %v, want NotFound", st.Code())
	}
}

func TestToStatus_Unauthenticated(t *testing.T) {
	err := ToStatus(errors.Unauthenticated("no token"))
	if err == nil {
		t.Fatal("ToStatus(Unauthenticated) should not return nil")
	}
	st, _ := status.FromError(err)
	if st.Code() != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", st.Code())
	}
}
