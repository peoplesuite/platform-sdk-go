package providers

import "github.com/BurntSushi/toml"

// TOMLDecoder decodes TOML bytes into cfg.
func TOMLDecoder(data []byte, cfg any) error {
	return toml.Unmarshal(data, cfg)
}
