package config

import (
	"errors"
	"strings"
	"testing"
)

type validConfig struct{}

func (v *validConfig) Validate() error {
	return nil
}

type invalidConfig struct {
	Err error
}

func (v *invalidConfig) Validate() error {
	return v.Err
}

func TestValidate_ImplementsInterface_NoError(t *testing.T) {
	cfg := &validConfig{}

	if err := Validate(cfg); err != nil {
		t.Fatalf("Validate returned unexpected error: %v", err)
	}
}

func TestValidate_ImplementsInterface_WithErrorWrapped(t *testing.T) {
	baseErr := errors.New("boom")
	cfg := &invalidConfig{Err: baseErr}

	err := Validate(cfg)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "config validation failed") {
		t.Fatalf("expected wrapped error message, got %q", err.Error())
	}
	if !errors.Is(err, baseErr) {
		t.Fatalf("expected error to wrap base error")
	}
}

func TestValidate_NoInterface_NoError(t *testing.T) {
	cfg := struct{ Field string }{Field: "value"}

	if err := Validate(&cfg); err != nil {
		t.Fatalf("Validate returned unexpected error: %v", err)
	}
}
