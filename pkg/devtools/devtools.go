package devtools

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/fullstorydev/grpcui/standalone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// lazyGRPCUIHandler defers connecting to the gRPC target until the first request.
// This avoids startup races where the HTTP server is constructed before the
// gRPC server has begun listening, which would otherwise cause an immediate
// connection failure and prevent /grpcui from being registered at all.
type lazyGRPCUIHandler struct {
	mu     sync.Mutex
	init   sync.Once
	ctx    context.Context
	target string
	opts   []grpc.DialOption

	inner http.Handler
	err   error
}

func (h *lazyGRPCUIHandler) ensureInitialized() {
	h.init.Do(func() {
		ctx := h.ctx
		if ctx == nil {
			ctx = context.Background()
		}

		opts := h.opts
		if len(opts) == 0 {
			opts = []grpc.DialOption{
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			}
		}

		conn, err := grpc.DialContext(ctx, h.target, opts...)
		if err != nil {
			h.err = err
			return
		}

		h.inner, h.err = standalone.HandlerViaReflection(ctx, conn, h.target)
	})
}

func (h *lazyGRPCUIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ensureInitialized()

	if h.err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintf(w, "grpcui is not available yet: %v", h.err)
		return
	}

	h.inner.ServeHTTP(w, r)
}

// NewGRPCUIHandler creates an HTTP handler that serves grpcui backed by server reflection.
// Unlike the original implementation, this version defers dialing the target until the
// first request is received, so that /grpcui can always be registered at startup even if
// the gRPC server is not yet ready to accept connections.
func NewGRPCUIHandler(ctx context.Context, target string, opts ...grpc.DialOption) (http.Handler, error) {
	return &lazyGRPCUIHandler{
		ctx:    ctx,
		target: target,
		opts:   opts,
	}, nil
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
