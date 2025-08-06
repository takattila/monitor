#!/usr/bin/env bash
set -e

# List all packages
all_packages=$(go list ./...)

# Filter: only include packages that do NOT contain 'package main'
filtered_packages=()
for pkg in $all_packages; do
  # Get the directory of the package
  pkg_dir=$(go list -f '{{.Dir}}' "$pkg")

  # If no 'package main' is found in the files, include the package
  if ! grep -q '^package main' "$pkg_dir"/*.go 2>/dev/null; then
    filtered_packages+=("$pkg")
  fi
done

# Run tests on the filtered packages
go test -coverprofile="coverage.out" -covermode=atomic -v "${filtered_packages[@]}"

# Generate coverage report
go tool cover -func=coverage.out

