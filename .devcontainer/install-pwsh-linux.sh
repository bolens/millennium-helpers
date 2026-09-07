#!/usr/bin/env bash
# Install the CI-pinned PowerShell runtime after verifying its official checksum.
set -euo pipefail
version="${1:-7.4.7}"
[[ "$version" == 7.4.7 ]] || { echo 'Update the version and reviewed checksums together.' >&2; exit 2; }
case "$(uname -m)" in
  x86_64) arch=x64; checksum=cfd57927b7a0d9f2da400471ea8dadc3ceb52f5e4a4a9a51c20495b4071f055a ;;
  aarch64|arm64) arch=arm64; checksum=e5d34d4777c4d8841eade59dfdb6a8ceb0bb64b690863e4bf976eb698021f446 ;;
  *) echo 'PowerShell container supports x86-64 and ARM64.' >&2; exit 2 ;;
esac
archive="$(mktemp)"
trap 'rm -f "$archive"' EXIT
curl --fail --show-error --location --retry 3 --output "$archive" \
  "https://github.com/PowerShell/PowerShell/releases/download/v${version}/powershell-${version}-linux-${arch}.tar.gz"
printf '%s  %s\n' "$checksum" "$archive" | sha256sum --check -
install -d /opt/microsoft/powershell/7
tar -xzf "$archive" -C /opt/microsoft/powershell/7
chmod a+x /opt/microsoft/powershell/7/pwsh
ln -sfn /opt/microsoft/powershell/7/pwsh /usr/local/bin/pwsh
# PowerShell expands these variables, not Bash.
# shellcheck disable=SC2016
pwsh -NoLogo -NoProfile -Command '$PSVersionTable.PSVersion.ToString()'
