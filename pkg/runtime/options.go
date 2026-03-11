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

	// DevTools controls optional developer tooling endpoints such as /grpcui and /proto-docs.
	DevTools DevToolsConfig

	Workers []Worker

	StartHooks []func(context.Context) error
	StopHooks  []func(context.Context) error
}

// DevToolsConfig configures optional developer tooling endpoints.
type DevToolsConfig struct {
	// Enabled toggles all dev tools. When false, runtime does not attach any dev routes.
	Enabled bool

	// GRPCTarget is the address grpcui should use when connecting via reflection, e.g. "localhost:9091".
	GRPCTarget string

	// ProtoDocsDir is the directory containing static documentation generated from protobuf definitions.
	// Typically this is the buf pseudomuto-doc output directory, for example "./docs/proto".
	ProtoDocsDir string
}

// DefaultOptions returns options with default HTTP and gRPC ports.
func DefaultOptions() Options {
	return Options{
		HTTPPort: 8080,
		GRPCPort: 9090,
	}
}
