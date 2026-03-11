package main

import (
	"context"

	"go.uber.org/zap"

	pkggrpc "github.com/peoplesuite/platform-sdk-go/pkg/grpc"
	"github.com/peoplesuite/platform-sdk-go/pkg/runtime"
)

func main() {
	logger, _ := zap.NewProduction()

	// gRPC server with recovery, tracing, logging, health, and reflection
	grpcServer := pkggrpc.NewServer(pkggrpc.ServerConfig{
		Logger:           logger,
		EnableReflection: true,
	})

	// Runtime with gRPC only (no user HTTP handlers). HTTP server will still start to serve dev tools.
	rt, err := runtime.New(runtime.Options{
		ServiceName: "grpc-example",
		Version:     "1.0.0",
		Environment: "development",
		HTTPPort:    8080,
		GRPCPort:    9090,
		HTTPHandler: nil, // gRPC-only example
		GRPCServer:  grpcServer,
		DevTools: runtime.DevToolsConfig{
			Enabled:      true,
			GRPCTarget:   "localhost:9090",
			ProtoDocsDir: "./docs/proto",
		},
	})
	if err != nil {
		panic(err)
	}

	_ = rt.Run(context.Background())
}
