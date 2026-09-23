# Documentation

Contract-first Millennium installation and maintenance.

## Start here

| Need | Owning document |
| --- | --- |
| Use the project | [README.md](../README.md) |
| Change the repository | [AGENTS.md](../AGENTS.md) |
| Deliver or recover | [RELEASING.md](../RELEASING.md) |
| Plan substantial changes | [.specify/memory/project-guide.md](../.specify/memory/project-guide.md) |
| Non-negotiable constraints | [.specify/memory/constitution.md](../.specify/memory/constitution.md) |

## Architecture

[The CLI contract](../spec/cli-contract.yaml) owns commands, flags, channels, completions, man
options, and MCP surfaces. Shared Go code implements behavior, with small platform adapters. Update
the contract before generating facade changes. The [architecture
diagram](architecture/millennium-helpers.html) maps the implementation.

## Deployment and recovery

[Installation](../README.md) owns user setup. [RELEASING.md](../RELEASING.md) and the [release
runbook](release_runbook.md) own distribution and recovery. Preserve legacy invocation compatibility
while exposing `millennium` for new installs. Source validation does not authorize changes to a live
Steam installation.

## Database and state

The CLI manages filesystem installation and diagnostic state, not a repository-owned database
service. [Uninstall dry-run behavior](uninstall_dryrun.md) owns deletion previews. [Security
troubleshooting](security_troubleshooting.md) owns sanitized diagnostics. Preserve user
configuration and distinguish package rollback from reversing installer side effects.

## Documentation maintenance

Keep decisions, invariants, failure modes, and recovery requirements in the owning document. Link to
commands, defaults, schemas, and generated catalogs instead of copying them. Change the owner and
affected references together. Update this index when adding or moving a guide, and verify relative
links and heading anchors. Historical specs and audits describe their recorded revision, not current
runtime proof. A topic without an implementation stays explicitly unimplemented.

## Topic guides

Index of Millennium Helpers docs. `make check-docs` enforces that this index, the project
[README](../README.md) **Further reading** table, and each guide’s
**Related** section stay cross-linked.

## Guides

| Doc | Audience | Summary |
| --- | --- | --- |
| [licensing.md](licensing.md) | Users & packagers | Helpers MIT + Millennium client MIT, vendored notice, upgrade behavior |
| [mcp.md](mcp.md) | Users | MCP server tools, registration with Claude / Cursor / Windsurf |
| [security_troubleshooting.md](security_troubleshooting.md) | Users | Sudoers / Task Scheduler design, common failures |
| [steam_deck.md](steam_deck.md) | Users | Steam Deck (SteamOS) and Flatpak Steam hooks / overrides |
| [uninstall_dryrun.md](uninstall_dryrun.md) | Users | Dry-run mode and manual uninstall per install method |
| [release_runbook.md](release_runbook.md) | Maintainers | Cutting a `vX.Y.Z` release (preflight → tag → packaging) |
| [architecture/millennium-helpers.html](architecture/millennium-helpers.html) | Contributors | Interactive command, scheduler, MCP, platform, and client architecture |
| [site-evidence/README.md](site-evidence/README.md) | Contributors | Pages accessibility audit and desktop/mobile layout evidence |

## Project root docs

| Doc | Summary |
| --- | --- |
| [README.md](../README.md) | Install, commands, config, further reading |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Dev requirements, layout, versioning, packaging, licensing |
| [packaging/README.md](../packaging/README.md) | from-source / bin / git packaging matrix (Arch, brew, Scoop, Nix, deb, rpm, Chocolatey) |
| [SECURITY.md](../SECURITY.md) | Vulnerability reporting and security design overview |
| [CHANGELOG.md](../CHANGELOG.md) | Release notes |
| [LICENSE](../LICENSE) | Millennium Helpers MIT license |
| [third_party/](../third_party/README.md) | Vendored third-party notices (Millennium client) |

## Manual pages

Installed with the helpers (`man millennium`, `man millennium-diag`, …). Each page has a `LICENSE`
section pointing at [licensing.md](licensing.md); command-specific pages also point at the matching
guide where one exists (for example `millennium-mcp(1)` → [mcp.md](mcp.md)).

## Keeping links in sync

1. Add or rename a guide under `docs/` → update **this index**, [README § Further reading](../README.md#further-reading), and topical **Related** footers.
2. Run `make check-docs` (includes licensing checks; also part of `make lint` / `make check-all`).

## Related

- [Project README](../README.md) · [CONTRIBUTING.md](../CONTRIBUTING.md) · [SECURITY.md](../SECURITY.md) · [CHANGELOG.md](../CHANGELOG.md) · [licensing.md](licensing.md)

- [Editor setup](../.vscode/README.md)
