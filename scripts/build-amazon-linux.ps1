$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$dist = Join-Path $root "dist"
New-Item -ItemType Directory -Force -Path $dist | Out-Null

go test ./...
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

go build -trimpath -ldflags="-s -w" -o (Join-Path $dist "GanzamApi") .
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "Built dist/GanzamApi for Amazon Linux amd64"
