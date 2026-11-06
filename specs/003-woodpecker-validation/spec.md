# Portable Woodpecker validation

Run the existing make check-all and make build gates with pinned Go 1.25.13, then the existing Linux Go race suite as an isolated non-root user.

Run on PR heads, main pushes and manual validation in immutable, ordinary Docker
containers. Supply no publishing credentials, host mounts or Docker socket.
Preserve existing platform, actual Docker, security and publication checks until
an equivalent native gate is observed. Exact native PR and genuine main success
are required before retiring duplicate hosted checks. No product release or
operational installation is part of this migration.
