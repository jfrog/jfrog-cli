#!/bin/bash

CLI_OS="na"
CLI_UNAME="na"
CLI_MAJOR_VER="v2"
VERSION="[RELEASE]"

if [ $# -eq 1 ]
then
    VERSION=$1
    echo "Downloading version $VERSION of JFrog CLI..."
else
    echo "Downloading the latest version of JFrog CLI..."
fi

if $(echo "${OSTYPE}" | grep -q msys); then
    CLI_OS="windows"
    URL="https://releases.jfrog.io/artifactory/jfrog-cli/${CLI_MAJOR_VER}/${VERSION}/jfrog-cli-windows-amd64/jfrog.exe"
    FILE_NAME="jfrog.exe"
elif $(echo "${OSTYPE}" | grep -q darwin); then
    CLI_OS="mac"
    if [[ $(uname -m) == 'arm64' ]]; then
      URL="https://releases.jfrog.io/artifactory/jfrog-cli/${CLI_MAJOR_VER}/${VERSION}/jfrog-cli-mac-arm64/jfrog"
    else
      URL="https://releases.jfrog.io/artifactory/jfrog-cli/${CLI_MAJOR_VER}/${VERSION}/jfrog-cli-mac-386/jfrog"
    fi
    FILE_NAME="jfrog"
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
            exit -1
            ;;
    esac
    URL="https://releases.jfrog.io/artifactory/jfrog-cli/${CLI_MAJOR_VER}/${VERSION}/jfrog-cli-${CLI_OS}-${ARCH}/jfrog"
    FILE_NAME="jfrog"
fi

curl -XGET "$URL" -L -k -g > $FILE_NAME
chmod u+x $FILE_NAME