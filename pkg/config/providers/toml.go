package providers

import "github.com/BurntSushi/toml"

func TOMLDecoder(data []byte, cfg any) error {
	return toml.Unmarshal(data, cfg)
}
