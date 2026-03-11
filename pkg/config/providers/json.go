package providers

import "encoding/json"

func JSONDecoder(data []byte, cfg any) error {
	return json.Unmarshal(data, cfg)
}
