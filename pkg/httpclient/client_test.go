package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestNew_DefaultOptions(t *testing.T) {
	c := New(Options{})
	require.NotNil(t, c)
	require.NotNil(t, c.httpClient)
	assert.Equal(t, 3, c.retryMax)
	assert.Equal(t, time.Second, c.retryWait)
	assert.Equal(t, 30*time.Second, c.retryMaxW)
	tr, ok := c.httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Equal(t, 100, tr.MaxIdleConns)
	assert.Equal(t, 10, tr.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, tr.IdleConnTimeout)
}

func TestNew_CustomPoolAndRetry(t *testing.T) {
	c := New(Options{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     60 * time.Second,
		RetryMax:            5,
		RetryWait:           2 * time.Second,
		RetryMaxWait:        60 * time.Second,
	})
	require.NotNil(t, c)
	assert.Equal(t, 5, c.retryMax)
	assert.Equal(t, 2*time.Second, c.retryWait)
	assert.Equal(t, 60*time.Second, c.retryMaxW)
	tr, ok := c.httpClient.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Equal(t, 50, tr.MaxIdleConns)
	assert.Equal(t, 5, tr.MaxIdleConnsPerHost)
	assert.Equal(t, 60*time.Second, tr.IdleConnTimeout)
}

func TestNew_ConfiguresHTTPClient(t *testing.T) {
	baseURL := "https://example.com"
	timeout := 5 * time.Second

	t.Run("with TLS verification", func(t *testing.T) {
		c := New(Options{
			BaseURL:   baseURL,
			Timeout:   timeout,
			VerifyTLS: true,
		})

		require.NotNil(t, c)
		require.NotNil(t, c.httpClient)
		assert.Equal(t, baseURL, c.baseURL)
		assert.Equal(t, timeout, c.httpClient.Timeout)

		// Transport and TLS settings
		tr, ok := c.httpClient.Transport.(*http.Transport)
		require.True(t, ok, "transport should be *http.Transport")
		require.NotNil(t, tr.TLSClientConfig)
		assert.False(t, tr.TLSClientConfig.InsecureSkipVerify)
		assert.Equal(t, 100, tr.MaxIdleConns)
		assert.Equal(t, 10, tr.MaxIdleConnsPerHost)
		assert.Equal(t, 90*time.Second, tr.IdleConnTimeout)
	})

	t.Run("without TLS verification", func(t *testing.T) {
		c := New(Options{
			BaseURL:   baseURL,
			Timeout:   timeout,
			VerifyTLS: false,
		})

		require.NotNil(t, c)
		tr, ok := c.httpClient.Transport.(*http.Transport)
		require.True(t, ok, "transport should be *http.Transport")
		require.NotNil(t, tr.TLSClientConfig)
		assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
	})
}

func TestDoJSON_Success(t *testing.T) {
	type requestBody struct {
		Name string `json:"name"`
	}
	type responseBody struct {
		Message string `json:"message"`
	}

	// Test server that echoes JSON
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "value", r.Header.Get("X-Test-Header"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		defer func() { _ = r.Body.Close() }()

		var rb requestBody
		require.NoError(t, json.Unmarshal(body, &rb))
		assert.Equal(t, "example", rb.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(responseBody{Message: "ok"})
	}))
	defer ts.Close()

	client := New(Options{
		BaseURL:   ts.URL,
		Timeout:   2 * time.Second,
		VerifyTLS: true,
	})

	var respBody responseBody
	resp, err := client.DoJSON(
		context.Background(),
		http.MethodPost,
		"/test",
		map[string]string{"X-Test-Header": "value"},
		requestBody{Name: "example"},
		&respBody,
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "ok", respBody.Message)
}

