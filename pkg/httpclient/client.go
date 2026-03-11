// Package httpclient provides an HTTP client with JSON helpers, optional base URL,
// TLS options, retry with exponential backoff on 5xx/connection errors, and optional OTel tracing.
// Use for outbound service calls; transport-level errors use HTTPError (status + body).
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	retryMax   int
	retryWait  time.Duration
	retryMaxW  time.Duration
	tracer     trace.Tracer
}

func New(opts Options) *Client {
	retryMax := opts.RetryMax
	if retryMax <= 0 {
		retryMax = 3
	}
	retryWait := opts.RetryWait
	if retryWait <= 0 {
		retryWait = time.Second
	}
	retryMaxW := opts.RetryMaxWait
	if retryMaxW <= 0 {
		retryMaxW = 30 * time.Second
	}

	transport := newTransport(opts)

	return &Client{
		httpClient: &http.Client{Timeout: opts.Timeout, Transport: transport},
		baseURL:    opts.BaseURL,
		retryMax:   retryMax,
		retryWait:  retryWait,
		retryMaxW:  retryMaxW,
		tracer:     opts.Tracer,
	}
}

// shouldRetry returns true for 5xx responses or temporary/connection errors.
func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		var netErr *net.OpError
		return errors.As(err, &netErr)
	}
	if resp != nil && resp.StatusCode >= 500 {
		return true
	}
	return false
}

// createReqFn builds a new request per attempt (so body can be re-read on retry).
type createReqFn func() (*http.Request, error)

// doWithRetry runs createReq+Do up to c.retryMax times with exponential backoff on 5xx/connection errors.
func (c *Client) doWithRetry(ctx context.Context, createReq createReqFn) (*http.Response, error) {
	var lastErr error
	wait := c.retryWait
	for attempt := 0; attempt < c.retryMax; attempt++ {
		req, err := createReq()
		if err != nil {
			return nil, err
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.retryMax-1 && shouldRetry(nil, err) {
				timer := time.NewTimer(wait)
				select {
				case <-ctx.Done():
					timer.Stop()
					return nil, ctx.Err()
				case <-timer.C:
				}
				if wait < c.retryMaxW {
					wait *= 2
					if wait > c.retryMaxW {
						wait = c.retryMaxW
					}
				}
				continue
			}
			return nil, err
		}
		if attempt < c.retryMax-1 && shouldRetry(resp, nil) {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			timer := time.NewTimer(wait)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
			if wait < c.retryMaxW {
				wait *= 2
				if wait > c.retryMaxW {
					wait = c.retryMaxW
				}
			}
			continue
		}
		return resp, nil
	}
	return nil, lastErr
}

// do runs the request with optional span and retry. createReq is called once per attempt.
func (c *Client) do(ctx context.Context, method, urlStr string, createReq createReqFn) (*http.Response, error) {
	if c.tracer != nil {
		ctx, span := c.tracer.Start(ctx, "httpclient."+method, trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", urlStr),
		))
		defer span.End()
		resp, err := c.doWithRetry(ctx, createReq)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		return resp, nil
	}
	return c.doWithRetry(ctx, createReq)
}

func (c *Client) DoJSON(
	ctx context.Context,
	method string,
	endpoint string,
	headers map[string]string,
	reqBody any,
	respBody any,
) (*http.Response, error) {
	urlStr := c.baseURL + endpoint
	createReq := func() (*http.Request, error) {
		var bodyReader io.Reader
		if reqBody != nil {
			data, err := json.Marshal(reqBody)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal JSON: %w", err)
			}
			bodyReader = bytes.NewBuffer(data)
		}
		req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req, nil
	}
	resp, err := c.do(ctx, method, urlStr, createReq)
	if err != nil {
		return nil, err
	}
	if respBody != nil {
		if err := DecodeJSONResponse(resp, respBody); err != nil {
			return resp, err
		}
	}
	return resp, nil
}

// GetRaw performs a GET request and returns the raw response body as bytes.
// This is useful for large responses where you want to avoid unmarshaling into memory.
// The response body is fully read and the connection is closed.
func (c *Client) GetRaw(
	ctx context.Context,
	endpoint string,
	headers map[string]string,
) ([]byte, *http.Response, error) {
	urlStr := c.baseURL + endpoint
	createReq := func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req, nil
	}
	resp, err := c.do(ctx, http.MethodGet, urlStr, createReq)
	if err != nil {
		return nil, nil, err
	}
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, resp, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, resp, nil
}

// GetRawToFile performs a GET request and streams the response body directly to a temporary file.
// This is optimized for very large responses (>20MB) to avoid loading the entire response into memory.
// The file is created in the system's temporary directory with a unique name.
// Returns the file path, HTTP response, and any error.
// The caller is responsible for deleting the temporary file after use.
func (c *Client) GetRawToFile(
	ctx context.Context,
	endpoint string,
	headers map[string]string,
) (string, *http.Response, error) {
	urlStr := c.baseURL + endpoint
	createReq := func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return req, nil
	}
	resp, err := c.do(ctx, http.MethodGet, urlStr, createReq)
	if err != nil {
		return "", nil, err
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "http-response-*.tmp")
	if err != nil {
		_ = resp.Body.Close()
		return "", resp, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpFilePath := tmpFile.Name()

	// Stream response body to file
	written, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFilePath) // Clean up on error
		_ = resp.Body.Close()
		return "", resp, fmt.Errorf("failed to write response to file: %w", err)
	}

	// Close file and response body
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFilePath) // Clean up on error
		_ = resp.Body.Close()
		return "", resp, fmt.Errorf("failed to close temp file: %w", err)
	}
	_ = resp.Body.Close()

	// Verify file was written
	if written == 0 {
		_ = os.Remove(tmpFilePath) // Clean up empty file
		return "", resp, fmt.Errorf("response body was empty")
	}

	return tmpFilePath, resp, nil
}

// Do performs a raw HTTP request with the provided request object.
// Use for one-off requests (e.g. multipart/form-data). Retry is not applied because the body
// cannot be re-read; use DoJSON or GetRaw for retry and tracing. If tracer is set, a span is
// still recorded.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if req.Context() == nil {
		req = req.WithContext(ctx)
	} else {
		req = req.Clone(ctx)
	}
	if c.baseURL != "" && req.URL != nil && !req.URL.IsAbs() {
		base, err := url.Parse(c.baseURL)
		if err == nil {
			req.URL = base.ResolveReference(req.URL)
		}
	}
	method := req.Method
	if method == "" {
		method = http.MethodGet
	}
	urlStr := req.URL.String()
	if c.tracer != nil {
		ctx, span := c.tracer.Start(ctx, "httpclient."+method, trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", urlStr),
		))
		defer span.End()
		req = req.WithContext(ctx)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		return resp, nil
	}
	return c.httpClient.Do(req)
}
