package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func DecodeJSONResponse(resp *http.Response, target any) error {
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return NewHTTPError(resp.StatusCode, body)
	}

	if target == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}
