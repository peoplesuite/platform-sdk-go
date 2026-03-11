package config

import (
	"fmt"

	"peoplesuite/platform-sdk-go/pkg/config/providers"
)

type Loader struct {
	providers []providers.Provider
}

func NewLoader(p ...providers.Provider) *Loader {
	return &Loader{
		providers: p,
	}
}

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
