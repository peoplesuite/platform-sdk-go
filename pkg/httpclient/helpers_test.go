package httpclient

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestStartSpan_NilTracer(t *testing.T) {
	ctx := context.Background()
	outCtx, span := startSpan(ctx, nil, "GET", "http://example.com")
	require.NotNil(t, outCtx)
	assert.Nil(t, span)
}

func TestStartSpan_WithTracer(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	ctx := context.Background()
	outCtx, span := startSpan(ctx, tracer, "POST", "http://example.com/foo")
	require.NotNil(t, span)
	require.NotNil(t, outCtx)
	span.End()
}

func TestFinishSpan_NilSpan(t *testing.T) {
	finishSpan(nil, &http.Response{StatusCode: 200}, nil)
}

func TestFinishSpan_WithError(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	_, span := tracer.Start(context.Background(), "test")
	finishSpan(span, nil, errors.New("request failed"))
}

func TestFinishSpan_WithResp(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	_, span := tracer.Start(context.Background(), "test")
	finishSpan(span, &http.Response{StatusCode: 201}, nil)
}
