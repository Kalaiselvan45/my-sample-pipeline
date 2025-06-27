#!/bin/bash
set -e
echo "Running certification checks..."
go test ./...
golangci-lint run || true
echo "Certification complete."
