package providers

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type EnvProvider struct {
	Prefix string
}

func NewEnv(prefix string) *EnvProvider {
	return &EnvProvider{
		Prefix: prefix,
	}
}

func (p *EnvProvider) Name() string {
	return "env"
}

func (p *EnvProvider) Load(cfg any) error {

	if err := envconfig.Process(p.Prefix, cfg); err != nil {
		return fmt.Errorf("env load error: %w", err)
	}

	return nil
}
