# millennium-helpers Spec Kit project guide

A cross-platform Go installer and maintenance CLI for Millennium, driven by one CLI
contract.

Read this guide with `AGENTS.md` and `.specify/memory/constitution.md` before
specifying, planning, or implementing a substantial change. It is project-owned
guidance, not an upstream-managed template.

## Source and ownership map

- `spec/cli-contract.yaml`
- `go/`
- `VERSION`
- `docs/release_runbook.md`
- `tests/`

## Specification and plan decisions

Begin command, flag, channel, completion, man-page, and MCP changes in the CLI contract.
Keep feature behavior in shared Go code with small OS-specific adapters. Preserve legacy
invocation compatibility without adding parallel executable families.

## Acceptance evidence

Cover dry-run behavior, explicit confirmation, absent installations, malformed input,
sanitized diagnostics, and generated façade consistency. Identify Linux, Windows, and
macOS coverage separately and preserve CGO-disabled builds.

## Validation and operational limits

```sh
make check-cli-contract
make test-go
make check-all
```

Use make sync-cli-facade when the contract changes, then inspect generated diffs. Real
Steam/Millennium installation, repair, scheduling, and destructive operations are not
fixtures. Packaging and version changes follow the release playbook.

## Working through Spec Kit

Use Spec Kit for new capabilities, architectural or security-sensitive changes,
migrations, and coordinated changes that need a written contract. Keep narrow fixes,
dependency updates, and prose maintenance in the normal PR workflow.

For a new feature, record observable acceptance criteria in `spec.md`, source ownership
and constitution checks in `plan.md`, and evidence-bearing work in `tasks.md` under the
feature directory created by Spec Kit. Resolve material unknowns before implementation.
Mark tasks complete only after their stated verification, and distinguish completed,
skipped, blocked, and manual checks. Retain completed feature documents as decision
history; do not backfill feature specifications for already finished code.

Keep `.specify/templates/`, `.specify/scripts/`, and generated Codex skills under their
integration manifests. Use this guide and the constitution for local customization.
Regenerate managed files through Spec Kit and verify that project-owned memory survives
updates. Follow `RELEASING.md` for push, merge, release or delivery, and recovery.
