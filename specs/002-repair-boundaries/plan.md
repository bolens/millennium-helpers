# Implementation plan

Use standard-library Go packages for the Linux binary manifest and invoking-user
context. Keep Steam discovery shared with theme and diagnostics. Retain explicit
resource overrides and isolate sudo context tests through injected account lookups.

Validate rollback integrity before any rename; canonicalize helper modes only after
validation. Do not run Linux actions on Darwin. Test incomplete manifests, missing
Lua helper, private modes, malformed backups, dry-run, and platform boundaries.

Add a read-only GitHub-hosted devcontainer workflow using the existing pinned checkout
and Dockerfile. Run a disposable writable checkout without host credentials or Docker
socket inside the container. Group CodeQL updates in the existing Dependabot config.

Constitution: CLI contract first; shared Go behavior; confirmation/dry-run preserved;
no credentials or diagnostic private paths; full local and affected platform gates.
