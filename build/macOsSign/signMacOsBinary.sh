#!/bin/bash

# Assign environment variables to local variables
# Base64 encoded certificate data
APPLE_CERT_DATA=$APPLE_CERT_DATA
# Passphrase used to open the certificate
APPLE_CERT_PASSWORD=$APPLE_CERT_PASSWORD
# Apple Developer Team ID
APPLE_TEAM_ID=$APPLE_TEAM_ID
# Set temp dir as runner temp dir
TEMP_DIR=$RUNNER_TEMP

# Validate input parameters
if [ -z "$APPLE_CERT_DATA" ] || [ -z "$APPLE_CERT_PASSWORD" ] || [ -z "$APPLE_TEAM_ID" ] ; then
    echo "Error: Missing environment variable."
    exit 1
fi

# Save the decoded certificate data to a temporary file
echo "Saving Certificate to temp files"
echo "$APPLE_CERT_DATA" | base64 --decode > "$TEMP_DIR"/certs.p12

# Create a new keychain and set it as the default
echo "Creating keychains..."
security create-keychain -p "$APPLE_CERT_PASSWORD" macos-build.keychain
security default-keychain -s macos-build.keychain
security unlock-keychain -p "$APPLE_CERT_PASSWORD" macos-build.keychain
security set-keychain-settings -t 3600 -u macos-build.keychain

# Import the certificate into the keychain
echo "Importing certificate into keychain..."
security import "$TEMP_DIR"/certs.p12 -k ~/Library/Keychains/macos-build.keychain -P "$APPLE_CERT_PASSWORD" -T /usr/bin/codesign

# Verify the identity in the keychain
echo "Verifying identity..."
security find-identity -p codesigning -v

# Unlock the keychain to allow signing in terminal without asking for password
echo "Unlocking the keychain"
security unlock-keychain -p "$APPLE_CERT_PASSWORD" macos-build.keychain
security set-key-partition-list -S apple-tool:,apple:, -s -k "$APPLE_CERT_PASSWORD" -D "$APPLE_TEAM_ID" -t private  macos-build.keychain

# Sign the binary
echo "Signing the binary..."
codesign -s "$APPLE_TEAM_ID" --force jfrog-cli

# Verify the binary is signed
echo "Verifying binary is signed"
codesign -vd ./jfrog-cli