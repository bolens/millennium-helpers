# Tasks

- [x] Reproduce parser, configuration, transaction, and MCP defects.
- [x] Implement corrections and replace formatting-only protocol tests.
- [x] Verify focused tests, race suite, contract sync, full local gate, and Windows
      cross-compilation. Native platform execution remains a remote merge gate.
- [x] Complete separate self-review. Exact-head native/platform gates and protected
      merge remain delivery requirements tracked on the pull request.

Audit coverage: dispatcher exit handling, strict diagnostic and helper-config
parsing, credential acknowledgements, helper/client config read-modify-write,
install/rollback serialization, Windows backup allocation, JSON-RPC boundaries,
log snapshots/rotation, and privileged safe-filesystem operations. Removed the test-driven JSON whitespace formatter and unused
legacy selectors and their constant-return tests after confirming no production
callers. Added direct safe-filesystem and cross-process lock tests. Remaining
unmeasured diagnostic checks are explicit nulls, not completed validation.

Local evidence includes direct safefs coverage of 65.3% and directory-lock coverage
of 85.7%, including process-exit release. The tests cover security-relevant guards
and filesystem outcomes rather than manufacturing complete syscall-error coverage.
No independent reviewer was used. The expanded native quality script and all local
checks pass with Go 1.26.11. Direct Windows execution is required for lock and log
rotation behavior, not inferred from cross-compilation.

The final self-review also found float64 conversion of unrelated JSON numbers
and unsafe MCP-host registration on unreadable/non-object documents. Shared strict
object decoding now preserves numeric text, and registration uses the same kernel
lock boundary, validates its namespace, and preserves invalid input unchanged.
