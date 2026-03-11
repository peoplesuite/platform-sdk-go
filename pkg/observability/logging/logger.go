package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config defines logger configuration.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	LogLevel       string
}

// NewLogger builds a zap logger suitable for Datadog ingestion.
func NewLogger(cfg Config) (*zap.Logger, error) {
	zapCfg := zap.NewProductionConfig()

	if cfg.Environment != "production" {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	level, err := zapcore.ParseLevel(cfg.LogLevel)
	if err == nil {
		zapCfg.Level = zap.NewAtomicLevelAt(level)
	}

	logger, err := zapCfg.Build(
		zap.Fields(
			zap.String("service", cfg.ServiceName),
			zap.String("version", cfg.ServiceVersion),
			zap.String("env", cfg.Environment),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	zap.ReplaceGlobals(logger)

	return logger, nil
}

// MustLogger panics on error.
func MustLogger(cfg Config) *zap.Logger {
	l, err := NewLogger(cfg)
	if err != nil {
		panic(err)
	}
	return l
}
