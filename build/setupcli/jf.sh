#!/bin/sh
set -u

# This script is downloading the OS-specific JFrog CLI binary with the name - 'jf', adds it to PATH and prints greeting message

CLI_OS="na"
CLI_MAJOR_VER="v2-jf"
VERSION="[RELEASE]"
FILE_NAME="jf"
SETUP_COMMAND="jf setup"
GREEN_COLOR='\033[0;32m'
REMOVE_COLOR='\033[0m'

print_installation_greeting () {
  echo "${GREEN_COLOR}Thank you for installing JFrog CLI! 🐸 ${REMOVE_COLOR}"
}

if [ $# -eq 1 ]; then
    SETUP_COMMAND="$SETUP_COMMAND $1"
fi

echo "Downloading the latest version of JFrog CLI..."
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

URL="https://releases.jfrog.io/artifactory/jfrog-cli/${CLI_MAJOR_VER}/${VERSION}/jfrog-cli-${CLI_OS}-${ARCH}/${FILE_NAME}"
echo "Downloading from: $URL"
curl -XGET "$URL" -L -k -g > $FILE_NAME
chmod +x $FILE_NAME

# Move executable to a destination in path.
# Order is by destination priority.
set -- "/usr/local/bin" "/usr/bin" "/opt/bin"
while [ -n "$1" ]; do
    # Check if destination is in path.
    if echo "$PATH"|grep "$1" -> /dev/null ; then
         if mv $FILE_NAME "$1" ; then
            echo ""
            echo "The $FILE_NAME executable was installed in $1"
            print_installation_greeting
            $SETUP_COMMAND "$BASE64_CRED"
            exit 0
        else
            echo ""
            echo "We'd like to install the JFrog CLI executable in $1. Please approve this installation by entering your password."
            if sudo mv $FILE_NAME "$1" ; then
                echo ""
                echo "The $FILE_NAME executable was installed in $1"
                print_installation_greeting
                $SETUP_COMMAND "$BASE64_CRED"
                exit 0
            fi
        fi
    fi
    shift
done

echo "could not find supported destination path in \$PATH"
exit 1
