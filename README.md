# platform-sdk-go

Core Go service framework for the PeopleSuite platform.

This repository provides the shared runtime and infrastructure required to build
microservices across the PeopleSuite ecosystem.

It standardizes:

- service startup
- HTTP servers
- gRPC servers
- configuration loading
- observability (logging, tracing, metrics)
- health checks
- graceful shutdown

The goal is that **every service boots and behaves the same way**.

---

# Package Overview

```
pkg/
config     – configuration loading and validation
runtime    – service lifecycle management
http       – HTTP server and middleware
grpc       – gRPC server helpers and interceptors
httpclient – HTTP client with retries and tracing
errors     – error kinds and mapping
observability – logging, tracing, metrics
health     – readiness and liveness checks
```

---

# Example

Minimal service:

```go
package main

import (
"context"
"net/http"

"google.golang.org/grpc"

"peoplesuite/platform-sdk-go/pkg/runtime"
)

func main() {

httpMux := http.NewServeMux()

grpcServer := grpc.NewServer()

rt, _ := runtime.New(runtime.Options{
ServiceName: "profile-svc",
Version: "1.0.0",
Environment: "production",

HTTPPort: 8080,
GRPCPort: 9090,

HTTPHandler: httpMux,
GRPCServer: grpcServer,
})

rt.Run(context.Background())
}
```

---

# Repository Structure

```

platform-sdk-go
│
├── pkg
│   ├── config
│   ├── runtime
│   ├── http
│   ├── grpc
│   ├── httpclient
│   ├── errors
│   ├── observability
│   └── health
│
├── examples
│   ├── shared
│   ├── config
│   ├── http-service
│   ├── grpc-service
│   └── worker
│
└── scripts

```

---

# Related Repositories

```

platform-contracts – protobuf / API schemas
platform-infra-go – infrastructure clients
platform-integrations – external integrations
platform-tools – developer tooling

```

---

# Development

| Command            | Description                    |
|--------------------|--------------------------------|
| `make lint`        | Run golangci-lint              |
| `make test`        | Run all tests                  |
| `make test-cover-pkg` | Test and coverage for `./pkg/...` |
| `make build`       | Build all packages             |
| `make tidy`        | Update dependencies (`go mod tidy`) |
| `make check`       | Format, vet, lint, and test     |

---

# License

Internal PeopleSuite platform library.