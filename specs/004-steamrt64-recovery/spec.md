# SteamRT3 recovery

Steam now runs a 64-bit Linux client in pressure-vessel. Existing helpers can
report healthy legacy hooks while the active client has no compatible bootstrap.

Acceptance criteria:
- Detect SteamRT3 from its client executable or opt-in marker without creating it.
- Report missing x64 client components separately from legacy integrity checks.
- Doctor upgrades incompatible clients before repairing hooks. Preserve the selected channel.
- Install and diagnose the SteamRT3 hook using the container /run/host path.
- Validate and checksum all three x64 components when present, including executable Lua modes.
- Keep legacy clients, injection toggles, dry-run, confirmation, safe directory traversal,
  and CGO-disabled cross-platform builds working.
- Do not change Steam channels, global preload settings, or user plugin configuration.
