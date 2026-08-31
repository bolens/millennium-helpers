# Pinned release metadata for Nix packages.
# Updated by scripts/ci/update-packaging-versions.sh on each release.
{
  version = "4.0.0";
  # SRI hash of millennium-helpers-v*-linux-amd64.tar.gz (release asset / -bin)
  srcAssetHash = "sha256-/iIAcCCHdjOObaAFOINAyaobOgsyQMuKpMGUihl23t8=";
  # Legacy alias used by older flakes
  srcHash = "sha256-/iIAcCCHdjOObaAFOINAyaobOgsyQMuKpMGUihl23t8=";
  # SRI hash of millennium-helpers-v*-src.tar.gz (from-source packages)
  srcGitHash = "sha256-U6txLI4Q7yJT2Rv8GeEuGM8yfjmSz1BFxCCK8Bpxvxg=";
}
