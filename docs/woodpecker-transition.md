# Woodpecker validation

Run the existing make check-all and make build gates with pinned Go 1.26.10, then the existing Linux Go race suite as an isolated non-root user.

The pipeline runs on pull requests and main pushes with immutable clone/runtime
images. It has no host mounts, Docker socket, privileged mode or publishing
secrets. Package installation is confined to the disposable job container.
Existing GitHub checks remain during native verification. Platform, real Docker,
security and release checks remain separate until equivalent coverage exists.
A container or lint result does not prove installation, runtime or publication.

Require exact-head PR checks, a reviewed squash merge, and genuine native main
success before retiring duplicates. Keep existing required status names until
their replacements have passed and host settings agree. Recover through a
corrective PR without changing live profiles or installations.

Review image digests monthly and after security releases. Existing Dockerfile
updaters do not monitor these Woodpecker YAML references. Direct tool pins and
image updates require a reviewed PR and successful native execution.

CI, vulnerability scans, and release builds use Go 1.26.10 for current standard-
library security fixes. The module language minimum remains Go 1.22. The hosted
quality job uses golangci-lint 2.12.2, which supports this compiler. Keep the
module toolchain, native image, hosted compiler, and scan default aligned.


## Related

- [Documentation index](README.md)
- [Project README](../README.md)
- [Release runbook](release_runbook.md)
