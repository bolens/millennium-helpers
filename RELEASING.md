# Release playbook

[Documentation](docs/README.md)

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

## Branch protection

The default branch requires pull requests, resolved conversations, linear
history, and an up-to-date branch with passing required checks, including the Go
quality, three platform tests, CLI contract, and workflow lint/security jobs.
These rules also apply to administrators; force pushes and branch deletion are
disabled. Zero approving reviews are required because this is a solo-maintainer
repository; review the complete diff before merging.

Keep required checks available on every pull request. Filter expensive work
inside jobs or use an always-running result job that rejects failures and
cancellations. Update the protection settings when renaming required jobs.

## Source lint

The Source lint workflow checks maintained javascript, css files selected by
[`.github/source-lint.json`](.github/source-lint.json) on every pull request
and push to `main`. Existing native checks remain part of the merge gate.
Use the [shared local reproduction instructions](https://github.com/bolens/.github/blob/7603518f305fb76f7bb1b9979f2692521f633b82/docs/source-lint.md)
with the same tooling revision pinned in
[the workflow](.github/workflows/source-lint.yml). Review exclusions when adding
source files; generated and imported files retain their native validation.
Require the new check to pass on the current PR head before merging.
