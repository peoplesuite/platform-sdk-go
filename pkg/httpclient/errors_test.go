package httpclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError_Error(t *testing.T) {
	err := NewHTTPError(404, []byte("not found"))
	assert.Equal(t, "http 404: not found", err.Error())
}

func TestHTTPError_EmptyBody(t *testing.T) {
	err := NewHTTPError(500, nil)
	assert.Equal(t, "http 500: ", err.Error())
}

func TestHTTPError_Fields(t *testing.T) {
	err := NewHTTPError(400, []byte("bad request"))
	assert.Equal(t, 400, err.StatusCode)
	assert.Equal(t, []byte("bad request"), err.Body)
}
