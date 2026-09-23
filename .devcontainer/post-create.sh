#!/usr/bin/env bash
# Install checkout dependencies without starting application or host services.
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
(cd go && go mod download)
# PowerShell expands these variables, not Bash.
# shellcheck disable=SC2016
pwsh -NoLogo -NoProfile -Command '$ErrorActionPreference = "Stop"; Install-Module Pester -RequiredVersion 5.7.1 -Scope CurrentUser -Force; Install-Module PSScriptAnalyzer -RequiredVersion 1.24.0 -Scope CurrentUser -Force'
bash .devcontainer/smoke.sh
