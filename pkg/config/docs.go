package config

/*
Package config provides a flexible configuration loading system.

Features:

• Multiple providers (file, env, etc.)
• Multiple formats (yaml, json, toml)
• Default value hooks
• Validation hooks
• Version checking
• Ordered provider overrides

Example:

	cfg := MyConfig{}

	loader := config.NewLoader(
		providers.NewFile("config.yaml", providers.YAMLDecoder),
		providers.NewEnv("APP"),
	)

	if err := loader.Load(&cfg); err != nil {
		panic(err)
	}
*/
