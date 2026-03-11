package providers

import (
	"strings"
	"testing"
)

type envTestConfig struct {
	Name string `envconfig:"NAME"`
	Port int    `envconfig:"PORT"`
}

func TestEnvProvider_Load_Success(t *testing.T) {
	t.Setenv("APP_NAME", "service")
	t.Setenv("APP_PORT", "9000")

	var cfg envTestConfig
	p := NewEnv("APP")

	if err := p.Load(&cfg); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Name != "service" || cfg.Port != 9000 {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestEnvProvider_Name(t *testing.T) {
	p := NewEnv("APP")
	if got := p.Name(); got != "env" {
		t.Errorf("Name() = %q, want %q", got, "env")
	}
}

func TestEnvProvider_Load_Error(t *testing.T) {
	// Invalid int for PORT so envconfig.Process fails when parsing.
	t.Setenv("ENVLOADERR_PORT", "not-an-int")
	t.Setenv("ENVLOADERR_NAME", "ok")

	var cfg envTestConfig
	p := NewEnv("ENVLOADERR")

	err := p.Load(&cfg)
	if err == nil {
		t.Fatalf("expected error for invalid env, got nil")
	}
	if !strings.Contains(err.Error(), "env load error") {
		t.Logf("Load error (expected): %v", err)
	}
}
