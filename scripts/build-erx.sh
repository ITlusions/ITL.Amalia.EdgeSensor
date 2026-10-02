#!/bin/sh
set -eu

mkdir -p dist
CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat \
  go build -trimpath -ldflags='-s -w' \
  -o dist/edge-sensor-linux-mipsle ./cmd/edge-sensor

echo 'Built dist/edge-sensor-linux-mipsle'