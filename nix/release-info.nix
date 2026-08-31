# Pinned release metadata for Nix packages.
# Updated by scripts/ci/update-packaging-versions.sh on each release.
{
  version = "4.0.0";
  # SRI hash of millennium-helpers-v*-linux-amd64.tar.gz (release asset / -bin)
  srcAssetHash = "sha256-y3Tv5y12eTMGfzDZShgLOoDZSAp2y/1TC9W7lwHgGzI=";
  # Legacy alias used by older flakes
  srcHash = "sha256-y3Tv5y12eTMGfzDZShgLOoDZSAp2y/1TC9W7lwHgGzI=";
  # SRI hash of millennium-helpers-v*-src.tar.gz (from-source packages)
  srcGitHash = "sha256-qQ/jTifRKvwEJFj/nnAEkSJsJ3J6UzoJHp8OfzCt5x0=";
}
