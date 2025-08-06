#!/bin/bash
set -e

# Csak nem-main csomagok listázása
filtered_packages=$(go list -f '{{.Name}} {{.ImportPath}}' ./... | awk '$1 != "main" {print $2}')

# Tesztek futtatása párhuzamosan
go test -coverprofile="coverage.out" -covermode=atomic -v -p 4 $filtered_packages

# Lefedettségi jelentés
go tool cover -func=coverage.out

