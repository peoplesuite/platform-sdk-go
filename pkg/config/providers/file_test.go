package providers

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fileTestConfig struct {
	Value string `yaml:"value" json:"value" toml:"value"`
}

func TestFileProvider_Load_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.txt")
	data := []byte("dummy")

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	var decoded []byte
	decoder := func(d []byte, cfg any) error {
		decoded = append(decoded, d...)
		return nil
	}

	p := NewFile(path, decoder)

	var cfg fileTestConfig
	if err := p.Load(&cfg); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if string(decoded) != string(data) {
		t.Fatalf("decoder did not receive expected data; got %q want %q", decoded, data)
	}
}

func TestFileProvider_Load_FileMissing(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "nonexistent-file")
	p := NewFile(missingPath, func([]byte, any) error { return nil })

	var cfg fileTestConfig
	if err := p.Load(&cfg); err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

func TestFileProvider_Name(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cfg.yaml")
	p := NewFile(path, func([]byte, any) error { return nil })
	want := "file:" + path
	if got := p.Name(); got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

func TestFileProvider_Load_DecodeError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.txt")

	if err := os.WriteFile(path, []byte("ignored"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	decodeErr := errors.New("decode failed")
	decoder := func([]byte, any) error {
		return decodeErr
	}

	p := NewFile(path, decoder)

	var cfg fileTestConfig
	if err := p.Load(&cfg); err == nil {
		t.Fatalf("expected error, got nil")
	}
}
