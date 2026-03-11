package grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// RegisterHealth registers the standard gRPC health service.
func RegisterHealth(srv *grpc.Server) {

	healthSrv := health.NewServer()

	healthpb.RegisterHealthServer(srv, healthSrv)

	healthSrv.SetServingStatus(
		"",
		healthpb.HealthCheckResponse_SERVING,
	)
}
