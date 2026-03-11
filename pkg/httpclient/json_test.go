package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeJSONResponse_Success(t *testing.T) {
	type payload struct {
		Value string `json:"value"`
	}

	// Create a fake response
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(rec).Encode(payload{Value: "test"})

	resp := rec.Result()
	// No defer resp.Body.Close() - DecodeJSONResponse handles it

	var out payload
	err := DecodeJSONResponse(resp, &out)
	require.NoError(t, err)
	assert.Equal(t, "test", out.Value)
}

func TestDecodeJSONResponse_NilTarget(t *testing.T) {
	// Ensure body is drained and no error when target is nil
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)
	_, _ = rec.Write([]byte(`{"ignored":"value"}`))

	resp := rec.Result()
	// No defer resp.Body.Close() - DecodeJSONResponse handles it

	err := DecodeJSONResponse(resp, nil)
	require.NoError(t, err)

	// Body should be drained; second read should be EOF
	_, readErr := io.ReadAll(resp.Body)
	assert.NoError(t, readErr)
}

func TestDecodeJSONResponse_InvalidJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)
	_, _ = rec.Write([]byte(`{invalid-json`))

	resp := rec.Result()
	// No defer resp.Body.Close() - DecodeJSONResponse handles it

	var out struct{}
	err := DecodeJSONResponse(resp, &out)
	require.Error(t, err)
}

func TestDecodeJSONResponse_HTTPErrorStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusInternalServerError)
	_, _ = rec.Write([]byte("server error"))

	resp := rec.Result()
	// No defer resp.Body.Close() - DecodeJSONResponse handles it

	var out struct{}
	err := DecodeJSONResponse(resp, &out)
	require.Error(t, err)

	// Verify it's an HTTPError
	httpErr, ok := err.(*HTTPError)
	require.True(t, ok, "error should be *HTTPError")
	assert.Equal(t, http.StatusInternalServerError, httpErr.StatusCode)
	assert.Equal(t, "server error", string(httpErr.Body))
}
