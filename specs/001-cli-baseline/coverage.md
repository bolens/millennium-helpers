# Requirement coverage

| Requirement | Source and acceptance evidence |
| --- | --- |
| FR-001 | `spec/cli-contract.yaml`, check-cli-contract and sync-cli-facade --check; completion runtime fixtures. |
| FR-002 | `go/internal`, command/argv compatibility tests and platform-specific build-tagged fixtures. |
| FR-003 | `go/internal/upgrade`, upgrade and rollback tests; dry-run returns before platform mutation. |
| FR-004 | `go/internal/archive`, `go/internal/theme/zip.go`, traversal and theme mutation fixtures. |
| FR-005 | `go/internal/mcp/dispatch.go`, dispatch tests for purge and disable-errors; schemas derive from the CLI contract. |
| FR-006 | Makefile lint/check-version/check-docs/check-man/check-completions and shell packaging fixtures. |

## Verification receipt

Full native make check-all passed, including shell/Go lint, vulnerability checks, Go test packages, generated CLI facade checks, packaging version/docs/completions checks, and 259 shell checks. Separate self-review traced rollback selection/dry-run, safe extraction, MCP confirmation, and declarative facade ownership.
