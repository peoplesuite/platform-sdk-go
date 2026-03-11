package config

type DefaultsApplier interface {
	ApplyDefaults()
}

func ApplyDefaults(cfg any) error {

	if d, ok := cfg.(DefaultsApplier); ok {
		d.ApplyDefaults()
	}

	return nil
}
