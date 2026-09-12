#!/usr/bin/env bash
#
# apm-smoke-test.sh — comprehensive, live smoke test of `jf agent apm` against a real
# Artifactory instance. Unlike run-apm-tests.sh (which runs the go test suite's own throwaway
# repo/packages), this script drives the real `jf`/`apm` CLIs directly against packages you can
# inspect afterwards, covering every feature bucket:
#
#   setup / install / publish / native-flag+--server-id passthrough / native-apm-direct /
#   auth types (basic-auth-as-configured + a freshly minted access token) /
#   build-info collection, required-flags behavior, properties, checksums, and readback /
#   error handling (bad package, bad flag, bad server) / --dev / --module / --frozen / --zip
#
# Publishes a fresh 3-level dependency chain under owner "jfrog" into the existing
# "udaykb-apm-local" repo:
#   jfrog/simple      (no dependencies)
#   jfrog/consumer    (depends on jfrog/simple)
#   jfrog/transitive  (depends on jfrog/consumer)
#
# Not covered here (needs infra this script doesn't set up — see run-apm-tests.sh instead):
#   apm CLI minimum-version gating, multiple registries in one install, registry
#   precedence/fallback, --project scoping (no known project key on this instance).
#
# Usage:
#   ./apm-smoke-test.sh                # run everything
#   ./apm-smoke-test.sh --keep         # don't delete the local scratch dir afterwards
#
# Env vars (all optional, shown with their defaults):
#   JF          Path to a jf binary built with agent-apm support.
#               Default: $HOME/apm-bughunt/$USER/bin/jf if present, else "jf" on PATH.
#   SERVER_ID   jf server config to use.        Default: apmtest
#   REPO        Artifactory agentpackages repo. Default: udaykb-apm-local
#   OWNER       Package owner/namespace.        Default: jfrog
#   WORKDIR     Scratch directory for package sources + install test projects.
#               Default: $HOME/apm-smoke-test

set -uo pipefail

KEEP=false
[[ "${1:-}" == "--keep" ]] && KEEP=true

DEFAULT_JF="$HOME/apm-bughunt/${USER:-$(id -un)}/bin/jf"
if [[ -x "${JF:-}" ]]; then
  : # explicit JF env var wins
elif [[ -x "$DEFAULT_JF" ]]; then
  JF="$DEFAULT_JF"
else
  JF="jf"
fi
SERVER_ID="${SERVER_ID:-apmtest}"
REPO="${REPO:-udaykb-apm-local}"
OWNER="${OWNER:-jfrog}"
WORKDIR="${WORKDIR:-$HOME/apm-smoke-test}"
STAMP="$(date +%Y%m%d%H%M%S 2>/dev/null || echo run)"
TMP_SERVER_ID="apm-smoke-token-${STAMP}"

FAILURES=0
log()  { echo "==> $*" >&2; }       # stderr: some functions (e.g. publish_package) return data via stdout
ok()   { echo "    OK: $*" >&2; }
fail() { echo "    FAIL: $*" >&2; FAILURES=$((FAILURES + 1)); }
skip() { echo "    SKIP: $*" >&2; }

cleanup() {
  "$JF" config rm "$TMP_SERVER_ID" --quiet >/dev/null 2>&1 || true
  if [[ "$KEEP" == false ]]; then
    rm -rf "$WORKDIR"
  fi
}
trap cleanup EXIT

# run_expect_success <label> -- <command...>
# Retries up to 3 times on a 403/Forbidden: this shared instance has a known transient
# token-permission race right after 'jf setup apm' mints a fresh registry token (same thing
# the apmbughunt skill's seeder script retries on) - not a real, persistent auth failure.
run_expect_success() {
  local label="$1"; shift; [[ "$1" == "--" ]] && shift
  local attempt=0
  while :; do
    if out="$("$@" 2>&1)"; then
      ok "$label"
      echo "$out" | sed 's/^/      /' >&2
      return 0
    fi
    attempt=$((attempt + 1))
    if echo "$out" | grep -qiE "403|forbidden" && [[ $attempt -le 3 ]]; then
      log "    transient 403/Forbidden — retrying in 3s (attempt $attempt/3): $label"
      sleep 3
      continue
    fi
    fail "$label (expected success, got exit $?)"
    echo "$out" | sed 's/^/      /' >&2
    return 1
  done
}

# run_expect_failure <label> -- <command...>
run_expect_failure() {
  local label="$1"; shift; [[ "$1" == "--" ]] && shift
  if out="$("$@" 2>&1)"; then
    fail "$label (expected failure, but it succeeded)"
  else
    ok "$label (failed as expected: $(echo "$out" | tail -1))"
  fi
}

