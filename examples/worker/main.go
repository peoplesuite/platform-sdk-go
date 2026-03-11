package main

import (
	"context"
	"time"

	"peoplesuite/platform-sdk-go/pkg/runtime"
)

func main() {
	// Worker that runs until context is cancelled
	worker := func(ctx context.Context) error {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				// In a real app you might poll a queue, process jobs, etc.
				// Here we just run until shutdown.
			}
		}
	}

	rt, err := runtime.New(runtime.Options{
		ServiceName: "worker-example",
		Version:     "1.0.0",
		Environment: "development",
		HTTPPort:    8080,
		GRPCPort:    9090,
		HTTPHandler: nil,
		GRPCServer:  nil,
		Workers:     []runtime.Worker{worker},
	})
	if err != nil {
		panic(err)
	}

	rt.Run(context.Background())
}
