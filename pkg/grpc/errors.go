package grpc

import "peoplesuite/platform-sdk-go/pkg/errors"

func ToStatus(err error) error {
	return errors.ToGRPC(err)
}
