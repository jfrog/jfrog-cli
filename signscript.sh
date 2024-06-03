#!/bin/bash

# Get input parameters from ENV
APPLE_CERT_DATA=$APPLE_CERT_DATA
APPLE_CERT_PASSWORD=$APPLE_CERT_PASSWORD
APPLE_TEAM_ID=$APPLE_TEAM_ID

# Validate input parameters
if [ -z "$APPLE_CERT_DATA" ] ; then
    echo "Error: Missing input APPLE_CERT_DATA parameters."
    exit 1
fi

if [ -z "$APPLE_CERT_PASSWORD" ] ; then
    echo "Error: Missing input APPLE_CERT_PASSWORD parameters."
    exit 1
fi
if  [ -z "$APPLE_TEAM_ID" ]; then
    echo "Error: Missing input   APPLE_TEAM_ID parameters."
    exit 1
fi

# Set temp directory
RUNNER_TEMP="/Users/runner/work/_temp"

echo "Saving Certificate to temp files"
echo "$APPLE_CERT_DATA" | base64 --decode > $RUNNER_TEMP/certs.p12


echo "Creating keychains..."
security create-keychain -p "$APPLE_CERT_PASSWORD" macos-build.keychain
security default-keychain -s macos-build.keychain
security unlock-keychain -p "$APPLE_CERT_PASSWORD" macos-build.keychain
security set-keychain-settings -t 3600 -u macos-build.keychain


echo "Certificate into keychain..."
# Import certs to keychain
security import /Users/runner/work/_temp/certs.p12 -k ~/Library/Keychains/macos-build.keychain -P "$APPLE_CERT_PASSWORD" -T /usr/bin/codesign

echo "verifying identity..."
# Verify keychain things
security find-identity -p codesigning -v

echo "unlocking the key"
security unlock-keychain -p "$APPLE_CERT_PASSWORD" macos-build.keychain
security set-key-partition-list -S apple-tool:,apple:, -s -k "$APPLE_CERT_PASSWORD" -D "$APPLE_TEAM_ID" -t private  macos-build.keychain


echo "Sign the binary..."
codesign -s "$APPLE_TEAM_ID" --force jfrog-cli

echo "Verify binary is signed"
codesign -vd ./jfrog-cli