.PHONY: tidy lint test test-cover test-cover-pkg vet clean check build format help

help:
	@echo "Targets:"
	@echo "  tidy            - go mod tidy"
	@echo "  format          - go fmt ./..."
	@echo "  vet             - go vet ./..."
	@echo "  lint            - golangci-lint run"
	@echo "  test            - go test ./..."
	@echo "  test-cover      - go test ./pkg/... -cover -coverprofile=coverage.out (examples excluded)"
	@echo "  test-cover-pkg  - same as test-cover (90%% coverage target applies to pkg/ only)"
	@echo "  build           - go build ./..."
	@echo "  clean           - remove coverage and build artifacts"
	@echo "  check           - format, vet, lint, test"

tidy:
	go mod tidy

format:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./...

# Coverage for pkg/ only (examples excluded; 90%% target applies to pkg/)
test-cover test-cover-pkg:
	go test ./pkg/... -cover -coverprofile=coverage.out

build:
	go build ./...

# Coverage and test output artifacts to remove on clean
CLEAN_ARTIFACTS := coverage.out cov cov_http cov_http.out cov_httpclient cov.out cov2.out coverage coverage_errors coverage_providers e.out full.out full2.out httpclient_cov.out

# Remove coverage and build artifacts; use OS-specific commands for Windows and Unix
clean:
ifeq ($(OS),Windows_NT)
	-del /q /f $(CLEAN_ARTIFACTS) 2>nul
else
	-rm -f $(CLEAN_ARTIFACTS)
endif
	go clean -testcache

check: format vet lint test