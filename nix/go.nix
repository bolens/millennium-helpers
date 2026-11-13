# Shared security patch for Nix source packages and the development shell.
# The locked nixpkgs Go 1.26 compiler predates CVE-2026-97054.
{ pkgs }:
pkgs.go_1_26.overrideAttrs (_: {
  version = "1.26.11";
  src = pkgs.fetchurl {
    url = "https://go.dev/dl/go1.26.11.src.tar.gz";
    hash = "sha256-1EEBsP8n7SpMda1GwXucdDWSCeLrLKfJ32y0/baOGWk=";
  };
})
