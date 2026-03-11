package providers

import "testing"

type mergeConfig struct {
	Value string
}

type staticProvider struct {
	name  string
	value string
}

func (s *staticProvider) Name() string { return s.name }

func (s *staticProvider) Load(cfg any) error {
	c := cfg.(*mergeConfig)
	c.Value = s.value
	return nil
}

type errorProvider struct {
	name string
}

func (e *errorProvider) Name() string { return e.name }

func (e *errorProvider) Load(cfg any) error {
	return assertError // sentinel, defined below
}

var assertError = &testError{}

type testError struct{}

func (e *testError) Error() string { return "test error" }

func TestMergeProvider_Load_OrderAndOverride(t *testing.T) {
	cfg := &mergeConfig{}

	m := NewMerge(
		&staticProvider{name: "first", value: "one"},
		&staticProvider{name: "second", value: "two"},
	)

	if err := m.Load(cfg); err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Value != "two" {
		t.Fatalf("expected later provider to override value, got %q", cfg.Value)
	}
}

func TestMergeProvider_Load_PropagatesError(t *testing.T) {
	cfg := &mergeConfig{}

	m := NewMerge(
		&staticProvider{name: "ok", value: "one"},
		&errorProvider{name: "fail"},
	)

	if err := m.Load(cfg); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestMergeProvider_Name(t *testing.T) {
	m := NewMerge(&staticProvider{name: "a", value: "v"})
	if got := m.Name(); got != "merge" {
		t.Errorf("Name() = %q, want %q", got, "merge")
	}
}
