package devtools

import (
	"context"
	"net/http"
	"os"

	"github.com/fullstorydev/grpcui/standalone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewGRPCUIHandler creates an HTTP handler that serves grpcui backed by server reflection.
// If no dial options are provided, it dials the target with insecure transport credentials.
func NewGRPCUIHandler(ctx context.Context, target string, opts ...grpc.DialOption) (http.Handler, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(opts) == 0 {
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, err
	}

	return standalone.HandlerViaReflection(ctx, conn, target)
}

// ProtoDocsHandler serves static documentation generated from protobuf definitions.
// It expects docsDir to point at the output directory used by buf's pseudomuto-doc plugin.
func ProtoDocsHandler(docsDir string) http.Handler {
	// Allow callers to override the docs directory via an environment variable if needed.
	if override := os.Getenv("PROTO_DOCS_DIR"); override != "" {
		docsDir = override
	}

	return http.FileServer(http.Dir(docsDir))
}