# --- 0. Preflight -----------------------------------------------------------------------------
log "jf binary   : $JF"
if ! "$JF" agent apm --help >/dev/null 2>&1; then
  echo "ERROR: '$JF agent apm' isn't available. Point JF at a jf binary built with agent-apm" >&2
  echo "       support (see the 'jfrog-cli-local-build' skill, or the apmbughunt skill's build)." >&2
  exit 1
fi
command -v apm >/dev/null 2>&1 || { echo "ERROR: apm CLI not found on PATH." >&2; exit 1; }
log "apm CLI     : $(command -v apm) ($(apm --version 2>&1 | head -1))"

if ! "$JF" config show "$SERVER_ID" >/dev/null 2>&1; then
  echo "ERROR: jf server '$SERVER_ID' isn't configured. Run 'jf config add $SERVER_ID ...' first." >&2
  exit 1
fi
PLATFORM_URL="$("$JF" config export "$SERVER_ID" | base64 --decode | jq -r '.url')"
log "server      : $SERVER_ID ($PLATFORM_URL)"
log "repo        : $REPO"
log "owner       : $OWNER"
log "workdir     : $WORKDIR"
echo

# --- 1. FEATURE: Setup -------------------------------------------------------------------------
log "[1. setup] jf setup apm --repo=$REPO --server-id=$SERVER_ID"
run_expect_success "setup apm against $REPO" -- "$JF" setup apm --repo="$REPO" --server-id="$SERVER_ID"
echo

log "[1b. setup error handling] setup against an unconfigured server-id should fail cleanly"
run_expect_failure "setup apm with a bogus --server-id" -- \
  "$JF" setup apm --repo="$REPO" --server-id="no-such-server-$STAMP"
echo

mkdir -p "$WORKDIR"

# publish_package <pkg-name> [<dep-id> <dep-version>]
# Writes a minimal apm.yml (+ a token .apm/ dir) under $WORKDIR/<pkg-name>, publishes it with
# build-info, publishes that build-info, and echoes the version it actually published as
# (bumping past 409 "already exists" conflicts, same idea as the apmbughunt skill's seeder).
write_package_source() {
  local name="$1" version="$2" dep_id="${3:-}" dep_version="${4:-}"
  local dest="$WORKDIR/$name"
  mkdir -p "$dest/.apm/instructions"
  cat > "$dest/.apm/instructions/general.md" <<EOF
# $name
Smoke-test package published by apm-smoke-test.sh.
EOF
  if [[ -n "$dep_id" ]]; then
    cat > "$dest/apm.yml" <<EOF
name: $name
version: $version
author: $OWNER
license: UNLICENSED
dependencies:
  apm:
  - registry: $REPO
    id: $dep_id
    version: $dep_version
  mcp: []
includes: auto
scripts: {}
EOF
  else
    cat > "$dest/apm.yml" <<EOF
name: $name
version: $version
author: $OWNER
license: UNLICENSED
dependencies:
  apm: []
  mcp: []
includes: auto
scripts: {}
EOF
  fi
}

publish_package() {
  local name="$1" dep_id="${2:-}" dep_version="${3:-}"
  local dest="$WORKDIR/$name"
  local build_name="smoke-publish-${name}-${STAMP}"

  rm -rf "$dest"
  write_package_source "$name" "1.0.0" "$dep_id" "$dep_version"

  log "[2. publish] $OWNER/$name -> $REPO (build $build_name/1)"
  local attempt=0
  while :; do
    local version out
    version="$(grep '^version:' "$dest/apm.yml" | awk '{print $2}')"
    if out="$(cd "$dest" && "$JF" agent apm publish --package "$OWNER/$name" --registry "$REPO" \
        --server-id "$SERVER_ID" --build-name "$build_name" --build-number 1 -v 2>&1)"; then
      echo "$out" | sed 's/^/      /' >&2
      log "[2b. build-publish] jf rt bp $build_name 1 --server-id $SERVER_ID"
      "$JF" rt build-publish "$build_name" 1 --server-id "$SERVER_ID" 2>&1 | sed 's/^/      /' >&2
      ok "published $OWNER/$name@$version, build-info published as $build_name/1"
      echo "$version"
      return 0
    fi
    echo "$out" | sed 's/^/      /' >&2
    attempt=$((attempt + 1))
    if echo "$out" | grep -qi "already exists" && [[ $attempt -le 20 ]]; then
      local major minor patch next
      IFS='.' read -r major minor patch <<< "$version"
      next="${major}.${minor}.$((patch + 1))"
      log "    version conflict — bumping $version -> $next and retrying"
      sed -i.bak -e "s/^version: ${version}\$/version: ${next}/" "$dest/apm.yml"
      rm -f "$dest/apm.yml.bak"
      continue
    fi
    fail "publish of $name (not a version conflict, or retry limit reached)"
    echo "$version"
    return 1
  done
}

