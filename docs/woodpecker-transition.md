# Woodpecker validation

Run the existing make check-all and make build gates with pinned Go 1.26.11, then the existing Linux Go race suite as an isolated non-root user.

## CI organization

The [shared purpose phases](https://github.com/bolens/.github/blob/main/docs/woodpecker-transition.md#ci-organization)
map existing Make and Go commands to their current owners. `make check-all`
already includes lint, Go tests and the Bash suite; component rows are focused
entrypoints into that same contract.

| Phase | Check group / canonical local command | Native workflow / step | Retained hosted owner / boundary |
| --- | --- | --- | --- |
| `validate` | Source, CLI, packaging and docs contracts: `make lint`; focused docs: `make check-docs` | `woodpecker` / `core` | PR, main push and native manual validation; no installation or publisher secrets |
| `validate` | Shared source and changelog policy: `bash scripts/ci/check-policy.sh` | `woodpecker` / `source-and-changelog-policy` | Same full shared selectors and changelog checker. Required `ci/woodpecker/pr/woodpecker` replaces the removed hosted duplicates |
| `test` | Go and Bash suites: `make check-all`; focused Go: `make test-go` | `woodpecker` / `core` | Existing macOS, Windows/Pester, distro and development-image coverage remains hosted |
| `test` | Linux race suite: `(cd go && CGO_ENABLED=1 go test -buildvcs=false -race ./...)` with a C toolchain | `woodpecker` / `core` | Runs as an isolated non-root user; it does not replace other platform checks |
| `validate` | CLI build: `make build` | `woodpecker` / `core` | Generated output remains a build result, not a published package |
| `site` | Website source: `python3 scripts/ci/check-pages-site.py --site site --architecture docs/architecture/millennium-helpers.html --architecture-source docs/architecture/millennium-helpers.architecture.json` | Separate from the native core gate | Hosted Pages validates the versioned build and rendered accessibility |
| `publish` | Explicit version/tag and packaging procedure in the [release runbook](release_runbook.md) | Existing release owners | Hosted release gates, signing, packaging and Pages deployment retain their dependencies and trusted events |

The helper needs Node 24, Python 3 with venv/pip, Git and ShellCheck. It uses
shared tooling `7603518f305fb76f7bb1b9979f2692521f633b82`, locked npm dependencies
and Ruff 0.15.20 in disposable directories; it does not install host packages.
The workflow installs prerequisites only in its disposable container, then runs
the helper as UID 1000, matching the core workspace owner. Its pinned Node image
and direct tool pins require reviewed updates; Actions dependency automation
does not update these Woodpecker references.

Portable Go quality, CLI contracts, workflow audits, shell/Python lint,
docs/man/version checks, source policy, and Nix evaluation/builds now belong to
Woodpecker. Their duplicate GitHub workflow files and Go jobs are removed.
`check-native-quality.sh` pins golangci-lint 2.12.2, actionlint 1.7.12, and
zizmor 1.29.0 with offline medium-severity/medium-confidence blocking audits.
`check-nix.sh` evaluates every declared platform and builds available Linux
release, binary, and git packages. Nix runs in a disposable unprivileged container
with its internal build sandbox disabled, without host mounts or credentials.

The native PR status plus three native Go platform tests form the merge gate.
Cut over host settings only after observing the replacement native PR gate.
The release gate also requires the genuine main-push native status on the exact
release SHA. Native checks run on tags as well. GitHub retains native platform,
distro/container, packaging installation, CodeQL, Nushell completion, and publishing
coverage. Linux cross-compilation does not replace Windows or macOS execution.

The existing `woodpecker` alias remains stable. This index
does not change CLI contracts, version sources, generated files or publication.

The pipeline runs on pull requests and main pushes with immutable clone/runtime
images. It has no host mounts, Docker socket, privileged mode or publishing
secrets. Package installation is confined to the disposable job container.
Independent hosted platform, real Docker, security and release checks remain
separate until equivalent coverage exists.
A container or lint result does not prove installation, runtime or publication.

Require exact-head PR checks, a reviewed squash merge, and genuine native main
success before retiring duplicates. Keep existing required status names until
their replacements have passed and host settings agree. Recover through a
corrective PR without changing live profiles or installations.

Review image digests monthly and after security releases. Existing Dockerfile
updaters do not monitor these Woodpecker YAML references. Direct tool pins and
image updates require a reviewed PR and successful native execution.

CI, vulnerability scans, and release builds use Go 1.26.11 for current standard-
library security fixes. The module language minimum remains Go 1.22. The native
quality step uses golangci-lint 2.12.2, which supports this compiler. Keep the
module toolchain, native image, hosted compiler, and scan default aligned.
Nix source packages and the development shell share that security patch through
[`nix/go.nix`](../nix/go.nix). Binary packages require a newly built release asset.


## Related

- [Documentation index](README.md)
- [Project README](../README.md)
- [Release runbook](release_runbook.md)
