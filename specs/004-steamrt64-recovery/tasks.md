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

Live verification on 2026-11-16 confirmed that the Steam game guard refuses
maintenance while an app is running and that the beta activation guard keeps
the working v3.6.3 runtime intact. A verified reinstall of that same client
version updated all nine integrity entries; doctor then reported no issues
among measured checks. The user timer was enabled for beta.

Live relaunch exposed a nested shell-quote decoding defect. Publication of the
v4.1.0 draft was cancelled, its signed tag was preserved, and its packaging PR
was closed. The v4.1.1 candidate decodes generated single-quote escapes,
preserves empty captured arguments, and does not retry a failed sudo mutation.
Shell-execution tests cover exact argument/environment preservation and refuse
accidental command execution. An error-exit test proves sudo runs once.
Further live tests are deferred at the user's request while a Steam app is in
use; final release/package verification must not restart Steam.

The local gate exposed an existing scheduler test depending on the host account
home. Scheduler target resolution now shares usercontext's HOME/sudo boundary,
and disable tests stub systemctl while checking both scopes and dry-run
non-execution. The test-created beta timer is no longer installed; the original
unscheduled state is preserved while beta remains the selected channel.

Automatic relaunch logging follow-up (v4.1.2): Steam's installed steam.sh
treats -srt-logger-opened as an inherited logger connection. Detached helper
restarts had retained that flag despite having no inherited logging pipe. Linux
relaunch now filters the exact transport marker from current and legacy saved
argument forms, retaining other quoting, empty arguments, and user values.
Malformed saved quoting fails before launch and preserves recovery state.

On 2026-11-16, two consecutive live repair --yes --skip-theme runs with the
4.1.2 candidate exited successfully and automatically relaunched Steam. Each
cycle wrote fresh Started Millennium and Loaded hhx64 entries to
console-linux.txt. The new Steam client's stdout was a pipe rather than
/dev/null; client and webhelper processes mapped their Millennium libraries.
Runtime binary hashes were unchanged across both cycles. No manual restart
was used to produce this logging evidence.

Logging preferences are captured and replayed as well: an explicit
STEAM_RUNTIME_LOGGER opt-out and a custom STEAM_CLIENT_LOG_FOLDER are retained.
The real-shell regression test checks quoted logging preference values and
user arguments without executing their contents.
