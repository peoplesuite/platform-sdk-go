package providers

import "gopkg.in/yaml.v3"

// YAMLDecoder decodes YAML bytes into cfg.
func YAMLDecoder(data []byte, cfg any) error {
	return yaml.Unmarshal(data, cfg)
}
