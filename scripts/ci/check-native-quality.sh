#!/usr/bin/env bash
# Portable equivalents of the migrated hosted quality and workflow checks.
set -euo pipefail
root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)
cd -- "$root"
scratch=$(mktemp -d "${TMPDIR:-/tmp}/millennium-quality.XXXXXX")
trap 'rm -rf -- "$scratch"' EXIT
export GOBIN="$scratch/bin"
export PATH="$GOBIN:$PATH"
mkdir -p "$GOBIN"
# Go module downloads use the checksum database. Keep versions aligned with CI policy.
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12
python3 -m venv "$scratch/venv"
"$scratch/venv/bin/pip" install --disable-pip-version-check zizmor==1.29.0
(cd go && golangci-lint run --config=../.golangci.yml)
actionlint
shellcheck .devcontainer/*.sh
"$scratch/venv/bin/zizmor" --offline --min-severity medium --min-confidence medium .github/workflows
python3 -m compileall -q scripts
ruff check scripts
make check-packaging check-winget
# The remaining hosted Go matrix still runs native Linux/Windows/macOS tests.
# Compile Windows packages here as well, matching the migrated quality job.
(cd go
packages=$(go list -buildvcs=false ./...)
for pkg in $packages; do
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go test -buildvcs=false -c -o /dev/null "$pkg"
done)
