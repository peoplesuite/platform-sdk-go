package config

import "testing"

type testDefaultsConfig struct {
	Applied bool
}

func (c *testDefaultsConfig) ApplyDefaults() {
	c.Applied = true
}

func TestApplyDefaults_ImplementsInterface(t *testing.T) {
	cfg := &testDefaultsConfig{}

	if err := ApplyDefaults(cfg); err != nil {
		t.Fatalf("ApplyDefaults returned error: %v", err)
	}

	if !cfg.Applied {
		t.Fatalf("expected ApplyDefaults to set Applied=true")
	}
}

func TestApplyDefaults_NoInterface(t *testing.T) {
	// A type that does not implement DefaultsApplier should be a no-op.
	cfg := struct {
		Value int
	}{Value: 1}

	if err := ApplyDefaults(&cfg); err != nil {
		t.Fatalf("ApplyDefaults returned error: %v", err)
	}

	if cfg.Value != 1 {
		t.Fatalf("ApplyDefaults unexpectedly modified cfg: %+v", cfg)
	}
}
