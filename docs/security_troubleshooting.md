# Security Design & Troubleshooting

This document outlines the security architecture and troubleshooting steps for the Millennium Helper utility suite.

For general usage instructions, see the main [README.md](../README.md). Full
docs index: [README.md](README.md). For MCP server setup details, see
[mcp.md](mcp.md). Licensing: [licensing.md](licensing.md).

---

## Security Design

To allow background user-level timers to run updates that modify system directories, the updater scripts must run with elevated privileges.

This setup achieves this securely:
1. **Sudoers Autoconfiguration (Linux)**: During `sudo ./install.sh`, the installer detects the original invoking user (`SUDO_USER`) and automatically configures a secure drop-in file at `/etc/sudoers.d/millennium-helpers`.
2. **Write-Protected Scripts**: Helper scripts are copied into `/usr/local/bin/` owned by `root:root` with `755` permissions, meaning normal users cannot edit or tamper with them.
3. **Task Scheduler (Windows)**: Scheduled tasks are registered with elevated credentials using native Windows Task Scheduler security boundaries.
4. **Channel allow-list**: Scheduler enable paths only accept `stable|beta|main`. Poisoned `update_channel` values in config are ignored on load and rejected before embedding into cron/systemd/Task Scheduler command lines.
5. **Scheduler-only hooks**: `pre-update` / `post-update` require `MILLENNIUM_SCHEDULER=1` (set by the unit/cron). Manual invocation is refused.
6. **Release checksums and provenance**: Non-`main` helper/client downloads hard-fail if the `.sha256` sidecar is missing or mismatches. Tip-of-main installs require `--allow-unsigned-main` / `-AllowUnsignedMain`. Same-origin GitHub sidecars are TOFU; published helper releases additionally include a checksummed SPDX SBOM and GitHub build-provenance attestations.
7. **Resource ceilings**: Go HTTP downloads are capped at 512 MiB and API responses at 4 MiB. Go archive extraction rejects more than 100,000 entries, any file larger than 512 MiB, or more than 2 GiB of total extracted data.
8. **Local archives**: `millennium upgrade --file` requires `--sha256` or explicit `--insecure-skip-verify`.
9. **Doctor**: Syncing helper scripts over root-owned install paths requires `millennium diag doctor --yes` after a verified release download.
10. **Archive extract**: Theme and Windows client zips reject path-traversal members (`..` / absolute paths).

Arch packages grant `%wheel` passwordless access to the four privileged commands. On a shared desktop with broad wheel membership, prefer the curl installer's user-specific sudoers rule.

---

## Troubleshooting & FAQ

### Steam WebHelper fails with Permission denied

If the error names `pressure-vessel-unruntime` under a `.millennium-pv-*`
directory, check the runtime helper permissions with `millennium diag`.
The `Runtime Helper Modes` check detects missing executable permissions on
`libmillennium_pvs64` and `libmillennium_luavm_x86`.

Preview recovery with `sudo millennium diag doctor --dry-run`, then run
`sudo millennium diag doctor --yes`. Doctor restores mode `0755` when binary
integrity is healthy. `sudo millennium repair --skip-theme --yes` also restores
these modes. Upgrades normalize the helper permissions even if the archive
contains them with mode `0644`. Private modes such as `0700` are also unhealthy
for a shared install. If an unprivileged diagnosis cannot read a helper, it asks
for a sudo check before deciding whether a reinstall is needed.

Linux integrity checks require all six client components, including the Lua
helper, in `checksums.txt`. Rollback rejects incomplete or corrupt backups before
moving the active install and restores helper modes before activation. Older
backups with incomplete manifests require a verified reinstall instead. Rollback
also requires valid version metadata. Install and rollback retain existing backups
on name collisions and never use unchecked metadata as a deletion path.

Elevated repair resolves the sudo caller's home and default XDG directories.
Ownership repair skips symlinks and rejects paths with symlinked parent
directories. Cache cleanup also rejects symlinked roots or parents and removes
children through open directory handles. Hook installation uses the same policy
in repair and upgrade. Failed ownership, hook, or theme repairs return a nonzero
status. `permissions_ok` checks ownership of discovered repair targets, including
the helpers configuration directory. Use the real Steam directory when an explicit `STEAM` override
points through a symlink.

### Steam shows a blank/black screen after upgrading Millennium
This is usually caused by outdated CEF cached files. Run the repair utility to fix local permissions and clear Steam's htmlcache:
```bash
# Linux
sudo millennium repair

# Windows
millennium repair
```

### The background timer is not running updates
1. Check the timer status:
   ```bash
   millennium schedule status
   ```
2. Verify that passwordless sudo (Linux) or Scheduled Tasks (Windows) are configured correctly.
3. If you want the updates to run even when you are logged out of your session on Linux, enable user lingering:
   ```bash
   loginctl enable-linger $USER
   ```

### Steam Deck or Flatpak Steam issues
Hooks, sandbox overrides, Desktop Mode install, and post-SteamOS-update recovery are covered in the [Steam Deck & Flatpak Troubleshooting](steam_deck.md) guide.

Elevated helpers config writes assign caller ownership before atomic replacement.
Repair and doctor capture Steam's session before shutdown, verify the client has
exited, and attempt relaunch even after maintenance fails. A relaunch failure keeps
the saved session for recovery and makes the command fail.

## Steam starts but Millennium is missing on Linux

SteamRT3 runs the 64-bit client inside pressure-vessel. Run `millennium diag`
to check **SteamRT3 Compatibility** and **Bootstrap Hooks**, then preview repairs
with `millennium diag doctor --dry-run`. JSON includes `steamrt3`,
`steamrt3_compatible`, `steamrt3_detail`, and `hooks_ok` for this path.

