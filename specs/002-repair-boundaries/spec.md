# Repair and validation boundaries

Repair, diagnosis, and rollback must agree on the invoking user, supported platform,
and required client files. These corrections preserve existing commands and consent.

## Acceptance

- Elevated repair uses the sudo caller's home and default XDG directories, not root's.
  Invalid caller resolution fails without changing files. Explicit Steam overrides remain supported.
  Ownership repair does not follow symlinks in targets, ancestors, or descendants.
- Linux install, integrity checks, permissions, and rollback share the same required files,
  including the Lua helper. Private executable modes are unhealthy for a shared installation.
- Missing or corrupt files select reinstall, not chmod. Permission-only failures select chmod.
- Rollback validates a backup and restores helper modes before replacing the active install.
  A rejected backup leaves the active install and backup contents intact. Dry-run never chmods.
- Darwin and Windows do not receive Linux runtime checks or Linux doctor actions.
- CI builds the default devcontainer, runs post-create/smoke as vscode, then native checks.
  CodeQL subactions update together. Existing required jobs retain their names.
- The installed helpers are updated only after validation, with a recoverable previous binary.

## Follow-up seam acceptance

- Hook writes from repair and upgrade reject symlinked roots and architecture parents.
- Invalid or missing candidate version metadata rejects rollback before activation.
  Active metadata cannot escape the backup root, and name collisions retain earlier backups.
- Elevated config writes remain readable and writable by the invoking user.
- Ownership diagnostics inspect the actual repair targets. Failed repairs return failure.
- Repairs do not start until Steam has exited. After a successful stop, both successful
  and failed maintenance attempt relaunch using the captured graphical environment.
  Failed relaunch retains recovery state and contributes to the command's failure status.
