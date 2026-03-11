package config

// DefaultsApplier is implemented by config structs that can apply default values.
type DefaultsApplier interface {
	ApplyDefaults()
}

// ApplyDefaults calls ApplyDefaults on cfg if it implements DefaultsApplier.
func ApplyDefaults(cfg any) error {

	if d, ok := cfg.(DefaultsApplier); ok {
		d.ApplyDefaults()
	}

	return nil
}
