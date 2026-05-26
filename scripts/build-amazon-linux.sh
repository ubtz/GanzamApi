#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

mkdir -p dist

go test ./...

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

go build -trimpath -ldflags="-s -w" -o dist/GanzamApi .

echo "Built dist/GanzamApi for Amazon Linux amd64"
