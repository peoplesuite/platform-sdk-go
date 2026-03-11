package config

import "fmt"

// Validator is implemented by config structs that can validate themselves.
type Validator interface {
	Validate() error
}

// Validate calls Validate on cfg if it implements Validator.
func Validate(cfg any) error {

	if v, ok := cfg.(Validator); ok {

		if err := v.Validate(); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

	}

	return nil
}
