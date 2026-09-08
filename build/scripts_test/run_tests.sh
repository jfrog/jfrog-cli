#!/bin/sh
set -eu

# Dependency-free unit tests for the install-script checksum verification
# added in JGC-543. Drives each script against local file:// fixtures built
# on the fly, in a scratch directory, for whatever OS/ARCH this machine
# actually is (mirroring each script's own uname-based detection) — so these
# tests produce real pass/fail evidence on any dev machine and in CI alike,
# without network access and without depending on JGC-542's checksums being
# published yet.

REPO_ROOT=$(cd "$(dirname "$0")/../.." && pwd)
VERSION="1.2.3"
FAILURES=0

# Mirror the CLI_OS/ARCH detection every install script performs via uname,
# so generated fixtures match whatever this test actually runs on.
if uname -s | grep -q -E -i "(cygwin|mingw|msys|windows)"; then
    CLI_OS="windows"
    ARCH="amd64"
elif uname -s | grep -q -i "darwin"; then
    CLI_OS="mac"
    if [ "$(uname -m)" = "arm64" ]; then
      ARCH="arm64"
    else
      ARCH="386"
    fi
else
    CLI_OS="linux"
    MACHINE_TYPE="$(uname -m)"
    case $MACHINE_TYPE in
        i386 | i486 | i586 | i686 | i786 | x86)
            ARCH="386"
            ;;
        amd64 | x86_64 | x64)
            ARCH="amd64"
            ;;
        arm | armv7l)
            ARCH="arm"
            ;;
        aarch64)
            ARCH="arm64"
            ;;
        s390x)
            ARCH="s390x"
            ;;
        ppc64)
           ARCH="ppc64"
           ;;
        ppc64le)
           ARCH="ppc64le"
           ;;
        *)
            echo "Unknown machine type: $MACHINE_TYPE" >&2
            exit 1
            ;;
    esac
fi

if command -v sha256sum >/dev/null 2>&1; then
    sha256_of() { sha256sum "$1" | awk '{print $1}'; }
elif command -v shasum >/dev/null 2>&1; then
    sha256_of() { shasum -a 256 "$1" | awk '{print $1}'; }
else
    echo "Neither sha256sum nor shasum is available; cannot run these tests." >&2
    exit 1
fi

# make_fixture_tree FIXTURE_ROOT CLI_MAJOR_VER FILE_NAME CASE FIXTURE_VERSION
#   Builds one release-tree fixture under $FIXTURE_ROOT for the given case:
#     good              - correct checksum, as a bare 64-char hex digest (no
#                         filename column, no trailing newline) in a
#                         per-binary .sha256 sidecar, matching the real
#                         releases.jfrog.io response
#     mismatch          - well-formed but wrong checksum (same bare-digest
#                         shape as "good")
#     missing-manifest  - no .sha256 sidecar at all
#     missing-entry     - sidecar exists but is empty (empty/malformed checksum)
#   FIXTURE_VERSION is the literal directory name to build the release tree
#   under (normally $VERSION, but build/setupcli/jf.sh never reads $1 into
#   VERSION - it always looks under the literal "[RELEASE]" - so callers for
#   that script must pass "[RELEASE]" here instead).
make_fixture_tree() {
    fixture_root=$1
    cli_major_ver=$2
    file_name=$3
    case_name=$4
    fixture_version=$5

    rel_dir="jfrog-cli-${CLI_OS}-${ARCH}"
    version_dir="$fixture_root/$cli_major_ver/$fixture_version"
    mkdir -p "$version_dir/$rel_dir"
    printf 'dummy-%s-binary\n' "$file_name" > "$version_dir/$rel_dir/$file_name"

    sidecar="$version_dir/$rel_dir/$file_name.sha256"

    case "$case_name" in
        good)
            hash=$(sha256_of "$version_dir/$rel_dir/$file_name")
            printf '%s' "$hash" > "$sidecar"
            ;;
        mismatch)
            printf '%s' "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef" > "$sidecar"
            ;;
        missing-manifest)
            : # no .sha256 sidecar written - simulates a 404
            ;;
        missing-entry)
            : > "$sidecar" # sidecar exists but is empty - simulates empty/malformed checksum
            ;;
        *)
            echo "unknown fixture case: $case_name" >&2
            exit 1
            ;;
    esac
}

