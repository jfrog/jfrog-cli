#!/bin/bash

# This script is used to sign and notarize a binary for MacOS. It consumes the following environment variables:
#
# APPLE_CERT_DATA: The base64 encoded Apple certificate data.
# APPLE_CERT_PASSWORD: The password for the Apple certificate.
# APPLE_TEAM_ID: The Apple Team ID.
# APPLE_ACCOUNT_ID: The Apple Account ID.
# APPLE_APP_SPECIFIC_PASSWORD: The app-specific password for the Apple account.
#
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
# All of these environment variables should be currently in order for the process to work.

validate_app_template_structure() {
      if [ -z "$APP_TEMPLATE_PATH" ]; then
          echo "Error: APP_TEMPLATE_PATH is not set."
          return 1
      fi

      if [ ! -d "$APP_TEMPLATE_PATH" ]; then
          echo "Error: $APP_TEMPLATE_PATH is not a directory."
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

     # Extract the last path from the APP_TEMPLATE_PATH
      last_path=$(basename "$APP_TEMPLATE_PATH")
      # Remove the .app extension from the last path
      app_name_without_extension=${last_path%.app}
      # Export app_name_without_extension as an environment variable
      export BINARY_FILE_NAME=$app_name_without_extension

      # Check if the executable file exists in the MacOS folder
      if [ ! -f "$APP_TEMPLATE_PATH/Contents/MacOS/$EXECUTABLE_NAME" ]; then
              echo "Error: $EXECUTABLE_NAME not found inside the MacOS folder."
              return 1
      fi

      return 0
  }


validateInputs(){
  # Validate input parameters
  if [ -z "$APPLE_CERT_DATA" ] || [ -z "$APPLE_CERT_PASSWORD" ] || [ -z "$APPLE_TEAM_ID" ] || [ -z "$BINARY_FILE_NAME" ] ; then
      echo "Error: Missing environment variable."
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

# This function will prepare the keychain and certificate on the machine, needed for signing.
prepare_keychain_and_certificate() {
    # Assign environment variables to local variables
    # Set temp dir as runner temp dir
    TEMP_DIR=$RUNNER_TEMP
    KEYCHAIN_NAME="macos-build.keychain"
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
    if ! security import "$TEMP_DIR"/certs.p12 -k ~/Library/Keychains/$KEYCHAIN_NAME -P "$APPLE_CERT_PASSWORD" -T /usr/bin/codesign; then
        echo "Error: Failed to import certificate into keychain."
        exit 1
    fi

    # Verify the identity in the keychain
    echo "Verifying identity..."
    security find-identity -p codesigning -v

    # Unlock the keychain to allow signing in terminal without asking for password
    echo "Unlocking the keychain"
    security unlock-keychain -p "$APPLE_CERT_PASSWORD" $KEYCHAIN_NAME
    security set-key-partition-list -S apple-tool:,apple:, -s -k "$APPLE_CERT_PASSWORD" -D "$APPLE_TEAM_ID" -t private  $KEYCHAIN_NAME
}

# Signs the binary file copies into the apple bundle template
# The template is needed for notarizing the app
sign_binary(){
  # Sign the binary
  echo "Signing the binary..."
  if ! codesign -s  "$APPLE_TEAM_ID"  --timestamp --deep --options runtime --force "$BINARY_FILE_NAME"; then
      echo "Error: Failed to sign the binary."
      exit 1
  fi
}

# Sends the app for notarization and staples the certificate to the app.
# Binary files cannot be notarized as standalone files, they must be zipped and unzipped later on.
notarize_app(){
  # Move binary inside the app bundle template
  if ! mv "$BINARY_FILE_NAME" "$APP_TEMPLATE_PATH"/Contents/MacOS/"$BINARY_FILE_NAME" ; then
      echo "Error: Failed to move the binary to the app template. Please check files exists"
      exit 1
  fi
  # Zip it using ditto
  temp_zipped_name="$BINARY_FILE_NAME"-zipped

  if ! ditto -c -k --keepParent "$APP_TEMPLATE_PATH" ./"$temp_zipped_name"; then
      echo "Error: Failed to zip the app."
      exit 1
  fi

  # Notarize the zipped app
  if ! xcrun notarytool submit "$temp_zipped_name" --apple-id "$APPLE_ACCOUNT_ID" --team-id "$APPLE_TEAM_ID" --password "$APPLE_APP_SPECIFIC_PASSWORD"  --force --wait; then
      echo "Error: Failed to notarize the app."
      exit 1
  fi
  echo "Notarization successful."

  # Unzip the notarized app
  unzip -o "$temp_zipped_name"

  # Staple ticket to the app
  if ! xcrun stapler staple "$BINARY_FILE_NAME".app; then
      echo "Error: Failed to staple the ticket to the app"
      exit 1
  fi
  echo "Stapling successful."


}

cleanup(){
  echo "Deleting keychain.."
  security delete-keychain "$KEYCHAIN_NAME"
  echo "Delete Certificate..."
  rm -rf "$TEMP_DIR"/certs.p12
}


# Setup
validateInputs
prepare_keychain_and_certificate
# Sign & Notarize
sign_binary
notarize_app
# Cleanup
cleanup

