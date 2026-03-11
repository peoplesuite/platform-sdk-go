package main

import (
	"log"

	"github.com/peoplesuite/platform-sdk-go/examples/shared"
	"github.com/peoplesuite/platform-sdk-go/pkg/config"
	"github.com/peoplesuite/platform-sdk-go/pkg/config/providers"
)

func main() {
	var cfg shared.AppConfig

	loader := config.NewLoader(
		providers.NewMerge(
			providers.NewFile("examples/config/config.yaml", providers.YAMLDecoder),
			providers.NewFile("examples/config/config.toml", providers.TOMLDecoder),
			providers.NewFile("examples/config/config.json", providers.JSONDecoder),
		),
		providers.NewEnv("APP"),
	)

	if err := loader.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	log.Printf("loaded config: %+v\n", cfg)
}
