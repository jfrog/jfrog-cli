#!/usr/bin/env bash
#
# run-apm-tests.sh — run the `jf agent apm` e2e test suite (agent_apm_test.go) locally.
#
# Mirrors .github/workflows/agentApmTests.yml's "Run agent apm tests" step:
#   go test -v github.com/jfrog/jfrog-cli --timeout 0 --test.apm [--jfrog.url=... --jfrog.adminToken=...]
#
# Usage:
#   ./run-apm-tests.sh                              # run every APM test against localhost:8081
#   JFROG_URL=https://host JFROG_ADMIN_TOKEN=xxx ./run-apm-tests.sh   # run against an external JFrog Platform
#   ./run-apm-tests.sh -run TestApmInstallWithBuildInfo               # run one test (any `go test` flag works)
#   ./run-apm-tests.sh -run 'TestApmInstall|TestApmPublish'           # run a feature group (regex)
#   ./run-apm-tests.sh --list                        # print every APM test name and exit
#
# Env vars:
#   JFROG_URL          JFrog Platform URL to test against (default: http://localhost:8081/, i.e. a
#                       locally running Artifactory — start one yourself, e.g. via `docker run` or
#                       the apmbughunt skill, before running this script).
#   JFROG_ADMIN_TOKEN   Admin access token for JFROG_URL. Required whenever JFROG_URL is not
#                       localhost; optional locally (the test harness can self-mint one).

set -euo pipefail
cd "$(dirname "$0")"

if [[ "${1:-}" == "--list" ]]; then
  grep -oh '^func \(Test[A-Za-z0-9_]*\)' agent_apm_test.go | sed 's/^func //'
  exit 0
fi

# --- 1. Make sure the `apm` CLI (Microsoft's Agent Package Manager) is on PATH -------------------
# jf shells out to the real `apm` binary; the test suite doesn't stub it.
if ! command -v apm >/dev/null 2>&1; then
  echo "==> apm CLI not found; installing to \$HOME/.local/bin (no sudo needed)..."
  curl -sSL https://aka.ms/apm-unix | APM_INSTALL_DIR="$HOME/.local/bin" sh
  export PATH="$HOME/.local/bin:$PATH"
fi
echo "==> Using apm: $(command -v apm) ($(apm --version 2>&1 | head -1))"

# --- 2. Resolve the target JFrog Platform --------------------------------------------------------
JFROG_URL="${JFROG_URL:-http://localhost:8081/}"
EXTRA_FLAGS=(--jfrog.url="${JFROG_URL}")
if [[ -n "${JFROG_ADMIN_TOKEN:-}" ]]; then
  EXTRA_FLAGS+=(--jfrog.adminToken="${JFROG_ADMIN_TOKEN}")
fi

echo "==> Target JFrog Platform: ${JFROG_URL}"
if ! curl -sf --max-time 5 "${JFROG_URL%/}/artifactory/api/system/ping" >/dev/null 2>&1; then
  echo "!!  Could not reach ${JFROG_URL} — is Artifactory running there?"
  echo "    Start a local instance first, or point at an external one:"
  echo "      JFROG_URL=https://your-server JFROG_ADMIN_TOKEN=xxx ./run-apm-tests.sh"
  exit 1
fi

# --- 3. Run the suite ------------------------------------------------------------------------------
echo "==> Running: go test -v github.com/jfrog/jfrog-cli --timeout 0 --test.apm ${EXTRA_FLAGS[*]} $*"
go test -v github.com/jfrog/jfrog-cli --timeout 0 --test.apm "${EXTRA_FLAGS[@]}" "$@"
