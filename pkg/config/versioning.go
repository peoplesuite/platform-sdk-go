package config

import "fmt"

type Versioned interface {
	ConfigVersion() int
}

const CurrentConfigVersion = 1

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
