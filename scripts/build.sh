#!/usr/bin/env bash

# Echo out all commands for monitoring progress
set -x

CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-s -w" -o bootstrap  cron/alerts/main.go
