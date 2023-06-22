#!/usr/bin/env bash
set -e

go test -coverprofile="coverage.out" -covermode="atomic" -v ./...
go tool cover -func coverage.out 