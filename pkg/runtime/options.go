package runtime

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
)

type Worker func(context.Context) error

type Options struct {
	ServiceName string
	Version     string
	Environment string

	HTTPPort int
	GRPCPort int

	HTTPHandler http.Handler
	GRPCServer  *grpc.Server

	Workers []Worker

	StartHooks []func(context.Context) error
	StopHooks  []func(context.Context) error
}

func DefaultOptions() Options {
	return Options{
		HTTPPort: 8080,
		GRPCPort: 9090,
	}
}
