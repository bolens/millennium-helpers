# Restore Millennium and Steam logging on Linux

For an existing Millennium installation on **native Linux x86-64 Steam**, including
Arch. This is not a Flatpak or NixOS procedure. For Deck or Flatpak, see
[Steam Deck troubleshooting](steam_deck.md). If only logging is broken, go to
[step 4](#4-restore-logging-and-start-steam).

The official [Millennium v3.6.3 stable release](https://github.com/SteamClientHomebrew/Millennium/releases/tag/v3.6.3)
contains the required x64 runtime. We verified it with Steam's SteamRT3 client on
2026-11-16. A beta/main channel can still resolve to an older, incompatible build.
The commands below deliberately select the verified stable archive. Millennium's
version is separate from the millennium-helpers version.

Prefer an updated distribution package when available. This manual workaround
changes files outside the package manager, whose recorded version stays unchanged.
Use a terminal with Bash, curl, tar, sha256sum, sudo, and the standard Linux tools.
Run each block as your regular user. Only the installation commands use sudo.

## 1. Exit Steam completely

Close all games, then choose **Steam → Exit**. Both commands should print nothing:

```bash
# Do not replace libraries while either process is still running.
pgrep -a -x steam
pgrep -a -x steamwebhelper
```

## 2. Install the verified runtime

This block verifies the download, backs up the current runtime, and installs the
release files. It stops before installation if checksum verification fails.

```bash
bash <<'BASH'
set -euo pipefail
if pgrep -x steam >/dev/null || pgrep -x steamwebhelper >/dev/null; then
  echo "Exit Steam before continuing." >&2
  exit 1
fi
work_dir=$(mktemp -d)
trap 'rm -rf -- "$work_dir"' EXIT
version=3.6.3
archive="millennium-v${version}-linux-x86_64.tar.gz"
release="https://github.com/SteamClientHomebrew/Millennium/releases/download/v${version}"
# Verify the official archive before changing installed files.
curl -fL "$release/$archive" -o "$work_dir/$archive"
curl -fL "$release/millennium-v${version}-linux-x86_64.sha256" -o "$work_dir/check.sha256"
(cd "$work_dir" && sha256sum --check check.sha256)
tar -xzf "$work_dir/$archive" -C "$work_dir"
# Preserve the existing runtime for rollback.
if [ -d /usr/lib/millennium ]; then
  backup="/usr/lib/millennium.before-steamrt3-$(date +%Y%m%d%H%M%S)"
  sudo cp -a /usr/lib/millennium "$backup"
  echo "Saved previous installation: $backup"
fi
sudo install -d -m 755 /usr/lib/millennium
sudo install -m 644 "$work_dir/usr/lib/millennium/"*.so /usr/lib/millennium/
# The Lua and pressure-vessel helpers need executable permissions.
sudo install -m 755 "$work_dir/usr/lib/millennium/libmillennium_pvs64" \
  "$work_dir/usr/lib/millennium/libmillennium_luavm_x86" \
  "$work_dir/usr/lib/millennium/libmillennium_luavm_x64" /usr/lib/millennium/
printf '%s\n' "$version" | sudo tee /usr/lib/millennium/version.txt >/dev/null
# Record integrity checks for the full runtime payload.
(cd "$work_dir/usr/lib/millennium" && \
  sha256sum libmillennium_*.so libmillennium_luavm_* libmillennium_pvs64) > "$work_dir/checksums.txt"
sudo install -m 644 "$work_dir/checksums.txt" /usr/lib/millennium/checksums.txt
BASH
```

## 3. Add the 64-bit client hook

Run this block as your regular user while Steam remains closed:

```bash
bash <<'BASH'
set -euo pipefail
if pgrep -x steam >/dev/null || pgrep -x steamwebhelper >/dev/null; then
  echo "Exit Steam before continuing." >&2
  exit 1
fi
# Change this assignment if Steam is installed elsewhere.
steam_root=$(readlink -f "$HOME/.steam/steam")
test -f "$steam_root/steamrt64/steam"
hook="$steam_root/steamrt64/libXtst.so.6"
test ! -d "$hook"
# Preserve any existing hook.
if [ -e "$hook" ] || [ -L "$hook" ]; then
  backup="$hook.before-steamrt3-$(date +%Y%m%d%H%M%S)"
  cp -a -- "$hook" "$backup"
  echo "Saved previous hook: $backup"
fi
# This target resolves inside Steam's pressure-vessel container.
ln -sfn /run/host/usr/lib/millennium/libmillennium_bootstrap_x64.so "$hook"
readlink "$hook"
BASH
```

The `/run/host` target resolves inside pressure-vessel and may look broken from the
host. Use **bootstrap_x64** for this hook. Keep the existing legacy and webhelper
hooks. The `bootstrap_hhx64` library serves steamwebhelper, not the main client.
No global `LD_PRELOAD` or `PRESSURE_VESSEL_APP_LD_PRELOAD` override is required.

## 4. Restore logging and start Steam

Remove **`-srt-logger-opened`** from any custom launcher, wrapper, or saved restart
command. Steam adds this internal flag after connecting its logger. Reusing it
without the inherited logger connection prevents a fresh log from opening.

With Steam fully closed, run:

```bash
# Enable logging, use Steam's default log folder, and avoid debugger mode.
# Start through the normal launcher as your regular user.
env -u DEBUGGER -u STEAM_CLIENT_LOG_FOLDER STEAM_RUNTIME_LOGGER=1 steam
```

These environment settings apply to this launch. If logging fails again through a
desktop shortcut, remove `STEAM_RUNTIME_LOGGER=0` and the stale flag from that
shortcut or wrapper. If you intentionally use `STEAM_CLIENT_LOG_FOLDER`, keep its
custom value and look there for logs instead of the default `logs` directory.

For users of millennium-helpers, version **4.1.2 or newer** removes the stale flag
on automatic relaunch while preserving explicit logging preferences.

## 5. Verify loading and fresh logs

Check that Millennium's menu appears in Steam. In another terminal, run:

```bash
# Check that timestamps match this launch, rather than an older successful start.
tail -n 150 "$HOME/.steam/steam/logs/console-linux.txt" |
  grep -E 'Started Millennium|SteamRT3 bootstrap|Loaded hhx64'
```

Fresh `Started Millennium` and `Loaded hhx64` messages confirm Millennium started
in the client and webhelper. For a custom Steam root, replace
`$HOME/.steam/steam` with its actual path. To follow new output:

```bash
# Ctrl+C stops viewing the log without closing Steam.
tail -n 0 -F "$HOME/.steam/steam/logs/console-linux.txt"
```

## Package updates and rollback

Once the distribution package contains the compatible runtime, update or reinstall
it so the package database and installed files agree. Avoid downgrading the working
runtime to an older beta without the x64 components.

To reverse this workaround, exit Steam completely, restore the runtime directory
and hook from the printed `.before-steamrt3-*` backups, then restart Steam.
An older runtime may still fail with the new client. Keep the backups until a
compatible package update has been verified.

## Related

- [Documentation index](README.md)
- [Project README](../README.md)
- [Security and helper diagnostics](security_troubleshooting.md)
- [Steam Deck and Flatpak](steam_deck.md)
