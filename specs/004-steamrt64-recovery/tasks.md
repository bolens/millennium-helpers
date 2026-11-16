# Tasks

- [x] Update contract, shared manifest, hook plans, diagnostics, and doctor behavior.
- [x] Cover legacy installs, incomplete x64 groups, modes, container targets,
      wrong hooks, and unsafe architecture parents with regression tests.
- [ ] Run full local validation and cross-platform builds.
- [ ] Review final changes and verify remote checks before merge.
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
