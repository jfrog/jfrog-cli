#!/bin/sh
set -u

# This script is downloading the OS-specific JFrog CLI binary with the name - 'jfrog', and adds it to PATH

CLI_OS="na"
CLI_MAJOR_VER="v2"
VERSION="[RELEASE]"
FILE_NAME="jfrog"

if [ $# -eq 1 ]; then
    VERSION=$1
    echo "Downloading version $VERSION of JFrog CLI..."
else
    echo "Downloading the latest version of JFrog CLI..."
fi
echo ""

if uname -s | grep -q -E -i "(cygwin|mingw|msys|windows)"; then
    CLI_OS="windows"
    ARCH="amd64"
    FILE_NAME="${FILE_NAME}.exe"
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
            echo "Unknown machine type: $MACHINE_TYPE"
            exit 1
            ;;
    esac
fi

BASE_URL="${JFROG_CLI_RELEASES_BASE_URL:-https://releases.jfrog.io/artifactory/jfrog-cli}"
URL="${BASE_URL}/${CLI_MAJOR_VER}/${VERSION}/jfrog-cli-${CLI_OS}-${ARCH}/${FILE_NAME}"
echo "Downloading from: $URL"
curl -XGET "$URL" -L -g -o "$FILE_NAME"

# Verify the download against the SHA256 checksum sidecar published for this
# binary (JGC-542) before it is ever chmod +x'd or moved onto PATH. The
# checksum lives at the exact same path as the binary, with .sha256 appended
# (a per-binary sidecar, not a shared manifest - see JGC-542's corrected design).
CHECKSUM_URL="${URL}.sha256"
CHECKSUM_TMP="${FILE_NAME}.sha256.tmp"

echo "Verifying checksum against: $CHECKSUM_URL"
if ! curl -sS --fail -L -g "$CHECKSUM_URL" -o "$CHECKSUM_TMP"; then
    echo "ERROR: could not download the checksum from $CHECKSUM_URL" >&2
    echo "  (downloaded binary: $URL)" >&2
    echo "Refusing to install an unverified $FILE_NAME binary." >&2
    rm -f "$FILE_NAME" "$CHECKSUM_TMP"
    exit 1
fi

EXPECTED_SHA256=$(awk 'NR==1 { print $1 }' "$CHECKSUM_TMP")
rm -f "$CHECKSUM_TMP"

if [ -z "$EXPECTED_SHA256" ]; then
    echo "ERROR: empty or malformed checksum at $CHECKSUM_URL" >&2
    echo "  (downloaded binary: $URL)" >&2
    echo "Refusing to install an unverified $FILE_NAME binary." >&2
    rm -f "$FILE_NAME"
    exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_SHA256=$(sha256sum "$FILE_NAME" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
    ACTUAL_SHA256=$(shasum -a 256 "$FILE_NAME" | awk '{print $1}')
else
    echo "ERROR: neither sha256sum nor shasum is available to verify the download." >&2
    rm -f "$FILE_NAME"
    exit 1
fi

if [ "$ACTUAL_SHA256" != "$EXPECTED_SHA256" ]; then
    echo "ERROR: checksum mismatch for $FILE_NAME downloaded from $URL" >&2
    echo "  expected (from $CHECKSUM_URL): $EXPECTED_SHA256" >&2
    echo "  actual:                        $ACTUAL_SHA256" >&2
    rm -f "$FILE_NAME"
    exit 1
fi

echo "Checksum verified ($ACTUAL_SHA256)."
chmod +x "$FILE_NAME"

# Move executable to a destination in path.
# Order is by destination priority.
set -- "/usr/local/bin" "/usr/bin" "/opt/bin"
while [ -n "$1" ]; do
    # Check if destination is in path.
    if echo "$PATH"|grep "$1" -> /dev/null ; then
        if mv $FILE_NAME "$1" ; then
            echo ""
            echo "The $FILE_NAME executable was installed in $1"
            exit 0
        else
            echo ""
            echo "We'd like to install the JFrog CLI executable in $1. Please approve this installation by entering your password."
            if sudo mv $FILE_NAME "$1" ; then
                echo ""
                echo "The $FILE_NAME executable was installed in $1"
                exit 0
            fi
        fi
    fi
    shift
done

echo "could not find supported destination path in \$PATH"
exit 1
