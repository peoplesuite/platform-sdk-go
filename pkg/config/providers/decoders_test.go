package providers

import "testing"

type decoderConfig struct {
	Value string `yaml:"value" json:"value" toml:"value"`
}

func TestYAMLDecoder(t *testing.T) {
	var cfg decoderConfig
	data := []byte("value: test")

	if err := YAMLDecoder(data, &cfg); err != nil {
		t.Fatalf("YAMLDecoder error: %v", err)
	}
	if cfg.Value != "test" {
		t.Fatalf("unexpected value: %q", cfg.Value)
	}
}

func TestJSONDecoder(t *testing.T) {
	var cfg decoderConfig
	data := []byte(`{"value":"test"}`)

	if err := JSONDecoder(data, &cfg); err != nil {
		t.Fatalf("JSONDecoder error: %v", err)
	}
	if cfg.Value != "test" {
		t.Fatalf("unexpected value: %q", cfg.Value)
	}
}

func TestTOMLDecoder(t *testing.T) {
	var cfg decoderConfig
	data := []byte("value = \"test\"")

	if err := TOMLDecoder(data, &cfg); err != nil {
		t.Fatalf("TOMLDecoder error: %v", err)
	}
	if cfg.Value != "test" {
		t.Fatalf("unexpected value: %q", cfg.Value)
	}
}