# run_case SCRIPT CLI_MAJOR_VER FILE_NAME CASE EXPECT
#   SCRIPT         path to the install script, relative to repo root
#   CLI_MAJOR_VER  "v2-jf" or "v2" (matches the script's own CLI_MAJOR_VER)
#   FILE_NAME      "jf" or "jfrog" (matches the script's own FILE_NAME)
#   CASE           good | mismatch | missing-manifest | missing-entry
#   EXPECT         "pass" or "fail" - on "fail", also asserts the script's
#                  scratch run directory was left empty (verification always
#                  rm -f's the partial binary before exiting, for every
#                  script, so this holds regardless of the binary's name)
run_case() {
    script=$1
    cli_major_ver=$2
    file_name=$3
    case_name=$4
    expect=$5

    # build/setupcli/jf.sh never assigns $1 to VERSION - it always builds its
    # URLs under the literal "[RELEASE]" path segment, and only uses $1 (if
    # given) to build the interactive `jf setup` command it execs on success.
    # So its fixture must be built under "[RELEASE]", and the script must be
    # invoked with no positional argument at all.
    case "$script" in
        */setupcli/*)
            fixture_version="[RELEASE]"
            script_arg=""
            ;;
        *)
            fixture_version="$VERSION"
            script_arg="$VERSION"
            ;;
    esac

    work_dir=$(mktemp -d)
    fixture_root="$work_dir/fixture"
    mkdir -p "$fixture_root"
    make_fixture_tree "$fixture_root" "$cli_major_ver" "$file_name" "$case_name" "$fixture_version"

    run_dir="$work_dir/run"
    mkdir -p "$run_dir"

    status=0
    (
        cd "$run_dir"
        # shellcheck disable=SC2086 # intentionally unquoted: empty means "no
        # positional arg at all" for setupcli, not an empty-string arg.
        JFROG_CLI_RELEASES_BASE_URL="file://$fixture_root" \
            sh "$REPO_ROOT/$script" $script_arg </dev/null
    ) || status=$?

    label="$script [$case_name]"
    if [ "$expect" = "pass" ] && [ "$status" -ne 0 ]; then
        echo "FAIL: $label expected exit 0, got $status"
        FAILURES=$((FAILURES + 1))
    elif [ "$expect" = "fail" ] && [ "$status" -eq 0 ]; then
        echo "FAIL: $label expected non-zero exit, got 0"
        FAILURES=$((FAILURES + 1))
    elif [ "$expect" = "fail" ] && [ -n "$(ls -A "$run_dir" 2>/dev/null)" ]; then
        echo "FAIL: $label left files behind after a failed verification: $(ls -A "$run_dir")"
        FAILURES=$((FAILURES + 1))
    else
        echo "PASS: $label"
    fi

    rm -rf "$work_dir"
}

# --- build/getcli/jf.sh (added in this task) ---
run_case build/getcli/jf.sh v2-jf jf good pass
run_case build/getcli/jf.sh v2-jf jf mismatch fail
run_case build/getcli/jf.sh v2-jf jf missing-manifest fail
run_case build/getcli/jf.sh v2-jf jf missing-entry fail

# --- build/getcli/jfrog.sh ---
run_case build/getcli/jfrog.sh v2 jfrog good pass
run_case build/getcli/jfrog.sh v2 jfrog mismatch fail
run_case build/getcli/jfrog.sh v2 jfrog missing-manifest fail
run_case build/getcli/jfrog.sh v2 jfrog missing-entry fail

# --- build/installcli/jf.sh (negative paths only - see task notes) ---
run_case build/installcli/jf.sh v2-jf jf mismatch fail
run_case build/installcli/jf.sh v2-jf jf missing-manifest fail
run_case build/installcli/jf.sh v2-jf jf missing-entry fail

# --- build/installcli/jfrog.sh (negative paths only) ---
run_case build/installcli/jfrog.sh v2 jfrog mismatch fail
run_case build/installcli/jfrog.sh v2 jfrog missing-manifest fail
run_case build/installcli/jfrog.sh v2 jfrog missing-entry fail

# --- build/setupcli/jf.sh (negative paths only) ---
run_case build/setupcli/jf.sh v2-jf jf mismatch fail
run_case build/setupcli/jf.sh v2-jf jf missing-manifest fail
run_case build/setupcli/jf.sh v2-jf jf missing-entry fail

if [ "$FAILURES" -ne 0 ]; then
    echo "$FAILURES case(s) failed."
    exit 1
fi
echo "All cases passed."
