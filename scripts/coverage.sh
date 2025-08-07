#!/bin/bash
set -e

# List packages excluding the main package
filtered_packages=$(go list -f '{{.Name}} {{.ImportPath}}' ./... | awk '$1 != "main" {print $2}')

# Run tests in parallel with coverage, saving report to coverage.out
go test -coverprofile="coverage.out" -covermode=atomic -v -p 4 $filtered_packages

# Show coverage report summary
go tool cover -func=coverage.out
