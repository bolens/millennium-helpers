# Agent guidance

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
