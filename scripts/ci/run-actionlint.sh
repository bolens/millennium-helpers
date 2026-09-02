#!/usr/bin/env bash
# Run actionlint on workflow files. Skips cleanly when actionlint is not installed.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

if ! command -v actionlint >/dev/null 2>&1; then
  echo "skip: actionlint not installed (see CONTRIBUTING.md)"
  exit 0
fi

shopt -s nullglob
workflows=(.github/workflows/*.yml .github/workflows/*.yaml)
if [[ ${#workflows[@]} -eq 0 ]]; then
  echo "no workflow files found"
  exit 0
fi

actionlint "${workflows[@]}"
if command -v zizmor >/dev/null 2>&1; then
  zizmor --offline --min-severity medium --min-confidence medium .github
else
  echo "skip: zizmor not installed (see CONTRIBUTING.md)"
fi
