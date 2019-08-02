#!/bin/bash

CLI_OS="na"
CLI_UNAME="na"

if [ $# -eq 0 ]
  then
	VERSION=\$latest
	echo "Downloading the latest version of JFrog CLI..."
  else
	VERSION=$1
	echo "Downloading version $1 of JFrog CLI..."
fi

if $(echo "${OSTYPE}" | grep -q msys); then
    CLI_OS="windows"
    URL="https://api.bintray.com/content/jfrog/jfrog-cli-go/${VERSION}/jfrog-cli-windows-amd64/jfrog.exe?bt_package=jfrog-cli-windows-amd64"
    FILE_NAME="jfrog.exe"
elif $(echo "${OSTYPE}" | grep -q darwin); then
    CLI_OS="mac"
    URL="https://api.bintray.com/content/jfrog/jfrog-cli-go/${VERSION}/jfrog-cli-mac-386/jfrog?bt_package=jfrog-cli-mac-386"
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
        *)
            echo "Unknown machine type: $MACHINE_TYPE"
            exit -1
            ;;
    esac
    URL="https://api.bintray.com/content/jfrog/jfrog-cli-go/${VERSION}/jfrog-cli-${CLI_OS}-${ARCH}/jfrog?bt_package=jfrog-cli-${CLI_OS}-${ARCH}"
    FILE_NAME="jfrog"
fi

curl -XGET "$URL" -L -k > $FILE_NAME
chmod u+x $FILE_NAME
