package config

import (
	"strings"
	"testing"
)

type versionedConfig struct {
	Version int
}

func (v *versionedConfig) ConfigVersion() int {
	return v.Version
}

func TestCheckVersion_NotVersioned_NoError(t *testing.T) {
	cfg := struct{ Field string }{Field: "x"}

	if err := CheckVersion(&cfg); err != nil {
		t.Fatalf("CheckVersion returned unexpected error: %v", err)
	}
}

func TestCheckVersion_MatchingVersion(t *testing.T) {
	cfg := &versionedConfig{Version: CurrentConfigVersion}

	if err := CheckVersion(cfg); err != nil {
		t.Fatalf("CheckVersion returned unexpected error: %v", err)
	}
}

func TestCheckVersion_MismatchedVersion_Error(t *testing.T) {
	cfg := &versionedConfig{Version: CurrentConfigVersion + 1}

	err := CheckVersion(cfg)
	if err == nil {
		t.Fatalf("expected error for mismatched version, got nil")
	}
	if !strings.Contains(err.Error(), "config version mismatch") {
		t.Fatalf("expected mismatch message, got %q", err.Error())
	}
}
