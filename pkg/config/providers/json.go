package providers

import "encoding/json"

// JSONDecoder decodes JSON bytes into cfg.
func JSONDecoder(data []byte, cfg any) error {
	return json.Unmarshal(data, cfg)
}
