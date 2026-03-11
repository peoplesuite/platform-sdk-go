package runtime

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
)

// Worker is a function that runs until context is cancelled.
type Worker func(context.Context) error

// Options configures the runtime (HTTP, gRPC, workers, hooks).
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

// DefaultOptions returns options with default HTTP and gRPC ports.
func DefaultOptions() Options {
	return Options{
		HTTPPort: 8080,
		GRPCPort: 9090,
	}
}
