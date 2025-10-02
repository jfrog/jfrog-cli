#!/bin/bash

# Script Purpose: Automate the process of signing and notarizing a macOS binary.

# Input:
# - APPLE_CERT_DATA: Base64 encoded data of the Apple Developer certificate.
# - APPLE_CERT_PASSWORD: Password for the Apple Developer certificate.
# - APPLE_TEAM_ID: Identifier for the Apple Developer Team.
# - APPLE_ACCOUNT_ID: Apple Developer Account ID.
# - APPLE_APP_SPECIFIC_PASSWORD: Password for app-specific services on the Apple Developer Account.
# - APP_TEMPLATE_PATH: Path to the .app bundle template.

# Output:
# A signed and notarized binary file in the current directory, ready for distribution.

validate_app_template_structure() {
    [ ! -d "$APP_TEMPLATE_PATH" ] && { echo "Error: $APP_TEMPLATE_PATH directory does not exist."; exit 1; }
    [ ! -d "$APP_TEMPLATE_PATH/Contents" ] && { echo "Error: Contents directory does not exist in $APP_TEMPLATE_PATH."; exit 1; }
    [ ! -d "$APP_TEMPLATE_PATH/Contents/MacOS" ] && { echo "Error: MacOS directory does not exist in $APP_TEMPLATE_PATH/Contents."; exit 1; }
    [ ! -f "$APP_TEMPLATE_PATH/Contents/Info.plist" ] && { echo "Error: Info.plist file does not exist in $APP_TEMPLATE_PATH/Contents."; exit 1; }

    local app_name_without_extension
    app_name_without_extension=$(basename "$APP_TEMPLATE_PATH" .app)
    export BINARY_FILE_NAME=$app_name_without_extension

    [ ! -f "$APP_TEMPLATE_PATH/Contents/MacOS/$BINARY_FILE_NAME" ] && { echo "Error: $BINARY_FILE_NAME executable not found inside the MacOS folder."; exit 1; }
}

validate_inputs() {
    [ -z "$APPLE_CERT_DATA" ] && { echo "Error: Missing APPLE_CERT_DATA environment variable."; exit 1; }
    [ -z "$APPLE_CERT_PASSWORD" ] && { echo "Error: Missing APPLE_CERT_PASSWORD environment variable."; exit 1; }
    [ -z "$APPLE_TEAM_ID" ] && { echo "Error: Missing APPLE_TEAM_ID environment variable."; exit 1; }

    validate_app_template_structure
}

prepare_keychain_and_certificate() {
    local temp_dir
    temp_dir=$(mktemp -d)
    local keychain_name="macos-build.keychain"

    echo "$APPLE_CERT_DATA" | base64 --decode > "$temp_dir/certs.p12"

    security create-keychain -p "$APPLE_CERT_PASSWORD" $keychain_name
    security default-keychain -s $keychain_name
    security unlock-keychain -p "$APPLE_CERT_PASSWORD" $keychain_name
    security set-keychain-settings -t 3600 -u $keychain_name

    security import "$temp_dir/certs.p12" -k ~/Library/Keychains/$keychain_name -P "$APPLE_CERT_PASSWORD" -T /usr/bin/codesign || { echo "Error: Failed to import certificate into keychain."; exit 1; }

    security set-key-partition-list -S apple-tool:,apple: -s -k "$APPLE_CERT_PASSWORD" -D "$APPLE_TEAM_ID" -t private $keychain_name
}

sign_binary() {
    codesign -s "$APPLE_TEAM_ID" --timestamp --deep --options runtime --force "$APP_TEMPLATE_PATH/Contents/MacOS/$BINARY_FILE_NAME" || { echo "Error: Failed to sign the binary."; exit 1; }
    echo "Successfully signed the binary."
}

notarize_app() {
    local temp_dir
    temp_dir=$(mktemp -d)
    local current_dir
    current_dir=$(pwd)

    cp -r "$APP_TEMPLATE_PATH" "$temp_dir"
    cd "$temp_dir" || exit

    local temp_zipped_name="${BINARY_FILE_NAME}-zipped.zip"
    ditto -c -k --keepParent "$BINARY_FILE_NAME.app" "./$temp_zipped_name" || { echo "Error: Failed to zip the app."; exit 1; }

    xcrun notarytool submit "$temp_zipped_name" --apple-id "$APPLE_ACCOUNT_ID" --team-id "$APPLE_TEAM_ID" --password "$APPLE_APP_SPECIFIC_PASSWORD" --wait || { echo "Error: Failed to notarize the app."; exit 1; }
    echo "Notarization successful."

    unzip -o "$temp_zipped_name"
    xcrun stapler staple "$BINARY_FILE_NAME.app" || { echo "Error: Failed to staple the ticket to the app."; exit 1; }
    echo "Stapling successful."

    cp "./$BINARY_FILE_NAME.app/Contents/MacOS/$BINARY_FILE_NAME" "$current_dir"
    cd "$current_dir" || exit
    rm -rf "$temp_dir"
}

# Cleans up resources used during the process.
cleanup() {
    echo "Deleting keychain..."
    security delete-keychain "macos-build.keychain"
    echo "Deleting temporary certificate files..."
    rm -rf "$temp_dir/certs.p12"
}

main() {
    validate_inputs
    prepare_keychain_and_certificate
    sign_binary
    notarize_app
    cleanup
}

main