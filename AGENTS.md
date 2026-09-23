# Agent guidance

[Documentation](docs/README.md) maps architecture, deployment, state, and document ownership.

Read [.specify/memory/constitution.md](.specify/memory/constitution.md), [CONTRIBUTING.md](CONTRIBUTING.md), and the relevant
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
  [docs/release_runbook.md](docs/release_runbook.md) for release work.
- Use idiomatic Go, Bash, Python, and PowerShell conventions already documented
  in [CONTRIBUTING.md](CONTRIBUTING.md). Add dependencies only when existing or standard-library
  options are insufficient.
- Run the narrowest relevant tests first, then the contract/platform checks for
  touched surfaces. Use the repository's full local gate for cross-cutting or
  release-bound work and report unavailable platform checks.

## Planning and evidence

Use the [project guide](.specify/memory/project-guide.md) and
[constitution](.specify/memory/constitution.md) for substantial changes. The guide
owns Spec Kit scope, retained history, retrospective requirements, and acceptance
evidence. Prose maintenance uses the normal repository workflow.

## Context and handoffs

- Search before reading. Use bounded source excerpts for exploratory reads over
  350 lines, and inspect required guidance and actual source before editing.
- When delegation is permitted, assign a bounded question or output, paths, and
  check. Return source locations, changes, and verification gaps for final review.
- Keep durable corrections in the [project guide](.specify/memory/project-guide.md)
  or owning contract. Replace superseded advice and read it before reuse.
  Temporary progress belongs in task notes. Preserve existing authority rules.
