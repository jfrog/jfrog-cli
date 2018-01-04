#!/bin/bash

CLI_OS="na"
CLI_UNAME="na"

if $(echo "${OSTYPE}" | grep -q msys); then
    CLI_OS="windows"
    URL="https://api.bintray.com/content/jfrog/jfrog-cli-go/\$latest/jfrog-cli-windows-amd64/jfrog.exe?bt_package=jfrog-cli-windows-amd64"
    FILE_NAME="jfrog.exe"
elif $(echo "${OSTYPE}" | grep -q darwin); then
    CLI_OS="mac"
    URL="https://api.bintray.com/content/jfrog/jfrog-cli-go/\$latest/jfrog-cli-mac-386/jfrog?bt_package=jfrog-cli-mac-386"
    FILE_NAME="jfrog"
else
    CLI_OS="linux"
    if $(uname -m | grep -q 64); then
        CLI_UNAME="64"
        URL="https://api.bintray.com/content/jfrog/jfrog-cli-go/\$latest/jfrog-cli-linux-amd64/jfrog?bt_package=jfrog-cli-linux-amd64"
        FILE_NAME="jfrog"
    else
        CLI_UNAME="32"
        URL="https://api.bintray.com/content/jfrog/jfrog-cli-go/\$latest/jfrog-cli-linux-386/jfrog?bt_package=jfrog-cli-linux-386"
        FILE_NAME="jfrog"
    fi
fi

curl -XGET "$URL" -L -k > $FILE_NAME
chmod u+x $FILE_NAME