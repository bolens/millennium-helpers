#!/usr/bin/env bash
set -euo pipefail

# pre-commit installs PyYAML in its isolated Python environment. Preserve that
# environment for repository scripts that intentionally invoke `python3`.
site_packages="$(python -c 'import site; print(site.getsitepackages()[0])')"
export PYTHONPATH="${site_packages}${PYTHONPATH:+:${PYTHONPATH}}"

# Nested fixtures set PRE_COMMIT themselves when they test re-stage behavior.
unset PRE_COMMIT

exec make check-all
