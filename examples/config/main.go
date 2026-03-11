package main

import (
	"log"

	"peoplesuite/platform-sdk-go/examples/shared"
	"peoplesuite/platform-sdk-go/pkg/config"
	"peoplesuite/platform-sdk-go/pkg/config/providers"
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
