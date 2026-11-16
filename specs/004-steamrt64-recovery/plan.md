# Plan

Extend the shared client manifest with a complete optional x64 component group.
Share hook definitions between planning, diagnostics, upgrade, and repair. Keep
Linux descriptor-based writes and validate all sources before replacing links.
Add a SteamRT3 diagnostic and reuse the existing doctor upgrade action.

CLI contract first. No new flags or dependencies. Windows and Darwin retain
existing behavior. Verify fixtures before live repair. Validate with focused Go
tests, contract sync, full local gate, race tests, and cross-platform builds.

Upstream evidence: Millennium v3.6.3 release archive and scripts/install.sh as
inspected 2026-11-16. The SteamRT3 bootstrap is x64, not the webhelper hhx64 library.

## CI scope added by the user

Move portable Go quality, contract, workflow audits, Python, docs/man/version,
source policy, and Nix builds to Woodpecker. Remove migrated hosted workflows
and jobs. Preserve native platform and release checks. Migrate required status
settings after the native gate passes. GitHub runtime is available after workflow
approvals, so retain unmatched hosted work.
