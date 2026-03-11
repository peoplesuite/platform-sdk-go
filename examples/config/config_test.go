package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/peoplesuite/platform-sdk-go/examples/shared"
	"github.com/peoplesuite/platform-sdk-go/pkg/config"
	"github.com/peoplesuite/platform-sdk-go/pkg/config/providers"
)

func TestConfigLoad_LoadsFromFilesAndEnv(t *testing.T) {
	dir := t.TempDir()

	copyFile := func(name string) string {
		src := name
		dst := filepath.Join(dir, name)
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		if err := os.WriteFile(dst, data, 0o600); err != nil {
			t.Fatalf("write %s: %v", dst, err)
		}
		return dst
	}

	yamlPath := copyFile("config.yaml")
	tomlPath := copyFile("config.toml")
	jsonPath := copyFile("config.json")

	t.Setenv("APP_SERVICE", "env-service")

	var cfg shared.AppConfig

	loader := config.NewLoader(
		providers.NewMerge(
			providers.NewFile(yamlPath, providers.YAMLDecoder),
			providers.NewFile(tomlPath, providers.TOMLDecoder),
			providers.NewFile(jsonPath, providers.JSONDecoder),
		),
		providers.NewEnv("APP"),
	)

	if err := loader.Load(&cfg); err != nil {
		t.Fatalf("loader.Load error: %v", err)
	}

	if cfg.Service != "env-service" {
		t.Fatalf("expected service from env, got %q", cfg.Service)
	}
	if cfg.Port == 0 {
		t.Fatalf("expected non-zero Port after defaults/file config")
	}
}