# --- 2. FEATURE: Publish (+ build-info collection & publishing), 3-level chain ------------------
SIMPLE_VERSION="$(publish_package simple)"
echo
CONSUMER_VERSION="$(publish_package consumer "$OWNER/simple" "$SIMPLE_VERSION")"
echo
TRANSITIVE_VERSION="$(publish_package transitive "$OWNER/consumer" "$CONSUMER_VERSION")"
echo

log "[2c. publish error handling] publish without --package should fail cleanly"
run_expect_failure "publish with no --package" -- bash -c \
  "cd '$WORKDIR/simple' && '$JF' agent apm publish --registry '$REPO' --server-id '$SERVER_ID'"
echo

log "[2d. publish --zip] publish a pre-built zip as a new transitive version"
ZIP_VERSION="${TRANSITIVE_VERSION%.*}.$(( ${TRANSITIVE_VERSION##*.} + 1 ))"
write_package_source transitive-zip "$ZIP_VERSION" "$OWNER/consumer" "$CONSUMER_VERSION"
mv "$WORKDIR/transitive-zip/apm.yml" "$WORKDIR/transitive-zip/apm.yml.tmp"
sed "s/^name: transitive-zip/name: transitive/" "$WORKDIR/transitive-zip/apm.yml.tmp" > "$WORKDIR/transitive-zip/apm.yml"
rm -f "$WORKDIR/transitive-zip/apm.yml.tmp"
( cd "$WORKDIR/transitive-zip" && zip -qr "$WORKDIR/transitive-zip.zip" apm.yml .apm )
ZIP_BUILD="smoke-publish-zip-${STAMP}"
run_expect_success "publish transitive@$ZIP_VERSION via --zip" -- bash -c \
  "cd '$WORKDIR/transitive-zip' && '$JF' agent apm publish --package '$OWNER/transitive' --registry '$REPO' \
     --zip '$WORKDIR/transitive-zip.zip' --server-id '$SERVER_ID' --build-name '$ZIP_BUILD' --build-number 1 -v"
"$JF" rt build-publish "$ZIP_BUILD" 1 --server-id "$SERVER_ID" >/dev/null 2>&1 || true
echo

# --- 3. FEATURE: Checksum cross-verification ----------------------------------------------------
log "[3. checksum] downloading published $OWNER/simple@$SIMPLE_VERSION and hashing it locally"
SIMPLE_ARTIFACT_PATH="$REPO/$OWNER/simple/simple-$SIMPLE_VERSION.zip"
if dl_out="$("$JF" rt download "$SIMPLE_ARTIFACT_PATH" "$WORKDIR/downloaded-simple.zip" --flat --server-id "$SERVER_ID" 2>&1)"; then
  LOCAL_SHA256="$(shasum -a 256 "$WORKDIR/downloaded-simple.zip" | awk '{print $1}')"
  RECORDED_SHA256="$("$JF" rt curl -s -XGET "/api/build/smoke-publish-simple-${STAMP}/1" --server-id "$SERVER_ID" \
    | jq -r '.buildInfo.modules[0].artifacts[0].sha256' 2>/dev/null)"
  if [[ -n "$LOCAL_SHA256" && "$LOCAL_SHA256" == "$RECORDED_SHA256" ]]; then
    ok "downloaded-file SHA256 ($LOCAL_SHA256) matches build-info-recorded SHA256"
  else
    fail "checksum mismatch: local=$LOCAL_SHA256 build-info=$RECORDED_SHA256"
  fi
else
  fail "could not download $SIMPLE_ARTIFACT_PATH for checksum comparison: $dl_out"
fi
echo

# --- 4. FEATURE: Build-info properties (agentpackages.* stamped on the published artifact) ------
log "[4. build properties] agentpackages.* properties on the published artifact"
"$JF" rt curl -s -XGET "/api/storage/$SIMPLE_ARTIFACT_PATH?properties" --server-id "$SERVER_ID" \
  | jq '.properties | with_entries(select(.key | startswith("agentpackages")))' 2>/dev/null | sed 's/^/    /'
