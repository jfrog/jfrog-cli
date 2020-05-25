#!/bin/bash

exe_name="jfrog"
flags="-w -extldflags \"-static\""

# Executable name provided
if [ $# -ge 1 ]
  then
	exe_name="$1"
fi

# Encryption file provided as well
if [ $# -gt 1 ]
  then
  # Read encryption file, remove new lines and spaces
  encryptionFileContent=$(<"$2")
  encryptionFileContent="${encryptionFileContent//$'\n'/ }"
  encryptionFileContent="${encryptionFileContent//$' '/}"
  flags="-X 'github.com/jfrog/jfrog-cli/utils/config.EncryptionFile=${encryptionFileContent}' ${flags}"
fi

CGO_ENABLED=0 go build -o $exe_name -ldflags "${flags}" main.go