A healthy legacy install can still lack the x64 bootstrap, client library, and Lua
helper. Upgrade Millennium to v3.6.3 or newer through the owning package manager or
`millennium upgrade --channel stable`, then run `millennium repair`. Doctor retains
your configured update channel. If that channel still supplies an incompatible
client, choose a compatible release explicitly and rerun diagnostics.

The SteamRT3 hook is `steamrt64/libXtst.so.6` and targets
`/run/host/usr/lib/millennium/libmillennium_bootstrap_x64.so` inside the container.
The `hhx64` bootstrap belongs to steamwebhelper and cannot replace the x64 client
bootstrap. Diagnostics validate the host library behind `/run/host` rather than
requiring that container path to exist on the host. Legacy `ubuntu12_32` and
`ubuntu12_64` hooks remain installed. No global preload override is needed.

Fixture checks establish file and hook consistency. Confirm live recovery after a
Steam restart by checking for Millennium startup messages and its menu in Steam.

### Recovery without millennium-helpers

These steps are for native Linux x86-64 Steam, including Arch, when the installed
Millennium package still lacks SteamRT3 support. They are not a Flatpak or NixOS
procedure. Upstream [Millennium v3.6.3](https://github.com/SteamClientHomebrew/Millennium/releases/tag/v3.6.3)
already includes the required files. Prefer an updated distribution package when
available. The manual installation below changes files outside the package manager,
so its recorded version will remain unchanged until you reinstall an updated package.

1. Close games and exit Steam completely. `pgrep -x steam` should print nothing.
   Do not replace libraries while Steam is running.
2. Download and verify the upstream archive, preserve the current installation,
   then install its files. Run this block in a terminal:

   ```bash
   bash <<'BASH'
   set -euo pipefail
   if pgrep -x steam >/dev/null; then
     echo "Exit Steam before continuing." >&2
     exit 1
   fi
   work_dir=$(mktemp -d)
   version=3.6.3
   archive="millennium-v${version}-linux-x86_64.tar.gz"
   release="https://github.com/SteamClientHomebrew/Millennium/releases/download/v${version}"
   curl -fL "$release/$archive" -o "$work_dir/$archive"
   curl -fL "$release/millennium-v${version}-linux-x86_64.sha256" -o "$work_dir/check.sha256"
   (cd "$work_dir" && sha256sum --check check.sha256)
   tar -xzf "$work_dir/$archive" -C "$work_dir"
   if [ -d /usr/lib/millennium ]; then
     backup="/usr/lib/millennium.before-steamrt3-$(date +%Y%m%d%H%M%S)"
     sudo cp -a /usr/lib/millennium "$backup"
     echo "Saved previous installation: $backup"
   fi
   sudo install -d -m 755 /usr/lib/millennium
   sudo install -m 644 "$work_dir/usr/lib/millennium/"*.so /usr/lib/millennium/
   sudo install -m 755 "$work_dir/usr/lib/millennium/libmillennium_pvs64" \
     "$work_dir/usr/lib/millennium/libmillennium_luavm_x86" \
     "$work_dir/usr/lib/millennium/libmillennium_luavm_x64" /usr/lib/millennium/
   printf '%s\n' "$version" | sudo tee /usr/lib/millennium/version.txt >/dev/null
   (cd "$work_dir/usr/lib/millennium" && \
     sha256sum libmillennium_*.so libmillennium_luavm_* libmillennium_pvs64) > "$work_dir/checksums.txt"
   sudo install -m 644 "$work_dir/checksums.txt" /usr/lib/millennium/checksums.txt
   rm -rf -- "$work_dir"
   BASH
   ```

   Stop if checksum verification fails. The block exits before installing anything.
   The executable-mode repair is necessary because this release archive stores the
   Lua and pressure-vessel helpers without executable permissions.
3. Add the SteamRT3 hook as your regular user. This preserves an existing hook:

   ```bash
   bash <<'BASH'
   set -euo pipefail
   steam_root=$(readlink -f "$HOME/.steam/steam")
   test -f "$steam_root/steamrt64/steam"
   hook="$steam_root/steamrt64/libXtst.so.6"
   test ! -d "$hook"
   if [ -e "$hook" ] || [ -L "$hook" ]; then
     cp -a -- "$hook" "$hook.before-steamrt3-$(date +%Y%m%d%H%M%S)"
   fi
   ln -sfn /run/host/usr/lib/millennium/libmillennium_bootstrap_x64.so "$hook"
   readlink "$hook"
   BASH
   ```

   If Steam is installed elsewhere, set `steam_root` to its actual root instead.
   The `/run/host` target is resolved inside pressure-vessel. It may appear broken
   from the host. Use `bootstrap_x64`, not `bootstrap_hhx64`, for this hook.
4. Start Steam normally. Confirm Millennium appears in Steam and look for new
   `Started Millennium` and `SteamRT3 bootstrap` entries in
   `~/.steam/steam/logs/console-linux.txt`. Check their timestamps to avoid mistaking
   an older successful launch for the current one.
5. Once your distribution package catches up, reinstall or update it so the package
   database and installed files agree. For rollback, exit Steam, restore the saved
   installation and hook, then restart Steam. An old bootstrap still cannot load
   into the new client, so rollback only reverses this workaround.

This procedure does not need a global `LD_PRELOAD` or
`PRESSURE_VESSEL_APP_LD_PRELOAD` override.

## Related

- **Docs index:** [README.md](README.md)
- **Project:** [README.md](../README.md) · [SECURITY.md](../SECURITY.md) · [CONTRIBUTING.md](../CONTRIBUTING.md)
- **Guides:** [mcp.md](mcp.md) · [steam_deck.md](steam_deck.md) · [uninstall_dryrun.md](uninstall_dryrun.md) · [licensing.md](licensing.md)
