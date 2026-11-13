# Packaging matrix

| Variant | Meaning | Typical source |
| --- | --- | --- |
| **from-source** (`millennium-helpers`) | Tagged `vX.Y.Z`, build Go when applicable | GitHub tag archive + `make build` |
| **bin** (`millennium-helpers-bin`) | Published release assets | `millennium-helpers-v{VER}-{os}-{arch}.tar.gz` / `.zip` |
| **git** (`millennium-helpers-git`) | Tip of `main` | VCS / `main.zip` / flake source |

**AUR-standard naming** applies on Arch, Homebrew, Scoop, Nix, deb, and rpm.
Winget and Chocolatey are **bin** (plus Winget **git**); they do not ship a from-source package.

| Channel | from-source | bin | git |
| --- | --- | --- | --- |
| Arch | `packaging/millennium-helpers/` | `packaging/millennium-helpers-bin/` | `packaging/millennium-helpers-git/` |
| Homebrew | `Formula/millennium-helpers.rb` (+ `head`) | `Formula/millennium-helpers-bin.rb` | `head` on source formula |
| Scoop | `packaging/scoop/millennium-helpers.json` | `…-bin.json` | `…-git.json` |
| Nix | `packages.millennium-helpers` | `…-bin` | `…-git` |
| Winget | — | `packaging/winget/` | `packaging/winget-git/` |
| deb | `packaging/deb/millennium-helpers/` + `build-from-source.sh` | `…-bin/` + `build-bin.sh` | tip build via from-source script |
| rpm | `packaging/rpm/millennium-helpers.spec` | `…-bin.spec` | tip via from-source `.spec` |
| Chocolatey | — | `packaging/chocolatey/millennium-helpers/` | — |

Release CD publishes versioned, OS/arch-trimmed bin packs (each embeds the
matching Go dispatcher), controlled `-src.tar.gz` / `-src.zip` trees for
from-source packaging, and standalone `millennium-v{VER}-{os}-{arch}` binaries.
There are no legacy unversioned asset aliases.

Nix source packages and `nix develop` share the reviewed compiler in
[`nix/go.nix`](../nix/go.nix). Its Go 1.26.11 source and hash apply the standard-
library security patch to the locked nixpkgs compiler. Update it alongside the
native and hosted compiler pins and verify both source variants and the shell.
Prebuilt `-bin` packages retain the compiler embedded in the published asset;
changing the Nix compiler cannot patch an existing binary release.

The flake exposes source packages and CLI apps on x86_64-linux and
aarch64-linux. The published Nix binary asset is linux-amd64, so `-bin` is
available only on x86_64-linux. Darwin package outputs are not advertised by
the Linux packaging contract. Development shells remain declared on both Linux
architectures and both Darwin architectures. Intel Darwin uses a separate
maintained `nixpkgs-26.05-darwin` input because the main nixpkgs input removed
that system; its compiler uses the same reviewed Go patch. Check all declared
outputs when updating either input. Derivation evaluation checks configuration
and platform admission; it does not establish native execution or cross-build
coverage.

Run `nix flake check --all-systems --no-build --no-write-lock-file` before the
native package builds; the existing Nix CI job enforces this admission check.
