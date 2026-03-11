package shared

import "fmt"

// AppConfig is a sample config struct used by examples (config loading, services).
type AppConfig struct {
	Service string `yaml:"service" json:"service" toml:"service" envconfig:"SERVICE" doc:"Name of the service"`
	Port    int    `yaml:"port" json:"port" toml:"port" envconfig:"PORT" default:"8080" doc:"HTTP server port"`
	Debug   bool   `yaml:"debug" json:"debug" toml:"debug" envconfig:"DEBUG" default:"false" doc:"Enable debug logging"`
}

// ConfigVersion implements config versioning for the loader.
func (c *AppConfig) ConfigVersion() int { return 1 }

// ApplyDefaults sets default values for zero fields.
func (c *AppConfig) ApplyDefaults() {
	if c.Port == 0 {
		c.Port = 8080
	}
}

// Validate checks required fields.
func (c *AppConfig) Validate() error {
	if c.Service == "" {
		return fmt.Errorf("service must be defined")
	}
	return nil
}
