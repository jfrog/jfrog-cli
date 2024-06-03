#!/bin/bash


# The first argument is assigned to APPLE_CERT_DATA
APPLE_CERT_DATA=$1

# The second argument is assigned to APPLE_CERT_PASSWORD
APPLE_CERT_PASSWORD=$2

# The third argument is assigned to APPLE_TEAM_ID
APPLE_TEAM_ID=$3

# Export certs
echo "saving cert data to /tmp/certs.p12"
echo "$APPLE_CERT_DATA" | base64 --decode > ~/_tmp/certs.p12

echo "checking p12"
ls -la ~/_tmp | grep p12

echo "Creating keyhcains..."
# Create keychain
security create-keychain -p "$APPLE_CERT_PASSWORD" macos-build.keychain
security default-keychain -s macos-build.keychain
security unlock-keychain -p "$APPLE_CERT_PASSWORD" macos-build.keychain
security set-keychain-settings -t 3600 -u macos-build.keychain

echo "check keychains content"
# Check keychain content
ls -la ~/Library/Keychains




echo "importing.."
# Import certs to keychain
security import /tmp/certs.p12 -k ~/Library/Keychains/macos-build.keychain -P "$APPLE_CERT_PASSWORD" -T /usr/bin/codesign

# Verify keychain things
security find-identity -p codesigning -v

# Force the codesignature
codesign -s "$APPLE_TEAM_ID" -f jfrog-cli

codesign -vd ./jfrog-cli