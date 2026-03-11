package main

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pkghttp "peoplesuite/platform-sdk-go/pkg/http"
	"peoplesuite/platform-sdk-go/pkg/runtime"
)

func main() {
	logger, _ := zap.NewProduction()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	handler := pkghttp.Chain(
		pkghttp.RequestID,
		pkghttp.Logging(logger),
		pkghttp.Recovery(logger),
	)(mux)

	// Runtime needs both HTTP and gRPC; use minimal gRPC server with no services
	grpcServer := grpc.NewServer()

	rt, err := runtime.New(runtime.Options{
		ServiceName: "http-example",
		Version:     "1.0.0",
		Environment: "development",
		HTTPPort:    8080,
		GRPCPort:    9090,
		HTTPHandler: handler,
		GRPCServer:  grpcServer,
	})
	if err != nil {
		panic(err)
	}

	rt.Run(context.Background())
}
