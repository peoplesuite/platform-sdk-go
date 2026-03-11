package config

import "fmt"

// Versioned is implemented by config structs that declare a schema version.
type Versioned interface {
	ConfigVersion() int
}

// CurrentConfigVersion is the expected config schema version.
const CurrentConfigVersion = 1

// CheckVersion returns an error if cfg implements Versioned and its version does not match CurrentConfigVersion.
func CheckVersion(cfg any) error {

	v, ok := cfg.(Versioned)
	if !ok {
		return nil
	}

	if v.ConfigVersion() != CurrentConfigVersion {
		return fmt.Errorf(
			"config version mismatch: expected %d got %d",
			CurrentConfigVersion,
			v.ConfigVersion(),
		)
	}

	return nil
}
