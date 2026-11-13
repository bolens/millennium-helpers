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
`ci/woodpecker/pr/woodpecker`, Go quality, three platform tests, the CLI contract,
and workflow lint/security jobs.
These rules also apply to administrators; force pushes and branch deletion are
disabled. Zero approving reviews are required because this is a solo-maintainer
repository; review the complete diff before merging.

Keep required checks available on every pull request. Filter expensive work
inside jobs or use an always-running result job that rejects failures and
cancellations. Update the protection settings when renaming required jobs.

## Source lint

The native `source-and-changelog-policy` step checks maintained JavaScript and
CSS files selected by
[`.github/source-lint.json`](.github/source-lint.json) on every pull request
and push to `main`, using the shared selectors and existing changelog checker through
`bash scripts/ci/check-policy.sh`. See the [CI index](docs/woodpecker-transition.md#ci-organization).
The duplicate hosted Source lint and Changelog workflows are disabled after
exact-source native PR and main acceptance. Require the native
`ci/woodpecker/pr/woodpecker` check alongside the seven retained Go, CLI and
workflow-security checks.
Use the [shared local reproduction instructions](https://github.com/bolens/.github/blob/7603518f305fb76f7bb1b9979f2692521f633b82/docs/source-lint.md)
with the same tooling revision pinned in
[the helper](scripts/ci/check-policy.sh). The hosted workflow definitions remain
available for reviewed rollback. Review exclusions when adding
source files; generated and imported files retain their native validation.
Require all active checks to pass on the current PR head before merging.

## Woodpecker validation

[The transition contract](docs/woodpecker-transition.md) adds portable Linux
checks alongside existing host validation. Require actual native PR and main
success before retiring duplicates. Existing platform, security, packaging and
publishing authorities remain in place. No new tag or product release is part
of this CI change.
