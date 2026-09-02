# Millennium Helpers Constitution

## Core Principles

### I. Contract-First CLI
Commands, flags, platforms, channels, completions, man options, and MCP surfaces MUST begin in `spec/cli-contract.yaml`; generated façades follow from that source.

### II. Cross-Platform Go Core
Feature behavior belongs in Go and SHOULD be shared across platforms. Genuine OS differences use small build-tagged implementations without diverging public behavior unnecessarily.

### III. Safe System Modification
Install, repair, upgrade, scheduling, and destructive operations MUST support preview or dry-run where practical and require confirmation unless the user supplied an explicit non-interactive consent flag.

### IV. Sanitized Diagnostics and Compatibility
Diagnostics MUST not expose tokens, private paths, or sensitive state. New installs expose `millennium`; legacy invocation compatibility is preserved without multiplying new executable surfaces.

### V. Synchronized Distribution and Verification
Contracts, implementation, tests, help, docs, completions, man pages, MCP, and packaging MUST agree. Generated files come from sync targets; platform gaps and skipped checks are reported.

## Governance

`CONTRIBUTING.md` and the release runbook define detailed process. Contract or platform exceptions require explicit rationale, tests, and semantic versioning of this constitution.

**Version**: 1.0.0 | **Ratified**: 2026-08-15 | **Last Amended**: 2026-08-15