echo

# --- 5. FEATURE: Install (resolves the full chain) ----------------------------------------------
INSTALL_DIR="$WORKDIR/install-app"
rm -rf "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR"
INSTALL_BUILD="smoke-install-${STAMP}"
log "[5. install] $OWNER/transitive#$TRANSITIVE_VERSION into $INSTALL_DIR (build $INSTALL_BUILD/1)"
run_expect_success "install $OWNER/transitive#$TRANSITIVE_VERSION" -- bash -c \
  "cd '$INSTALL_DIR' && '$JF' agent apm install '$OWNER/transitive#$TRANSITIVE_VERSION' --target claude \
     --server-id '$SERVER_ID' --build-name '$INSTALL_BUILD' --build-number 1 -v"
log "[5b. build-publish] jf rt bp $INSTALL_BUILD 1 --server-id $SERVER_ID"
"$JF" rt build-publish "$INSTALL_BUILD" 1 --server-id "$SERVER_ID" 2>&1 | sed 's/^/    /'
log "    installed dependency tree:"
find "$INSTALL_DIR/apm_modules" -maxdepth 2 -mindepth 1 2>/dev/null | sed 's/^/      /' || true
echo

log "[5c. install error handling] installing a nonexistent package should fail cleanly"
INVALID_DIR="$WORKDIR/install-invalid"
mkdir -p "$INVALID_DIR"
run_expect_failure "install $OWNER/does-not-exist-$STAMP#9.9.9" -- bash -c \
  "cd '$INVALID_DIR' && '$JF' agent apm install '$OWNER/does-not-exist-$STAMP#9.9.9' --server-id '$SERVER_ID'"
echo

log "[5d. install --frozen] re-running install --frozen in an already-installed, unchanged dir"
run_expect_success "install --frozen (lockfile already in sync)" -- bash -c \
  "cd '$INSTALL_DIR' && '$JF' agent apm install --frozen --server-id '$SERVER_ID'"
echo

log "[5e. install --dev] installing $OWNER/simple as a dev dependency"
DEV_DIR="$WORKDIR/install-dev"
rm -rf "$DEV_DIR"; mkdir -p "$DEV_DIR"
run_expect_success "install --dev $OWNER/simple#$SIMPLE_VERSION" -- bash -c \
  "cd '$DEV_DIR' && '$JF' agent apm install --dev '$OWNER/simple#$SIMPLE_VERSION' --target claude --server-id '$SERVER_ID'"
grep -A2 '^  apm:' "$DEV_DIR/apm.yml" 2>/dev/null | sed 's/^/      /' || true
echo

log "[5f. install --module] build-info module scoping"
MODULE_BUILD="smoke-module-${STAMP}"
MODULE_DIR="$WORKDIR/install-module"
rm -rf "$MODULE_DIR"; mkdir -p "$MODULE_DIR"
run_expect_success "install --module=smoke-module" -- bash -c \
  "cd '$MODULE_DIR' && '$JF' agent apm install '$OWNER/simple#$SIMPLE_VERSION' --target claude \
     --server-id '$SERVER_ID' --module smoke-module --build-name '$MODULE_BUILD' --build-number 1"
"$JF" rt build-publish "$MODULE_BUILD" 1 --server-id "$SERVER_ID" >/dev/null 2>&1 || true
MODULE_ID="$("$JF" rt curl -s -XGET "/api/build/$MODULE_BUILD/1" --server-id "$SERVER_ID" | jq -r '.buildInfo.modules[0].id' 2>/dev/null)"
if [[ "$MODULE_ID" == "smoke-module" ]]; then
  ok "build-info module id is 'smoke-module' as requested"
else
  fail "expected module id 'smoke-module', got '$MODULE_ID'"
fi
echo

# --- 6. FEATURE: Build flags required (both-or-neither, validated upfront) -----------------------
log "[6. build flags] --build-name without --build-number should be rejected upfront"
PARTIAL_DIR="$WORKDIR/install-partial-buildflags"
rm -rf "$PARTIAL_DIR"; mkdir -p "$PARTIAL_DIR"
PARTIAL_BUILD="smoke-partial-${STAMP}"
run_expect_failure "install with only --build-name set (no --build-number)" -- bash -c \
  "cd '$PARTIAL_DIR' && '$JF' agent apm install '$OWNER/simple#$SIMPLE_VERSION' --target claude \
     --server-id '$SERVER_ID' --build-name '$PARTIAL_BUILD'"
echo

