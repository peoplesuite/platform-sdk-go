package providers

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// EnvProvider loads configuration from environment variables.
type EnvProvider struct {
	Prefix string
}

// NewEnv returns an EnvProvider that reads vars with the given prefix.
func NewEnv(prefix string) *EnvProvider {
	return &EnvProvider{
		Prefix: prefix,
	}
}

// Name returns the provider name.
func (p *EnvProvider) Name() string {
	return "env"
}

// Load fills cfg from environment variables using envconfig.
func (p *EnvProvider) Load(cfg any) error {

	if err := envconfig.Process(p.Prefix, cfg); err != nil {
		return fmt.Errorf("env load error: %w", err)
	}

	return nil
}
