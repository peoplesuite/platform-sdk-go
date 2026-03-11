package interceptors

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestAuthUnary_NoMetadata_ReturnsUnauthenticated(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := AuthUnary(func(ctx context.Context, token string) (context.Context, error) {
		return ctx, nil
	}, logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Foo"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	// No metadata in context
	resp, err := interceptor(context.Background(), nil, info, handler)
	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}
	if err == nil {
		t.Fatal("err is nil")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", status.Code(err))
	}
	if status.Convert(err).Message() != "missing metadata" {
		t.Errorf("message = %q, want missing metadata", status.Convert(err).Message())
	}
}

func TestAuthUnary_EmptyAuthorization_ReturnsUnauthenticated(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := AuthUnary(func(ctx context.Context, token string) (context.Context, error) {
		return ctx, nil
	}, logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Foo"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("other", "x"))
	resp, err := interceptor(ctx, nil, info, handler)
	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}
	if err == nil {
		t.Fatal("err is nil")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", status.Code(err))
	}
}

func TestAuthUnary_AuthFuncError_ReturnsUnauthenticated(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := AuthUnary(func(ctx context.Context, token string) (context.Context, error) {
		return nil, errors.New("bad token")
	}, logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Foo"}
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "bearer x"))
	resp, err := interceptor(ctx, nil, info, handler)
	if resp != nil {
		t.Errorf("resp = %v, want nil", resp)
	}
	if err == nil {
		t.Fatal("err is nil")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("code = %v, want Unauthenticated", status.Code(err))
	}
	if status.Convert(err).Message() != "invalid token" {
		t.Errorf("message = %q, want invalid token", status.Convert(err).Message())
	}
}

type tokenKey struct{}

func TestAuthUnary_Success_CallsHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	interceptor := AuthUnary(func(ctx context.Context, token string) (context.Context, error) {
		return context.WithValue(ctx, tokenKey{}, token), nil
	}, logger)
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Foo"}
	called := false
	handler := grpc.UnaryHandler(func(ctx context.Context, req any) (any, error) {
		called = true
		if ctx.Value(tokenKey{}) != "secret" {
			t.Errorf("token in context = %v, want secret", ctx.Value(tokenKey{}))
		}
		return "ok", nil
	})

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "secret"))
	resp, err := interceptor(ctx, nil, info, handler)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if !called {
		t.Error("handler was not called")
	}
	if resp != "ok" {
		t.Errorf("resp = %v, want ok", resp)
	}
}
