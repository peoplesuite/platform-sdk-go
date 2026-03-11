package runtime

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/peoplesuite/platform-sdk-go/pkg/devtools"
	pkghttp "github.com/peoplesuite/platform-sdk-go/pkg/http"
)

type httpServer struct {
	server *http.Server
	logger *zap.Logger
}

func newHTTPServer(opts Options, logger *zap.Logger) *httpServer {

	// If there is no base HTTP handler and dev tools are disabled, we don't start HTTP.
	if opts.HTTPHandler == nil && !opts.DevTools.Enabled {
		return nil
	}

	handler := opts.HTTPHandler

	// Attach developer tooling endpoints (/grpcui, /proto-docs) when enabled.
	if opts.DevTools.Enabled {
		mux := http.NewServeMux()

		// Static proto documentation, if configured.
		if opts.DevTools.ProtoDocsDir != "" {
			mux.Handle("/proto-docs/", http.StripPrefix("/proto-docs/", devtools.ProtoDocsHandler(opts.DevTools.ProtoDocsDir)))
		}

		// Embedded grpcui, if a gRPC target has been provided.
		if opts.DevTools.GRPCTarget != "" {
			if h, err := devtools.NewGRPCUIHandler(context.Background(), opts.DevTools.GRPCTarget); err != nil {
				logger.Warn("failed to initialize grpcui handler", zap.Error(err))
			} else {
				mux.Handle("/grpcui/", http.StripPrefix("/grpcui", h))
			}
		}

		// Mount the user-provided handler last so that it can own the root path.
		if handler != nil {
			mux.Handle("/", handler)
		}

		handler = mux
	}

	cfg := pkghttp.DefaultServerConfig(handler, opts.HTTPPort)

	return &httpServer{
		server: pkghttp.NewServer(cfg),
		logger: logger,
	}
}

func (r *Runtime) startHTTP(ctx context.Context, errCh chan<- error) {

	if r.httpServer == nil {
		return
	}

	go func() {

		r.logger.Info("HTTP server starting",
			zap.String("addr", r.httpServer.server.Addr))

		if err := pkghttp.Serve(ctx, r.httpServer.server, r.logger); err != nil {
			errCh <- fmt.Errorf("http server: %w", err)
		}

	}()
}
