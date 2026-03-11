# Examples

Runnable examples for the platform-sdk-go packages. Run from the repository root.

| Example | Description | Packages |
|---------|-------------|----------|
| **config** | Load configuration from YAML/TOML/JSON files and environment variables. | `pkg/config`, `pkg/config/providers` |
| **http-service** | Minimal HTTP service with runtime, middleware (RequestID, Logging, Recovery), and health. | `pkg/runtime`, `pkg/http`, `pkg/health`, `pkg/observability` |
| **grpc-service** | Minimal gRPC service with runtime, interceptors, health, and reflection. | `pkg/runtime`, `pkg/grpc`, `pkg/health`, `pkg/observability` |
| **worker** | Runtime with background workers only (no HTTP/gRPC servers). | `pkg/runtime` |

## Running

```bash
# Config loading (run from repo root; config files in examples/config/)
go run ./examples/config

# HTTP service (listens on :8080, health on :8081)
go run ./examples/http-service

# gRPC service (listens on :9090)
go run ./examples/grpc-service

# Worker (runs until SIGINT/SIGTERM)
go run ./examples/worker
```

## Layout

- **shared/** – Shared types (e.g. `AppConfig`) used by the config example and optionally by services.
- **config/** – Config loader example; includes `config.yaml`, `config.toml`, `config.json` in this directory.
