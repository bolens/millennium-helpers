# Agent guidance

Before Spec Kit planning or implementation, read
`.specify/memory/project-guide.md` with the project constitution. It maps
requirements to this repository's source, acceptance evidence, and validation.

Read `.specify/memory/constitution.md`, `CONTRIBUTING.md`, and the relevant
contract or runbook. The CLI contract is `spec/cli-contract.yaml`.

- Implement feature behavior in Go. Prefer shared cross-platform logic with
  small build-tagged OS files; do not recreate legacy shell/PowerShell feature
  trees.
- For commands, flags, channels, completions, man options, or MCP surfaces:
  edit the CLI contract first, run `make sync-cli-facade`, then implement and
  run `make check-cli-contract`. Do not hand-edit generated regions.
- Preserve legacy argv0 compatibility, but expose only `millennium` for new
  installs. Keep `CGO_ENABLED=0` compatibility.
- Destructive commands need dry-run and confirmation unless an explicit yes
  flag is supplied. Never expose credentials, private paths, or unsanitized
  diagnostics.
- Do not commit build/package artifacts or hand-edit `.SRCINFO`. Do not change
  versions, packaging, or release state without explicit scope; read
  `docs/release_runbook.md` for release work.
- Use idiomatic Go, Bash, Python, and PowerShell conventions already documented
  in `CONTRIBUTING.md`. Add dependencies only when existing or standard-library
  options are insufficient.
- Run the narrowest relevant tests first, then the contract/platform checks for
  touched surfaces. Use the repository's full local gate for cross-cutting or
  release-bound work and report unavailable platform checks.

## Spec-driven changes

Use Spec Kit for new capabilities, architecture, security-sensitive behavior,
migrations, and coordinated multi-file changes. Keep narrow fixes, dependency
updates, prose edits, and release housekeeping in the normal repository
workflow unless their risk warrants a written specification. Keep completed
feature directories under `specs/` as decision history. Backfill finished work
only when explicitly requested. Label those
specifications as retrospective baselines, record the inspected revision, and map
requirements to source and acceptance evidence. Separate observed behavior from
corrective requirements. Never imply the specification preceded its code or mark
unverified checks complete.

## Context and handoffs

- Locate source with targeted searches before reading. For exploratory reads of
  files over 350 lines, select relevant ranges. Read required guidance and actual
  source before edits or correctness claims; summaries do not replace them.
- When delegation is permitted, give each worker one question or concrete output,
  allowed paths, and a check. Return findings with source locations, changed paths,
  and verification gaps. Keep final review with the coordinating agent.
- Record durable user corrections in the [project guide](.specify/memory/project-guide.md)
  or owning contract with scope, reason, and evidence. Replace superseded advice;
  read relevant corrections before reusing assumptions. Keep temporary progress
  in task notes and preserve existing authority rules.
