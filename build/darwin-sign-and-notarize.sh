#!/bin/bash

# This script is used to sign and notarize a binary for MacOS. It consumes the following environment variables:
#
# APPLE_CERT_DATA: The base64 encoded Apple certificate data.
# APPLE_CERT_PASSWORD: The password for the Apple certificate.
# APPLE_TEAM_ID: The Apple Team ID.
# APPLE_ACCOUNT_ID: The Apple Account ID.
# APPLE_APP_SPECIFIC_PASSWORD: The app-specific password for the Apple account.
# APP_TEMPLATE_PATH: The path to the .app template folder used for notarization. It should have a specific structure:
# Create a folder containing the following structure:
#               YOUR_APP.app
#               ├── Contents
#                   ├── MacOS
#                   │   └── YOUR_APP (executable file)
#                   └── Info.plist
# Info.plist file contains apple specific app information which should be filled by the user.
# The name of the executable file should match the name of the YOUR_APP.app folder, i.e YOUR_APP.
#
# The output of the script is the signed and notarized binary file into the current directory.


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

    local last_path
    last_path=$(basename "$APP_TEMPLATE_PATH")
    local app_name_without_extension=${last_path%.app}
    export BINARY_FILE_NAME=$app_name_without_extension

    if [ ! -f "$APP_TEMPLATE_PATH/Contents/MacOS/$BINARY_FILE_NAME" ]; then
        echo "Error: $BINARY_FILE_NAME not found inside the MacOS folder."
        return 1
    fi

    return 0
}


validateInputs(){
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
    local temp_dir=$RUNNER_TEMP
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

    if ! xcrun notarytool submit "$temp_zipped_name" --apple-id "$APPLE_ACCOUNT_ID" --team-id "$APPLE_TEAM_ID" --password "$APPLE_APP_SPECIFIC_PASSWORD" --force --wait; then
        echo "Error: Failed to notarize the app."
        exit 1
    fi
    echo "Notarization successful."

    unzip -o "$temp_zipped_name"
    if ! xcrun stapler staple "$BINARY_FILE_NAME".app; then
        echo "Error: Failed to staple the ticket to the app"
        exit 1
    fi
    echo "Stapling successful."

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