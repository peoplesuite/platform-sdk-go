package providers

import "gopkg.in/yaml.v3"

func YAMLDecoder(data []byte, cfg any) error {
	return yaml.Unmarshal(data, cfg)
}
