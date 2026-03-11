package config

import (
	"fmt"

	"github.com/peoplesuite/platform-sdk-go/pkg/config/providers"
)

// Loader loads configuration by running a sequence of providers.
type Loader struct {
	providers []providers.Provider
}

// NewLoader returns a Loader that uses the given providers in order.
func NewLoader(p ...providers.Provider) *Loader {
	return &Loader{
		providers: p,
	}
}

// Load runs each provider in order and populates cfg.
func (l *Loader) Load(cfg any) error {
	for _, p := range l.providers {
		if err := p.Load(cfg); err != nil {
			return fmt.Errorf("config provider %s failed: %w", p.Name(), err)
		}
	}

	if err := ApplyDefaults(cfg); err != nil {
		return err
	}

	if err := CheckVersion(cfg); err != nil {
		return err
	}

	if err := Validate(cfg); err != nil {
		return err
	}

	return nil
}
