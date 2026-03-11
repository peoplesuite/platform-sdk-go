package config

import (
	"testing"
)

type loaderConfig struct {
	Value          string
	DefaultsCalled bool
	Validated      bool
	Version        int
}

func (c *loaderConfig) ApplyDefaults() {
	c.DefaultsCalled = true
	if c.Value == "" {
		c.Value = "default"
	}
}

func (c *loaderConfig) Validate() error {
	c.Validated = true
	return nil
}

func (c *loaderConfig) ConfigVersion() int {
	return c.Version
}

type loaderProvider struct {
	name  string
	value string
	err   error
}

func (p *loaderProvider) Name() string { return p.name }

func (p *loaderProvider) Load(cfg any) error {
	if p.err != nil {
		return p.err
	}
	c := cfg.(*loaderConfig)
	c.Value = p.value
	return nil
}

func TestLoader_Load_SingleProviderAndPostProcessing(t *testing.T) {
	cfg := &loaderConfig{Version: CurrentConfigVersion}

	p := &loaderProvider{name: "p1", value: ""}
	loader := NewLoader(p)

	if err := loader.Load(cfg); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !cfg.DefaultsCalled {
		t.Fatalf("expected defaults to be applied")
	}
	if !cfg.Validated {
		t.Fatalf("expected validation to be called")
	}
	if cfg.Value == "" {
		t.Fatalf("expected value to be set by defaults")
	}
}

func TestLoader_Load_MultipleProviders_OrderAndOverride(t *testing.T) {
	cfg := &loaderConfig{Version: CurrentConfigVersion}

	p1 := &loaderProvider{name: "first", value: "one"}
	p2 := &loaderProvider{name: "second", value: "two"}

	loader := NewLoader(p1, p2)

	if err := loader.Load(cfg); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Value != "two" {
		t.Fatalf("expected later provider to override value, got %q", cfg.Value)
	}
}

func TestLoader_Load_ProviderErrorStopsPipeline(t *testing.T) {
	cfg := &loaderConfig{Version: CurrentConfigVersion}

	p1 := &loaderProvider{name: "ok", value: "one"}
	p2 := &loaderProvider{name: "fail", err: assertLoaderError{}}

	loader := NewLoader(p1, p2)

	err := loader.Load(cfg)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// Defaults should not run when provider fails.
	if cfg.DefaultsCalled {
		t.Fatalf("expected defaults not to be applied on provider error")
	}
}

type assertLoaderError struct{}

func (e assertLoaderError) Error() string { return "provider error" }

func TestLoader_Load_VersionMismatch(t *testing.T) {
	cfg := &loaderConfig{Version: CurrentConfigVersion + 1}

	p := &loaderProvider{name: "p1", value: "value"}
	loader := NewLoader(p)

	if err := loader.Load(cfg); err == nil {
		t.Fatalf("expected version mismatch error, got nil")
	}
}

// validateFailConfig implements Validator and always returns an error (for Load error path coverage).
type validateFailConfig struct {
	loaderConfig
}

func (c *validateFailConfig) Validate() error {
	c.Validated = true
	return assertLoaderError{}
}

// noopProvider loads nothing and works with any config (for ValidateError test).
type noopProvider struct{ name string }

func (p *noopProvider) Name() string     { return p.name }
func (p *noopProvider) Load(_ any) error { return nil }

func TestLoader_Load_ValidateError(t *testing.T) {
	cfg := &validateFailConfig{}
	cfg.loaderConfig.Version = CurrentConfigVersion

	loader := NewLoader(&noopProvider{name: "noop"})

	err := loader.Load(cfg)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	if !cfg.Validated {
		t.Fatalf("expected Validate to be called")
	}
}
