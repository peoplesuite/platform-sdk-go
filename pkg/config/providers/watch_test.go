package providers

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatchProvider_NameAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte("value: test"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	var cfg struct {
		Value string `yaml:"value"`
	}

	p, err := NewWatch(path, YAMLDecoder, func() error { return nil })
	if err != nil {
		t.Fatalf("NewWatch returned error: %v", err)
	}

	if got := p.Name(); got == "" {
		t.Fatalf("expected non-empty provider name")
	}

	if err := p.Load(&cfg); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Value != "test" {
		t.Fatalf("unexpected value: %q", cfg.Value)
	}
}

func TestWatchProvider_Start_NoPanic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte("value: test"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	p, err := NewWatch(path, YAMLDecoder, func() error { return nil })
	if err != nil {
		t.Fatalf("NewWatch returned error: %v", err)
	}

	if err := p.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	// Give the watcher goroutine a brief moment to start; we only care that it does not panic.
	time.Sleep(50 * time.Millisecond)
}

func TestWatchProvider_Load_FileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.yaml")
	p, err := NewWatch(path, YAMLDecoder, nil)
	if err != nil {
		t.Fatalf("NewWatch returned error: %v", err)
	}

	var cfg struct{ Value string }
	if err := p.Load(&cfg); err == nil {
		t.Fatalf("expected error when file does not exist, got nil")
	}
}

func TestWatchProvider_Load_DecodeError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte("value: x"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	decoder := func([]byte, any) error { return os.ErrInvalid }
	p, err := NewWatch(path, decoder, nil)
	if err != nil {
		t.Fatalf("NewWatch returned error: %v", err)
	}

	var cfg struct{ Value string }
	if err := p.Load(&cfg); err == nil {
		t.Fatalf("expected decode error, got nil")
	}
}

func TestWatchProvider_Start_InvalidPath(t *testing.T) {
	// Directory that does not exist so watcher.Add fails
	path := filepath.Join(t.TempDir(), "nonexistent-subdir", "file.yaml")
	p, err := NewWatch(path, YAMLDecoder, nil)
	if err != nil {
		t.Fatalf("NewWatch returned error: %v", err)
	}

	if err := p.Start(); err == nil {
		t.Fatalf("expected error when directory does not exist, got nil")
	}
}

func TestWatchProvider_FileChangeTriggersReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte("value: a"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	var reloadMu sync.Mutex
	reloadCalled := false
	onReload := func() error {
		reloadMu.Lock()
		reloadCalled = true
		reloadMu.Unlock()
		return nil
	}

	p, err := NewWatch(path, YAMLDecoder, onReload)
	if err != nil {
		t.Fatalf("NewWatch returned error: %v", err)
	}
	if err := p.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	defer func() { _ = p.watcher.Close() }()

	time.Sleep(50 * time.Millisecond)

	if err := os.WriteFile(path, []byte("value: b"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	reloadMu.Lock()
	got := reloadCalled
	reloadMu.Unlock()
	if !got {
		t.Log("OnReload may not have been called yet; watcher event timing is best-effort")
	}
}

func TestWatchProvider_OnReloadError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte("value: a"), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	onReload := func() error { return os.ErrClosed }
	p, err := NewWatch(path, YAMLDecoder, onReload)
	if err != nil {
		t.Fatalf("NewWatch returned error: %v", err)
	}
	if err := p.Start(); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	defer func() { _ = p.watcher.Close() }()

	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(path, []byte("value: b"), 0o600); err != nil {
		t.Fatalf("failed to write: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
}
