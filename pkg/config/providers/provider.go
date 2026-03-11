package providers

type Provider interface {
	Name() string
	Load(cfg any) error
}
