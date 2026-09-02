# Changelog

All notable changes to Millennium Helpers are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Pages code examples use theme-aware shell syntax highlighting without changing copied commands.

### Fixed

- Explicit light mode keeps its light palette when the operating system prefers dark mode.

## [4.0.0] - 2026-08-31

### Changed
- **Breaking:** Go consumers must migrate imports from `github.com/bolens/millenium-helpers` to `github.com/bolens/millennium-helpers`. Installer, packaging, and documentation URLs use the corrected repository name.

### Fixed
- Unix Steam discovery rejects Wine paths instead of treating them as native installations
- From-source packaging accepts the controlled release archive layout

### Security
- Release and CI builds use Go 1.25.13 to include standard-library vulnerability fixes

## [3.2.1] - 2026-08-08

### Fixed
- Fish top-level completion offers `millennium config`, with dispatcher entries generated from the CLI contract to prevent future omissions

## [3.2.0] - 2026-08-07

### Added
- Native `millennium config` commands inspect client settings, toggle installed plugins, select or disable themes, and identify/disable components with directly attributed recent errors without relying on Steam's settings UI

### Fixed
- Error remediation ignores recognized prior Steam sessions, tolerates missing component directories, and rejects action-specific options outside their valid CLI and MCP actions

## [3.1.0] - 2026-08-07

### Added
- Reversible `millennium injection status|disable|enable` controls preserve configuration, plugins, themes, and installed client files
- Injection safety checks refuse foreign or conflicting bootstrap files and report partial hook states

### Changed
- MCP registration uses `millennium mcp`, and Windows scheduled updates resolve the unified dispatcher directly

### Removed
- Retired argv0 twin dispatch, legacy execution environment switches, metadata migration, and the unused legacy Go package

## [3.0.2] - 2026-08-05

### Changed
- Repository agent guidance documents safe contribution and validation practices
- Release CI pins third-party actions, verifies downloaded tooling, and publishes build-provenance attestations

### Fixed
- Remote archives, bootstrap downloads, and GitHub API responses enforce path, entry-count, and size limits
- MCP child processes terminate as a process tree when requests time out
- Configuration updates and the setup wizard preserve malformed or unreadable state instead of overwriting it
- Linux all-users upgrades link hooks for the target user
- Failed Windows installs and rollbacks restore the previous installation transactionally
- Backup retention applies configured count and maximum age on Linux and Windows
- Theme bulk updates and uninstall cleanup report partial failures

## [3.0.1] - 2026-07-15

### Fixed
- Arch PKGBUILD/`.SRCINFO` `millennium-helpers.sudoers` sha256sums after the v3.0.0 peel (makepkg validity check)
- `check-version` asserts Arch sudoers checksums stay in sync (skips absent package trees in CI fixtures)
- MCP JSON spacing helper no longer sizes buffers with `len+N` (CodeQL `go/allocation-size-overflow`)

## [3.0.0] - 2026-07-15

### Added

- `millennium install` and `millennium uninstall` provide verified release or main-track installation on Linux and Windows, including PATH, completions, schedules, and sudoers integration.
- A contract-driven generator keeps completions, man pages, MCP schemas, and dispatch allowlists synchronized with `spec/cli-contract.yaml`.

### Changed

- **Breaking:** New installs expose only the `millennium` command. Use `millennium <command>` instead of the former long-name executables. Existing long-name dispatch remains compatible when present.
- Release archives and packages ship the Go CLI instead of duplicate Bash and PowerShell feature implementations.
- Platform CI now tests the Go-owned command surface on Linux, Windows, and macOS.

### Removed

- Retired feature scripts, PATH dispatchers, shared installer libraries, the Windows piped installer, and completed migration trackers.

## [2.7.0] - 2026-07-15

### Added

- A unified Go CLI for scheduling, themes, diagnostics, upgrades, repair, purge, and MCP operations.
- A machine-readable CLI contract with synchronized completions, man pages, MCP schemas, and cross-platform smoke tests.
- Arch, Homebrew, Scoop, Nix, deb, rpm, and Chocolatey packaging with OS-specific release archives.
- Native Linux and Windows scheduling plus an interactive setup wizard.

### Changed

- Installers require the Go dispatcher and register platform schedulers through `millennium <command>`.
- Release assets are versioned by OS and architecture; source packages use controlled archives.

### Removed

- Shell and PowerShell PATH dispatchers, duplicate feature libraries, dead shared helpers, and the Python MCP fallback.

## [2.6.2] - 2026-07-10

### Fixed
- Windows↔Unix parity: schedule `config` under StrictMode; scheduled task uses `-Yes -Quiet`, theme update, and `updater.log`
- Windows upgrade honors `backup_max_age_days` and accepts GNU-style channel/force/file/rollback flags
- Windows repair supports `-SkipTheme` / `--skip-theme` and refreshes themes after force reinstall
- Setup wizard (Unix + Windows) preserves `backup_limit` / `backup_max_age_days` when rewriting config

## [2.6.1] - 2026-07-10

### Changed
- Arch `-git` packaging follows AUR VCS policy: no perpetual `pkgver` sync on every commit; `pkgver()` is authoritative at `makepkg` time; `.SRCINFO` regenerates only when the `-git` recipe changes (`make sync-git-srcinfo`)

