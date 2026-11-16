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
history, and an up-to-date branch with passing required checks, including
`ci/woodpecker/pr/woodpecker` and the three native Go platform tests.
These rules also apply to administrators; force pushes and branch deletion are
disabled. Zero approving reviews are required because this is a solo-maintainer
repository; review the complete diff before merging.

Keep required checks available on every pull request. Filter expensive work
inside jobs or use an always-running result job that rejects failures and
cancellations. Update the protection settings when renaming required jobs.

## Portable checks

Woodpecker owns shell/Python lint, docs/man/version consistency, CLI contracts,
Go quality and Linux race tests, shared source/changelog policy, workflow audits,
and Nix evaluation/builds. Migrated GitHub workflows and jobs are removed.
See the [CI ownership map](docs/woodpecker-transition.md#ci-organization).
The release gate requires a successful genuine Woodpecker main push on the exact
release SHA alongside the retained platform and publication checks.

GitHub retains native platform tests, distro/container and package installation
checks, CodeQL, and publishing. Keep these checks until replacements execute in
their intended environments. Image and direct tool pins require reviewed updates.
