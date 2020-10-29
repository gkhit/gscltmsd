#!/bin/sh

set -e
export GOFLAGS="-mod=vendor"

rm -rf bin
mkdir bin
env -w GOPRIVATE=github.com/gkhit
go build -ldflags "-s -w" -o bin/gscltmsd main.go