### Fixed
- Portable packaging CI sed (GNU vs BSD/macOS) and Windows InstallTrack I/O so Cross-Platform Test Suite stays green
- Diag install-track meta no longer breaks when `USERPROFILE` is unset or tests mock `Get-Content`

## [2.6.0] - 2026-07-10

### Added
- Helpers install tracks (`release` / `main` / `tag` / `checkout`) with `install-meta.json`, `--track`/`--tag` (Unix) and `-Track`/`-Tag` (Windows)
- Auto-migrate legacy installs without meta on first install/diag/doctor touch
- Winget tip-of-main manifests (`packaging/winget-git/`, `bolens.millenniumhelpers.git`)
- Millennium client channel `main` (tip-of-development) alongside `stable`/`beta` — separate from helpers track
- Diag JSON fields `helpers_track` / `helpers_ref`; track-aware doctor sync

### Changed
- Doctor and update checks follow the recorded helpers track (pinned tags stay pinned; `-git` packages are not compared to release tags)

### Fixed
- `sync-stable-srcinfo` no longer truncates `.SRCINFO` when `makepkg --printsrcinfo` fails (e.g. missing `install=` file in test trees)

## [2.5.0] - 2026-07-10

### Added
- Versioned Arch package (`packaging/millennium-helpers`) built from the Linux release tarball; release CD bumps pkgver + sha256 alongside Formula/Scoop/Winget/Nix
- Scoop `millennium-helpers-git` nightly manifest (tip of `main` via GitHub archive)
- Nix flake packages: `millennium-helpers` (release tarball) and `millennium-helpers-git` (flake source / latest commit)

### Changed
- Arch `-git` recipe moved to `packaging/millennium-helpers-git/` (AUR-ready layout)

### Fixed
- `millennium-schedule setup` under sudo verifies passwordless rules for the real user (`sudo -U`), matching doctor

## [2.4.0] - 2026-07-10

### Added
- Modular `millennium-diag` libraries on Linux (`scripts/lib/diag_*.sh`) and Windows (`scripts/windows/lib/Diag*.ps1`)
- Install-method detection (pacman/scoop/winget/manual/mixed) with release-tag comparison
- Doctor cleanup of unmanaged leftovers before package-upgrade hints
- PowerShell completion health checks and doctor repair
- `-Yes` / `--yes` auto scoop/winget/pacman package upgrade when helpers are outdated
- JSON fields: `install_method`, `mixed_install_ok`, `helpers_checkout`, `latest_release_tag`, `completions_ok`

### Fixed
- Pacman upgrades no longer blocked by unmanaged completion leftovers; doctor refuses to overwrite package-owned files
- Empty `DIAG_TEST_OBSOLETE_LIST` correctly means “no obsolete candidates” in tests
- Winget package ID uses `bolens.millenniumhelpers`
- Windows PowerShell 5.1 no longer fails to parse diag modules (ASCII-only sources; Scoop profile hook no longer calls `scoop` at profile load)

### Changed
- Thin loaders: Linux `diag.sh` (replaces `diag_report.sh`), Windows `Diag.ps1` (replaces `DiagReport.ps1`)
- Manual installs sync helper scripts from the latest release tarball/zip instead of `main`

## [2.3.0] - 2026-07-10

### Added
- PowerShell Tab completions with installer profile hooks (Windows + Scoop)
- Cross-shell completion tests (bash/zsh/fish/nushell) and nested zsh simulation
- Homebrew bash/zsh completion symlinks and Nushell completion install
- Isolated prefix install/uninstall coverage in `test_install.sh`

### Fixed
- Fish completions no longer ship bare `VERSION_PLACEHOLDER` tokens
- `sudo millennium-schedule` reaches the user systemd bus via shared `sysctl_user`
- Uninstall disables cron as well as systemd/LaunchAgent schedulers
- Windows uninstall removes the `MillenniumUpdate` scheduled task
- Shared library resolution works from Homebrew-style `prefix/bin` layouts
- Cron enable/disable under sudo targets the invoking user's crontab
- PKGBUILD conflict cleanup docs include bare `millennium` artifacts

### Changed
- Linux installer also installs Nushell completions into the user config path
- Release Windows zip includes `completions/powershell/`
- Manual uninstall docs cover PowerShell hooks, state dirs, and Scoop hooks

## [2.2.1] - 2026-07-10

### Security
- Verify SHA256 checksums in piped `install.sh` against release `.sha256` sidecars
- Verify SHA256 checksums during Windows `millennium-upgrade` (parity with Linux)
- Restrict Windows `config.json` ACLs when writing `github_token`

### Fixed
- Include `scripts/millennium-mcp.py` in the trimmed Windows release zip / Scoop CI staging
- Update standalone installer tests for trimmed release archive layout
- Republish release assets so packaging checksums match the Windows zip that includes MCP

### Changed
- Wire `test_millennium_dispatcher.sh` and `test_packaging_ci.sh` into the CI matrix
- Slim local `tests/run_tests.sh` to source shared assertions and defer packaging gates to CI

## [2.2.0] - 2026-07-09

See GitHub release notes for v2.2.0.
