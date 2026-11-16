# CLI and transaction consistency

Preserve commands, aliases, valid options, installation tracks, and MCP tools.
Reject ambiguous or ignored CLI input before network or filesystem operations.
Never echo configured credentials in mutation acknowledgements or dry-runs.

Serialize conflicting helper mutations with kernel-managed, process-scoped
locks. A competing operation must fail before altering installation, backup,
or configuration state. Locks must release on process exit without stale-file
recovery heuristics. Existing external applications do not participate in these
advisory locks and must still be closed before changing their files.

MCP must preserve request identifiers exactly, return structured protocol errors,
and never run a tool from a notification. Replace JSON whitespace assertions with
semantic protocol and failure-path tests. Preserve the existing tool catalog.

Use exact version metadata, deterministic ordering, and explicit failures in place
of invented time-based values. Keep Windows/macOS/Linux and CGO-disabled builds.
