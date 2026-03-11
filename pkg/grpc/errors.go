package grpc

import "github.com/peoplesuite/platform-sdk-go/pkg/errors"

// ToStatus converts a platform error to a gRPC status error.
func ToStatus(err error) error {
	return errors.ToGRPC(err)
}
