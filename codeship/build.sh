#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Echo out all commands for monitoring progress
set -x

# Build all the things
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/alerts  cron/alerts/main.go
