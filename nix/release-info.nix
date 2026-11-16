# Pinned release metadata for Nix packages.
# Updated by scripts/ci/update-packaging-versions.sh on each release.
{
  version = "4.1.0";
  # SRI hash of millennium-helpers-v*-linux-amd64.tar.gz (release asset / -bin)
  srcAssetHash = "sha256-nwPlfMW3cz3177pp+IlznjCZoxj6kts0D5UmJCniNbk=";
  # Legacy alias used by older flakes
  srcHash = "sha256-nwPlfMW3cz3177pp+IlznjCZoxj6kts0D5UmJCniNbk=";
  # SRI hash of millennium-helpers-v*-src.tar.gz (from-source packages)
  srcGitHash = "sha256-xFzu+7N4maIhcdUdttUoLxoq3naAL0+fe/utbX+EMBM=";
}
