# Tasks

- [x] Update contract, shared manifest, hook plans, diagnostics, and doctor behavior.
- [x] Cover legacy installs, incomplete x64 groups, modes, container targets,
      wrong hooks, and unsafe architecture parents with regression tests.
- [x] Run full local validation and Windows package cross-compilation. Native
      Windows/macOS execution remains a hosted merge gate.
- [x] Complete a separate local self-review of runtime, injection, integrity, CI,
      and release consumers. Remote checks must pass before merge.
- [x] Verify live loading separately from source/fixture validation.

Live evidence: on 2026-11-16, verified upstream v3.6.3 SHA256, normalized
archive executable modes for the installed older helper, upgraded through the
existing sudo-authorized helper, enabled the new hook, and restarted Steam.
The new Steam process mapped libmillennium_bootstrap_x64.so and
libmillennium_x64.so. Steamwebhelper mapped libmillennium_hhx64.so and the log
recorded Started Millennium. Package database version still records the prior
beta until the owning package is updated. The installed older helper writes only
legacy checksums, so the expanded integrity manifest requires a later install
through the updated helper or package tooling.

Local evidence: make check-all passed with 249 shell checks, Go unit and race
suites passed, the pinned native quality script passed, actionlint and zizmor
passed, and all declared Nix platforms evaluated. The secret scan reported only
two existing synthetic token fixtures and no skipped files. Staged gitleaks and
commit hooks passed. No independent reviewer was used.

Release activation follow-up: reject legacy archives and rollback candidates
before replacing the active install when a targeted account has SteamRT3. This
includes beta/main channels resolving to an older beta. Tests cover active and
hook preservation, untouched backup permissions, compatible components, custom
Steam roots, and all-users detection. Existing rollback fixtures isolate HOME
so they do not depend on the developer's installed Steam version.
