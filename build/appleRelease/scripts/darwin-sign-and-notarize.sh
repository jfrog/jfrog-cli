#!/bin/bash

# Script Purpose: This script automates the process of signing and notarizing a macOS binary.
# It leverages specific environment variables to access necessary Apple credentials and the app template path.
# The script ensures the binary conforms to Apple's requirements for notarization,
# including correct placement within the app bundle and proper signing with a developer certificate.
#
# Prerequisites:
# App Bundle Structure Requirement:
# The .app bundle must have a specific structure for the script to successfully sign and notarize the binary.
# This structure is crucial for the app's acceptance by macOS and includes:
# YOUR_APP.app/
# ├── Contents/
# │   ├── MacOS/
# │   │   └── YOUR_APP (This is the executable file that will be signed and notarized)
# │   └── Info.plist (Contains metadata and configurations for the app)
#
# Input:
# - APPLE_CERT_DATA: Base64 encoded data of the Apple Developer certificate.
# - APPLE_CERT_PASSWORD: Password for the Apple Developer certificate.
# - APPLE_TEAM_ID: Identifier for the Apple Developer Team.
# - APPLE_ACCOUNT_ID: Apple Developer Account ID.
# - APPLE_APP_SPECIFIC_PASSWORD: Password for app-specific services on the Apple Developer Account.
# - APP_TEMPLATE_PATH: Path to the .app bundle template, you created in the App Bundle Structure Requirement prerequisite.
#
# Output:
# Upon successful execution, the script outputs a signed and notarized binary file in the current directory, ready for distribution.

# Validates the structure of the app template directory.
validate_app_template_structure() {
    if [ ! -d "$APP_TEMPLATE_PATH" ]; then
        echo "Error: $APP_TEMPLATE_PATH directory does not exist."
        return 1
    fi

    if [ ! -d "$APP_TEMPLATE_PATH/Contents" ]; then
        echo "Error: Contents directory does not exist in $APP_TEMPLATE_PATH."
        return 1
    fi

    if [ ! -d "$APP_TEMPLATE_PATH/Contents/MacOS" ]; then
        echo "Error: MacOS directory does not exist in $APP_TEMPLATE_PATH/Contents."
        return 1
    fi

    if [ ! -f "$APP_TEMPLATE_PATH/Contents/info.plist" ]; then
        echo "Error: info.plist file does not exist in $APP_TEMPLATE_PATH/Contents."
        return 1
    fi
    # Extract the binary name from the app template path
    local last_path
    last_path=$(basename "$APP_TEMPLATE_PATH")
    local app_name_without_extension=${last_path%.app}
    export BINARY_FILE_NAME=$app_name_without_extension

    # Validate the binary file is the same name as the app ( apple constraint )
    if [ ! -f "$APP_TEMPLATE_PATH/Contents/MacOS/$BINARY_FILE_NAME" ]; then
        echo "Error: $BINARY_FILE_NAME not found inside the MacOS folder."
        return 1
    fi

    return 0
}

validate_inputs(){
  # Validate input parameters
  if [ -z "$APPLE_CERT_DATA" ]; then
      echo "Error: Missing APPLE_CERT_DATA environment variable."
      exit 1
  fi
  if [ -z "$APPLE_CERT_PASSWORD" ]; then
      echo "Error: Missing APPLE_CERT_PASSWORD environment variable."
      exit 1
  fi
  if [ -z "$APPLE_TEAM_ID" ]; then
      echo "Error: Missing APPLE_TEAM_ID environment variable."
      exit 1
  fi
  # Validate app template structure
  if ! validate_app_template_structure; then
      echo "Error: The structure of APP_TEMPLATE_PATH is invalid. Please ensure it contains the following:"
      echo "-                YOUR_APP.app
                             ├── Contents
                                 ├── MacOS
                                 │   └── YOUR_APP (executable file)
                                 └── Info.plist"
      echo "- A valid .app structure is needed in order to sign & notarize the binary"
      exit 1
  fi
}

# Prepares the keychain and certificate for signing.
prepare_keychain_and_certificate() {
    local temp_dir
    temp_dir=$(mktemp -d)
    local keychain_name="macos-build.keychain"

    echo "$APPLE_CERT_DATA" | base64 --decode > "$temp_dir"/certs.p12

    security create-keychain -p "$APPLE_CERT_PASSWORD" $keychain_name
    security default-keychain -s $keychain_name
    security unlock-keychain -p "$APPLE_CERT_PASSWORD" $keychain_name
    security set-keychain-settings -t 3600 -u $keychain_name

    if ! security import "$temp_dir"/certs.p12 -k ~/Library/Keychains/$keychain_name -P "$APPLE_CERT_PASSWORD" -T /usr/bin/codesign; then
        echo "Error: Failed to import certificate into keychain."
        exit 1
    fi

    security find-identity -p codesigning -v
    security unlock-keychain -p "$APPLE_CERT_PASSWORD" $keychain_name
    security set-key-partition-list -S apple-tool:,apple:, -s -k "$APPLE_CERT_PASSWORD" -D "$APPLE_TEAM_ID" -t private $keychain_name
}

# Signs the binary file
sign_binary() {
    if ! codesign -s "$APPLE_TEAM_ID" --timestamp --deep --options runtime --force "$APP_TEMPLATE_PATH"/Contents/MacOS/"$BINARY_FILE_NAME"; then
        echo "Error: Failed to sign the binary."
        exit 1
    fi
    echo "Successfully signed the binary."
}

# Notarizes the app and staples the certificate.
notarize_app() {
    # Prepare temp dir to zip and unzip the app.
    # This is needed because notarization requires a zipped file.
    local temp_dir
    temp_dir=$(mktemp -d)
    local current_dir
    current_dir=$(pwd)

    cp -r "$APP_TEMPLATE_PATH" "$temp_dir"
    cd "$temp_dir" || exit

    local temp_zipped_name="$BINARY_FILE_NAME"-zipped
    if ! ditto -c -k --keepParent "$BINARY_FILE_NAME".app "./$temp_zipped_name"; then
        echo "Error: Failed to zip the app."
        exit 1
    fi
    # Send the zipped app for notarization
    if ! xcrun notarytool submit "$temp_zipped_name" --apple-id "$APPLE_ACCOUNT_ID" --team-id "$APPLE_TEAM_ID" --password "$APPLE_APP_SPECIFIC_PASSWORD" --force --wait; then
        echo "Error: Failed to notarize the app."
        exit 1
    fi
    echo "Notarization successful."

    # Unzip the app and staple the ticket
    unzip -o "$temp_zipped_name"
    if ! xcrun stapler staple "$BINARY_FILE_NAME".app; then
        echo "Error: Failed to staple the ticket to the app"
        exit 1
    fi
    echo "Stapling successful."

    # Copy the signed and notarized binary to the base directory
    # Clear the temp directory
    cp ./"$BINARY_FILE_NAME".app/Contents/MacOS/"$BINARY_FILE_NAME" "$current_dir"
    cd "$current_dir" || exit
    rm -rf "$temp_dir"
}

# Cleans up resources used during the process.
cleanup() {
    echo "Deleting keychain.."
    security delete-keychain "$keychain_name"
    echo "Delete Certificate..."
    rm -rf "$temp_dir"/certs.p12
}


# Main execution flow
validate_inputs
prepare_keychain_and_certificate
sign_binary
notarize_app
cleanup