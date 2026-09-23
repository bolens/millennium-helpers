# Plan: Native CLI and distribution contracts

The [specification](spec.md) preserves existing behavior. Use the project guide
and constitution for implementation constraints. Keep upstream-managed templates,
helpers, and integration manifests unchanged.

## Source ownership

- `spec/cli-contract.yaml`
- `go/internal`
- `go/cmd`
- `scripts/ci`
- `tests`
- `Makefile`

## Constitution check

Preserve canonical source ownership, compatibility, bounded inputs, and native validation. This baseline changes project-owned documentation without altering managed integration files or applying live operations.

## Validation

```sh
make check-all
```

Run checks in an isolated checkout. Commands are instructions, not evidence of
a pass. Record results in `coverage.md`, keep incomplete work in `tasks.md`, and
follow `RELEASING.md` for reviewed delivery. No live operation is required solely
to create this retrospective baseline.
