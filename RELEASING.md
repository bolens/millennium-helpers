# Release playbook

The authoritative Millennium Helpers procedure is
[`docs/release_runbook.md`](docs/release_runbook.md). It covers the multi-platform
test matrix, version and packaging synchronization, protected squash-merge
flow, merged-SHA CI gate, signed tagging, draft publication, packaging PR,
checksums, SBOM, provenance, and recovery.

Use that runbook for every `vX.Y.Z` release. Never push directly to protected
`main`; run its complete validation matrix and verify packages, checksums, SBOM,
provenance, and install paths after publication. This root entry exists so release
instructions are discoverable at the same path across the fleet. Where this
file and the project runbook differ, the project runbook is authoritative and
must be corrected in the same pull request.

Fleet policy: <https://github.com/bolens/.github/blob/main/RELEASING.md>.
