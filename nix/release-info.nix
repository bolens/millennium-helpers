# Pinned release metadata for Nix packages.
# Updated by scripts/ci/update-packaging-versions.sh on each release.
{
  version = "4.1.2";
  # SRI hash of millennium-helpers-v*-linux-amd64.tar.gz (release asset / -bin)
  srcAssetHash = "sha256-wuFaCwgkwbs0qYAN7p+YIiCcLXtFoolHDRC9ltwLqow=";
  # Legacy alias used by older flakes
  srcHash = "sha256-wuFaCwgkwbs0qYAN7p+YIiCcLXtFoolHDRC9ltwLqow=";
  # SRI hash of millennium-helpers-v*-src.tar.gz (from-source packages)
  srcGitHash = "sha256-BYkx35dw419ZfZlM+L7M8JVVJLAh+Vz8OqclheN4HF0=";
}
