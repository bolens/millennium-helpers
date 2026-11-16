# Plan

Audit entry points, parser contracts, state writers, activation/rollback, JSON-RPC,
and tests. Work follows the SteamRT3/CI change without changing its validated head.

Implement narrow corrections for confirmed defects. Share a small directory lock
adapter using x/sys already present in the module. Unix uses flock on a directory;
Windows uses a persistent lock file with LockFileEx. Use immediate contention
errors and release descriptors in all paths. No timeout sleeps or stale PID files.

Update the CLI contract before parser policy changes. Exercise rejected input,
configuration preservation, two-process lock contention, malformed requests,
large JSON-RPC IDs, writer errors, and process timeouts through observable outputs.
Record remaining evidence gaps without treating stubs or skipped tests as success.
