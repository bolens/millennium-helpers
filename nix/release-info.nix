# Pinned release metadata for Nix packages.
# Updated by scripts/ci/update-packaging-versions.sh on each release.
{
  version = "4.1.2";
  # SRI hash of millennium-helpers-v*-linux-amd64.tar.gz (release asset / -bin)
  srcAssetHash = "sha256-NxG2xe+oUMQIUhD9dLcV+DE5HGxww8EE9aLfmW6jH9g=";
  # Legacy alias used by older flakes
  srcHash = "sha256-NxG2xe+oUMQIUhD9dLcV+DE5HGxww8EE9aLfmW6jH9g=";
  # SRI hash of millennium-helpers-v*-src.tar.gz (from-source packages)
  srcGitHash = "sha256-Ci9sCAHLkRtGpJ014H063LWE5xzWUNuYh/djGYyfjUk=";
}
