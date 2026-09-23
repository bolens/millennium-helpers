# Validation

Validated on Linux on 2026-09-23:

- Focused Go tests cover caller resolution, required Linux binaries, permissions,
  checksum failures, rollback preservation, and doctor repair selection.
- Ownership and cache regressions cover symlinked ancestors, symlinked roots,
  external-file preservation, and directory replacement after opening.
- `make check-all` passed, including 259 Bash tests.
- `make check-cli-contract` and the repository Actionlint/Zizmor gate passed.
- `make test-windows` passed all five PowerShell tests on Linux. Deliberate
  assertion and discovery failures both produced nonzero Make exit status.
- Affected tests cross-compiled for Windows and Darwin with CGO disabled.
  Native execution on those platforms remains a CI check.
- The pinned default devcontainer built with rootless Podman. Post-create,
  smoke, `make check-all`, and `make test-windows` passed as `vscode`.
  Final cache changes also passed focused tests inside that image.
- Separate Go/security and CI reviews completed. Reported ownership/cache
  escapes and Pester failure propagation defects were corrected and rechecked.
- The validated helpers executable replaced the installed binary atomically.
  The previous executable and a rollback script were retained in a root-owned
  backup. This is a local binary override, not a package database update.

Read-only installed diagnostics correctly flagged the existing client's older
checksum manifest, which omits Lua, and its non-executable Lua helper. Doctor
selects a verified reinstall for that incomplete manifest. No live client
reinstall or Steam restart was performed as part of this implementation.
