package interceptors

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/peoplesuite/platform-sdk-go/pkg/observability"
)

func TestMetricsUnary_SuccessUpdatesMetrics(t *testing.T) {
	m := observability.NewServerMetrics("test")
	reg := prometheus.NewRegistry()
	m.MustRegister(reg)

	interceptor := MetricsUnary(m)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	_, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("err = %v", err)
	}

	// Collect and ensure we have metrics (e.g. requests_total with code OK)
	metrics, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	if len(metrics) == 0 {
		t.Error("expected metrics after MetricsUnary")
	}
}

func TestMetricsUnary_ErrorReturnsError(t *testing.T) {
	m := observability.NewServerMetrics("test")
	interceptor := MetricsUnary(m)
	info := &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "not found")
	})

	resp, err := interceptor(context.Background(), nil, info, handler)
	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}
	if err == nil {
		t.Fatal("expected error")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("code = %v, want NotFound", status.Code(err))
	}
}

func TestMetricsStream_Success(t *testing.T) {
	m := observability.NewServerMetrics("test")
	interceptor := MetricsStream(m)
	info := &grpc.StreamServerInfo{FullMethod: "/svc/Stream"}
	handler := grpc.StreamHandler(func(srv any, ss grpc.ServerStream) error {
		return nil
	})
	ss := &mockServerStream{ctx: context.Background()}

	err := interceptor(nil, ss, info, handler)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
}
