{
  description = "Cross-platform utility scripts and Model Context Protocol (MCP) server for managing, upgrading, diagnosing, and controlling Millennium on Linux";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    nixpkgs-darwin.url = "github:NixOS/nixpkgs/nixpkgs-26.05-darwin";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, nixpkgs-darwin, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        # The main input no longer supports the declared Intel macOS shell.
        pkgs = import (if system == "x86_64-darwin" then nixpkgs-darwin else nixpkgs) { inherit system; };
        isLinux = pkgs.stdenv.hostPlatform.isLinux;
        hasBinary = system == "x86_64-linux";
        releaseInfo = import ./nix/release-info.nix;
        goToolchain = import ./nix/go.nix { inherit pkgs; };
        mkPackage = args: pkgs.callPackage ./nix/default.nix ({
          go = goToolchain;
          buildGoModule = pkgs.buildGoModule.override { go = goToolchain; };
        } // args);

        # From-source tagged release (default): build Go from controlled -src.tar.gz.
        millennium-helpers = mkPackage {
          version = releaseInfo.version;
          src = pkgs.fetchurl {
            url = "https://github.com/bolens/millennium-helpers/releases/download/v${releaseInfo.version}/millennium-helpers-v${releaseInfo.version}-src.tar.gz";
            hash = releaseInfo.srcGitHash;
          };
          buildGoDispatcher = true;
        };

        # Prebuilt release assets (scripts tarball; Go binary when embedded).
        millennium-helpers-bin = mkPackage {
          pname = "millennium-helpers-bin";
          version = releaseInfo.version;
          src = pkgs.fetchurl {
            url = "https://github.com/bolens/millennium-helpers/releases/download/v${releaseInfo.version}/millennium-helpers-v${releaseInfo.version}-linux-amd64.tar.gz";
            hash = releaseInfo.srcAssetHash or releaseInfo.srcHash;
          };
          unpackFlat = true;
          buildGoDispatcher = false;
        };

        millennium-helpers-git = mkPackage {
          pname = "millennium-helpers-git";
          version =
            if self ? shortRev then "unstable-${self.shortRev}"
            else if self ? dirtyShortRev then "unstable-${self.dirtyShortRev}"
            else "unstable-dirty";
          src = pkgs.lib.cleanSource ./.;
          buildGoDispatcher = true;
        };
      in
      {
        packages = pkgs.lib.optionalAttrs isLinux ({
          inherit millennium-helpers millennium-helpers-git;
          default = millennium-helpers;
        } // pkgs.lib.optionalAttrs hasBinary {
          inherit millennium-helpers-bin;
        });

        apps = pkgs.lib.optionalAttrs isLinux {
          default = {
            type = "app";
            program = "${millennium-helpers}/bin/millennium";
            meta.description = millennium-helpers.meta.description;
          };
          millennium-helpers-git = {
            type = "app";
            program = "${millennium-helpers-git}/bin/millennium";
            meta.description = millennium-helpers-git.meta.description;
          };
        } // pkgs.lib.optionalAttrs hasBinary {
          millennium-helpers-bin = {
            type = "app";
            program = "${millennium-helpers-bin}/bin/millennium";
            meta.description = millennium-helpers-bin.meta.description;
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            bash
            python3
            python3Packages.pyyaml
            curl
            unzip
            git
            shellcheck
            ruff
            goToolchain
            gnumake
          ];
        };
      }
    );
}
