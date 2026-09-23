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

## Follow-up seam fixes

- Directory-handle hook installation rejects symlinked architecture directories
  in both repair and upgrade. Outside-file preservation regressions pass.
- Rollback rejects missing or invalid version metadata before activation. Invalid
  active metadata cannot form deletion paths, and backup collisions preserve the
  earlier backup.
- Elevated config writes preserve caller ownership. Unit tests and a rootless
  container check confirm overrides preserve shared parent directory metadata.
- Diagnostics check actual ownership. Doctor's ownership step preserves cache
  files, and repair propagates ownership, hook, and theme errors.
- Shared maintenance verifies shutdown and attempts relaunch after work failure.
  Regressions cover graphical environment transfer, launcher failure, startup
  timeout, recovery-state retention, and combined work/relaunch errors.
- The final `make check-all` gate passed, including 259 Bash tests. All five
  PowerShell tests passed. Affected packages cross-compiled for Darwin and Windows
  with CGO disabled. Native platform execution remains a CI check.
- Independent filesystem and lifecycle/config reviews completed. Three additional
  findings were reproduced, fixed, and rechecked without remaining findings.
- The final binary was installed atomically and its SHA-256 matched the build.
  The original rollback backup remains intact. Installed read-only diagnostics
  ran successfully without restarting Steam or modifying the client.