# --- 7. FEATURE: Native-flag / --server-id passthrough -------------------------------------------
log "[7. passthrough] native --dry-run flag forwarded straight to apm (no re-install happens)"
( cd "$INSTALL_DIR" && "$JF" agent apm install --dry-run --server-id "$SERVER_ID" ) 2>&1 | sed 's/^/    /'
log "[7b. passthrough] a non-install/publish apm command forwarded through jf: 'apm outdated'"
( cd "$INSTALL_DIR" && "$JF" agent apm outdated --server-id "$SERVER_ID" ) 2>&1 | sed 's/^/    /' || true
echo

# --- 8. FEATURE: Native apm CLI directly, reusing jf-setup-written credentials ------------------
log "[8. native cli] bare 'apm install' (no jf wrapper) reusing ~/.apm/config.json credentials"
NATIVE_DIR="$WORKDIR/install-native"
rm -rf "$NATIVE_DIR"; mkdir -p "$NATIVE_DIR"
run_expect_success "apm install (native, no jf)" -- bash -c \
  "cd '$NATIVE_DIR' && apm install '$OWNER/simple#$SIMPLE_VERSION' --target claude"
echo

# --- 9. FEATURE: Auth types — basic-auth-as-configured (implicit above) + a minted access token --
log "[9. auth] minting a short-lived access token and authenticating a command with it"
log "    (the other auth type — basic-auth-as-$SERVER_ID-is-configured — is exercised implicitly"
log "     by every other command in this script, per TestApmAuthWithUsernamePassword's own finding"
log "     that jf auto-mints a refreshable token from basic auth under the hood)"
if TOKEN_JSON="$("$JF" access-token-create --server-id "$SERVER_ID" --expiry=600 \
    --description="apm-smoke-test ${STAMP}" --format=json 2>&1)"; then
  ACCESS_TOKEN="$(echo "$TOKEN_JSON" | jq -r '.access_token // .token // empty')"
  if [[ -n "$ACCESS_TOKEN" ]]; then
    "$JF" config add "$TMP_SERVER_ID" --url="$PLATFORM_URL" --access-token="$ACCESS_TOKEN" --interactive=false >/dev/null
    run_expect_success "apm outdated authenticated via minted access token" -- \
      "$JF" agent apm outdated --server-id "$TMP_SERVER_ID"
    "$JF" config rm "$TMP_SERVER_ID" --quiet >/dev/null 2>&1
  else
    fail "could not parse an access token out of: $TOKEN_JSON"
  fi
elif echo "$TOKEN_JSON" | grep -qi "authenticating with access token is currently mandatory"; then
  skip "access-token-create: this platform requires an existing access-token session to mint" \
       "one (a policy restriction, not an apm bug) — $SERVER_ID is basic-auth-configured, so" \
       "this auth-type variant can't be exercised without a token already in hand"
else
  fail "access-token-create failed: $TOKEN_JSON"
fi
echo

# --- 10. FEATURE: Build-info readback -----------------------------------------------------------
log "[10. build-info readback] $INSTALL_BUILD/1 — dependency checksums as published"
"$JF" rt curl -s -XGET "/api/build/$INSTALL_BUILD/1" --server-id "$SERVER_ID" \
  | jq '.buildInfo.modules[].dependencies[]? | {id, sha1, sha256, md5}' 2>/dev/null \
  | sed 's/^/    /'
echo

cat <<EOF
==> Done. $FAILURES check(s) failed.

    Published : $OWNER/simple@$SIMPLE_VERSION, $OWNER/consumer@$CONSUMER_VERSION,
                 $OWNER/transitive@$TRANSITIVE_VERSION, $OWNER/transitive@$ZIP_VERSION (via --zip)
                 (in $REPO on server '$SERVER_ID' — not deleted by this script)
    Builds     : smoke-publish-{simple,consumer,transitive}-${STAMP}/1, $ZIP_BUILD/1,
                 $INSTALL_BUILD/1, $MODULE_BUILD/1, $PARTIAL_BUILD (never published)
    Local files: $WORKDIR $([[ "$KEEP" == true ]] && echo "(kept)" || echo "(will be removed)")

Clean up the published packages/builds yourself when done, e.g.:
  jf rt delete "$REPO/$OWNER/{simple,consumer,transitive}/*" --server-id=$SERVER_ID --quiet
  jf rt build-discard smoke-publish-simple-${STAMP} --max-builds=0 --server-id=$SERVER_ID
EOF

[[ "$FAILURES" -eq 0 ]] || exit 1
