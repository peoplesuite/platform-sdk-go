package providers

import (
	"fmt"
	"os"
)

type Decoder func(data []byte, cfg any) error

type FileProvider struct {
	Path    string
	Decoder Decoder
}

func NewFile(path string, decoder Decoder) *FileProvider {
	return &FileProvider{
		Path:    path,
		Decoder: decoder,
	}
}

func (p *FileProvider) Name() string {
	return "file:" + p.Path
}

func (p *FileProvider) Load(cfg any) error {

	data, err := os.ReadFile(p.Path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}

	if err := p.Decoder(data, cfg); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	return nil
}

func readFile(path string) ([]byte, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	return data, nil
}