func TestDoJSON_ErrorStatus(t *testing.T) {
	// Server returns non-2xx
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer ts.Close()

	client := New(Options{
		BaseURL:   ts.URL,
		Timeout:   2 * time.Second,
		VerifyTLS: true,
	})

	var respBody struct{}
	resp, err := client.DoJSON(
		context.Background(),
		http.MethodGet,
		"/error",
		nil,
		nil,
		&respBody,
	)

	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestDoJSON_WithTracer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	tracer := noop.NewTracerProvider().Tracer("test")
	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true, Tracer: tracer})
	var out struct {
		OK bool `json:"ok"`
	}
	resp, err := client.DoJSON(context.Background(), http.MethodGet, "/", nil, nil, &out)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, out.OK)
}

func TestDoJSON_NoRespBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	resp, err := client.DoJSON(context.Background(), http.MethodGet, "/", nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGetRaw_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("raw body"))
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	body, resp, err := client.GetRaw(context.Background(), "/", map[string]string{"X-Custom": "v"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, []byte("raw body"), body)
	assert.Equal(t, "v", resp.Request.Header.Get("X-Custom"))
}

func TestGetRaw_Error(t *testing.T) {
	client := New(Options{BaseURL: "http://127.0.0.1:0", Timeout: time.Millisecond, VerifyTLS: true})
	_, resp, err := client.GetRaw(context.Background(), "/", nil)
	require.Error(t, err)
	assert.Nil(t, resp)
}

// TestGetRaw_BodyReadError uses a server that closes the connection after writing part of the body,
// so io.ReadAll gets an error (covers "failed to read response body" path).
func TestGetRaw_BodyReadError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		_, _ = w.Write([]byte("ab"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		if h, ok := w.(http.Hijacker); ok {
			conn, _, _ := h.Hijack()
			_ = conn.Close()
		}
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	body, resp, err := client.GetRaw(context.Background(), "/", nil)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Nil(t, body)
	assert.Contains(t, err.Error(), "read")
}

func TestGetRawToFile_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file content"))
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	path, resp, err := client.GetRawToFile(context.Background(), "/", nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer func() { _ = os.Remove(path) }()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, []byte("file content"), data)
}

func TestGetRawToFile_EmptyBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	path, resp, err := client.GetRawToFile(context.Background(), "/", nil)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "", path)
}

func TestGetRawToFile_DoReturnsError(t *testing.T) {
	// Unreachable URL so do() returns error (connection refused / timeout)
	client := New(Options{
		BaseURL:   "http://127.0.0.1:0",
		Timeout:   time.Millisecond,
		VerifyTLS: true,
	})
	path, resp, err := client.GetRawToFile(context.Background(), "/", nil)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "", path)
}

// TestGetRawToFile_BodyReadError uses a server that closes the connection mid-stream,
// so io.Copy gets an error (covers "failed to write response to file" path).
func TestGetRawToFile_BodyReadError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		_, _ = w.Write([]byte("ab"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		if h, ok := w.(http.Hijacker); ok {
			conn, _, _ := h.Hijack()
			_ = conn.Close()
		}
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	path, resp, err := client.GetRawToFile(context.Background(), "/", nil)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "", path)
	assert.Contains(t, err.Error(), "write")
}

func TestDo_WithTracer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tracer := noop.NewTracerProvider().Tracer("test")
	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true, Tracer: tracer})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/", nil)
	resp, err := client.Do(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()
}

func TestDo_WithoutTracer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/", nil)
	resp, err := client.Do(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()
}

func TestDo_RequestWithExistingContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL+"/", nil)
	resp, err := client.Do(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()
}

func TestDo_RelativeURLResolvedAgainstBaseURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/path", nil)
	resp, err := client.Do(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()
}

func TestDo_EmptyMethodDefaultsToGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	req, _ := http.NewRequestWithContext(context.Background(), "", ts.URL+"/", nil)
	req.Method = ""
	resp, err := client.Do(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()
}

func TestShouldRetry_NetOpError(t *testing.T) {
	opErr := &net.OpError{Op: "dial", Err: context.DeadlineExceeded}
	assert.True(t, shouldRetry(nil, opErr))
}

