#!/bin/bash

# Assign environment variables to local variables
# Set temp dir as runner temp dir
TEMP_DIR=$RUNNER_TEMP
KEYCHAIN_NAME="macos-build.keychain"

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
security create-keychain -p "$APPLE_CERT_PASSWORD" $KEYCHAIN_NAME
security default-keychain -s $KEYCHAIN_NAME
security unlock-keychain -p "$APPLE_CERT_PASSWORD" $KEYCHAIN_NAME
security set-keychain-settings -t 3600 -u $KEYCHAIN_NAME

# Import the certificate into the keychain
echo "Importing certificate into keychain..."
security import "$TEMP_DIR"/certs.p12 -k ~/Library/Keychains/$KEYCHAIN_NAME -P "$APPLE_CERT_PASSWORD" -T /usr/bin/codesign

# Verify the identity in the keychain
echo "Verifying identity..."
security find-identity -p codesigning -v

# Unlock the keychain to allow signing in terminal without asking for password
echo "Unlocking the keychain"
security unlock-keychain -p "$APPLE_CERT_PASSWORD" $KEYCHAIN_NAME
security set-key-partition-list -S apple-tool:,apple:, -s -k "$APPLE_CERT_PASSWORD" -D "$APPLE_TEAM_ID" -t private  $KEYCHAIN_NAME

# Sign the binary
echo "Signing the binary..."
codesign -s "$APPLE_TEAM_ID" --force jfrog-cli

# Verify the binary is signed
echo "Verifying binary is signed"
codesign -vd ./jfrog-cli

# Cleanup
echo "Deleting keychain.."
security delete-keychain $KEYCHAIN_NAME
echo "Delete Certificate..."
rm -rf "$TEMP_DIR"/certs.p12
