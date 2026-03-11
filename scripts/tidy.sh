#!/bin/bash

set -e

echo "Running go mod tidy..."
go mod tidy

echo "Formatting code..."
go fmt ./...

echo "Done."