func TestShouldRetry_OtherError(t *testing.T) {
	assert.False(t, shouldRetry(nil, context.Canceled))
}

func TestShouldRetry_5xx(t *testing.T) {
	resp := &http.Response{StatusCode: 503}
	assert.True(t, shouldRetry(resp, nil))
}

func TestShouldRetry_2xx(t *testing.T) {
	resp := &http.Response{StatusCode: 200}
	assert.False(t, shouldRetry(resp, nil))
}

func TestDoWithRetry_ContextCanceledDuringBackoff(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := New(Options{
		BaseURL:      ts.URL,
		Timeout:      time.Second,
		VerifyTLS:    true,
		RetryMax:     3,
		RetryWait:    time.Hour,
		RetryMaxWait: time.Hour,
	})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	_, err := client.DoJSON(ctx, http.MethodGet, "/", nil, nil, nil)
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestDoWithRetry_RetryThenSuccess(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := New(Options{
		BaseURL:      ts.URL,
		Timeout:      time.Second,
		VerifyTLS:    true,
		RetryMax:     3,
		RetryWait:    1 * time.Millisecond,
		RetryMaxWait: 10 * time.Millisecond,
	})
	resp, err := client.DoJSON(context.Background(), http.MethodGet, "/", nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, attempts, 2)
}

func TestDoWithRetry_5xxMultipleRetriesThenSuccess(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := New(Options{
		BaseURL:      ts.URL,
		Timeout:      time.Second,
		VerifyTLS:    true,
		RetryMax:     5,
		RetryWait:    1 * time.Millisecond,
		RetryMaxWait: 50 * time.Millisecond,
	})
	resp, err := client.DoJSON(context.Background(), http.MethodGet, "/", nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, attempts, 4)
}

func TestDoJSON_MarshalError(t *testing.T) {
	client := New(Options{BaseURL: "http://example.com", Timeout: time.Second})
	_, err := client.DoJSON(context.Background(), http.MethodPost, "/", nil, make(chan int), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal")
}

func TestDoJSON_DecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer ts.Close()

	client := New(Options{BaseURL: ts.URL, Timeout: time.Second, VerifyTLS: true})
	var out struct{ Message string }
	resp, err := client.DoJSON(context.Background(), http.MethodGet, "/", nil, nil, &out)
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, err.Error(), "decode")
}

func TestDo_WithTracerAndError(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	client := New(Options{BaseURL: "http://127.0.0.1:0", Timeout: time.Millisecond, VerifyTLS: true, Tracer: tracer})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1:0/", nil)
	resp, err := client.Do(context.Background(), req)
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestDoJSON_WithTracerAndError(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	client := New(Options{BaseURL: "http://127.0.0.1:0", Timeout: time.Millisecond, VerifyTLS: true, Tracer: tracer})
	var out struct{}
	resp, err := client.DoJSON(context.Background(), http.MethodGet, "/", nil, nil, &out)
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestDoWithRetry_NonRetryableError(t *testing.T) {
	client := New(Options{BaseURL: "http://invalid-domain-that-does-not-resolve.example", Timeout: time.Millisecond, VerifyTLS: true, RetryMax: 2, RetryWait: time.Millisecond})
	var out struct{}
	_, err := client.DoJSON(context.Background(), http.MethodGet, "/", nil, nil, &out)
	require.Error(t, err)
}

func TestDoWithRetry_ConnectionErrorRetriesThenFails(t *testing.T) {
	client := New(Options{
		BaseURL:      "http://127.0.0.1:1",
		Timeout:      time.Millisecond,
		VerifyTLS:    true,
		RetryMax:     3,
		RetryWait:    1 * time.Millisecond,
		RetryMaxWait: 5 * time.Millisecond,
	})
	var out struct{}
	_, err := client.DoJSON(context.Background(), http.MethodGet, "/", nil, nil, &out)
	require.Error(t, err)
}
