package providers

import "fmt"

// MergeProvider executes multiple providers sequentially.
// Later providers override values loaded by earlier providers.
//
// Typical usage:
//
//	base.yaml
//	production.yaml
//	local.yaml
//	env variables
//
// Example:
//
//	providers.NewMerge(
//	    providers.NewFile("base.yaml", providers.YAMLDecoder),
//	    providers.NewFile("prod.yaml", providers.YAMLDecoder),
//	    providers.NewFile("local.yaml", providers.YAMLDecoder),
//	)
type MergeProvider struct {
	providers []Provider
}

// NewMerge creates a provider that loads several providers sequentially.
func NewMerge(p ...Provider) *MergeProvider {
	return &MergeProvider{
		providers: p,
	}
}

// Name returns provider name
func (m *MergeProvider) Name() string {
	return "merge"
}

// Load executes all providers sequentially.
// Later providers override earlier values.
func (m *MergeProvider) Load(cfg any) error {

	for _, p := range m.providers {

		if err := p.Load(cfg); err != nil {
			return fmt.Errorf(
				"merge provider: %s failed: %w",
				p.Name(),
				err,
			)
		}

	}

	return nil
}
