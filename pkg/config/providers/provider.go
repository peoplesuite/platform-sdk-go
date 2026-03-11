package providers

// Provider loads configuration into a struct.
type Provider interface {
	Name() string
	Load(cfg any) error
}
