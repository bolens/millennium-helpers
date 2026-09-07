#!/usr/bin/env bash
# Install the Nushell version used by the completion CI job.
set -euo pipefail
case "$(uname -m)" in
  x86_64) target=x86_64-unknown-linux-gnu; checksum=3896777ebf3678f6d41736a5e995ba8360b338eb73c713254fc024e08ec72289 ;;
  aarch64|arm64) target=aarch64-unknown-linux-gnu; checksum=280163b9a2b3c54e45ded348b577a569cdf2a292e8fe9d77fe02405c3415c8c3 ;;
  *) echo 'Nushell container supports x86-64 and ARM64.' >&2; exit 2 ;;
esac
archive="$(mktemp)"
trap 'rm -f "$archive"' EXIT
curl --fail --show-error --location --retry 3 --output "$archive" \
  "https://github.com/nushell/nushell/releases/download/0.114.0/nu-0.114.0-${target}.tar.gz"
printf '%s  %s\n' "$checksum" "$archive" | sha256sum --check -
tar -xzf "$archive" -C /usr/local/bin --strip-components=1 "nu-0.114.0-${target}/nu"
nu --version
