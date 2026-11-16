#!/usr/bin/env bash
# Preserve the former hosted Nix evaluation and package build contract.
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.."
export NIX_CONFIG="${NIX_CONFIG:-}
experimental-features = nix-command flakes"
nix flake check --all-systems --no-build --no-write-lock-file
nix build --no-link .#millennium-helpers-git
release_version=$(tr -d '[:space:]' < VERSION)
for variant in source binary; do
  if [[ "$variant" == source ]]; then
    asset="millennium-helpers-v${release_version}-src.tar.gz"
    output=millennium-helpers
  else
    asset="millennium-helpers-v${release_version}-linux-amd64.tar.gz"
    output=millennium-helpers-bin
  fi
  if curl -fsI "https://github.com/bolens/millennium-helpers/releases/download/v${release_version}/${asset}" >/dev/null; then
    nix build --no-link ".#$output"
  else
    echo "Release asset unavailable; evaluate $output without building: $asset"
    nix eval ".#$output.version"
  fi
done
