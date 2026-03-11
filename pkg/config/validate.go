package config

import "fmt"

type Validator interface {
	Validate() error
}

func Validate(cfg any) error {

	if v, ok := cfg.(Validator); ok {

		if err := v.Validate(); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}

	}

	return nil
}
