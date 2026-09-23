# Feature specification: Native CLI and distribution contracts

**Created**: 2026-09-05
**Status**: Retrospective baseline
**Inspected revision**: `2b3dc6eb79ef3d63424fa46479c10eefdac0a9b0`
**Input**: The owner requested a fleet-wide Spec Kit retrofit and implementation audit.

The declarative CLI contract owns public commands, flags, platforms, channels, and MCP schemas; the Go core implements them.

This specification records existing contracts after implementation. It does not
claim that the original work followed Spec Kit. New behavior requires a separate
change contract. Existing feature specifications remain authoritative within their
own scope.

## User scenarios and testing

### User story 1: Use the supported interface (P1)

A user selects the documented entry point.

**Acceptance**: Its outputs and failure behavior follow the source and acceptance mapping.

### User story 2: Handle invalid input (P2)

Inputs or dependencies fail validation.

**Acceptance**: The named negative fixtures retain the rejection and recovery contracts.

### User story 3: Maintain the implementation (P3)

A maintainer changes a supported contract.

**Acceptance**: Update its authoritative source, documentation, and tests together.

## Requirements

- **FR-001**: Public CLI, MCP, completions, and man options MUST remain synchronized with the declarative CLI contract.
- **FR-002**: Shared feature behavior MUST live in the Go core with explicit platform adapters and preserved legacy invocation compatibility.
- **FR-003**: Upgrade and rollback MUST retain their documented dry-run, backup selection, validation, and failure behavior.
- **FR-004**: Archive extraction and theme paths MUST reject unsafe members and retain target ownership boundaries.
- **FR-005**: MCP tools MUST retain allowlisted operations and explicit confirmation for their documented destructive actions.
- **FR-006**: Packaging versions, docs, completions, licensing, and release interfaces MUST agree with VERSION and canonical sources.

## Success criteria

- **SC-001**: Every requirement has a named source owner and acceptance check in `coverage.md`.
- **SC-002**: The listed native checks pass for the reviewed candidate, with unavailable environments and operational checks recorded separately.
- **SC-003**: Retrofitting preserves existing interfaces and completed specifications. Any confirmed implementation gap is corrected under an explicit requirement before it is marked complete.

## Edge cases and operational limits

Local Linux fixtures do not claim Windows/macOS runtime or distro-container proof. No Steam installation, upgrade, relaunch, scheduling, packaging publication, or live MCP mutation was executed.
