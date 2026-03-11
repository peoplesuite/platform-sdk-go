package httpclient

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// startSpan creates a span for an outbound HTTP request.
func startSpan(
	ctx context.Context,
	tracer trace.Tracer,
	method string,
	url string,
) (context.Context, trace.Span) {

	if tracer == nil {
		return ctx, nil
	}

	ctx, span := tracer.Start(
		ctx,
		"httpclient."+method,
		trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", url),
		),
	)

	return ctx, span
}

// finishSpan records the response status.
func finishSpan(span trace.Span, resp *http.Response, err error) {

	if span == nil {
		return
	}

	defer span.End()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	if resp != nil {
		span.SetAttributes(
			attribute.Int("http.status_code", resp.StatusCode),
		)
	}
}
