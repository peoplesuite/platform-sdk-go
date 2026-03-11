package runtime

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	pkghttp "github.com/peoplesuite/platform-sdk-go/pkg/http"
)

type httpServer struct {
	server *http.Server
	logger *zap.Logger
}

func newHTTPServer(opts Options, logger *zap.Logger) *httpServer {

	if opts.HTTPHandler == nil {
		return nil
	}

	cfg := pkghttp.DefaultServerConfig(opts.HTTPHandler, opts.HTTPPort)

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
