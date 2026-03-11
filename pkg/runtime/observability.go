package runtime

import (
	"go.uber.org/zap"

	"github.com/peoplesuite/platform-sdk-go/pkg/observability/logging"
	"github.com/peoplesuite/platform-sdk-go/pkg/observability/tracing"
)

func initLogger(opts Options) (*zap.Logger, error) {

	tracing.Init(tracing.Config{
		ServiceName:    opts.ServiceName,
		ServiceVersion: opts.Version,
		Environment:    opts.Environment,
	})

	return logging.NewLogger(logging.Config{
		ServiceName:    opts.ServiceName,
		ServiceVersion: opts.Version,
		Environment:    opts.Environment,
		LogLevel:       "info",
	})
}
