#!/bin/bash

cliExecutableName=$1
releaseVersion=$2
goarch=$3

# This script downloads signed macOS binaries for a specific version and architecture.

# Function to retrieve the specific artifact URL with retries
get_specific_artifact_url_with_retries() {
    local max_retries=5
    local cooldown=15 # Cooldown in seconds between retries
    local retry_count=0

    while [ $retry_count -lt $max_retries ]; do
        # Fetch the list of artifacts from GitHub
        response=$(curl -L \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer $GITHUB_ACCESS_TOKEN" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            -s https://api.github.com/repos/eyaldelarea/jfrog-cli/actions/artifacts)

        # Parse the response to find the URL of the desired artifact
        artifactUrl=$(echo "$response" | jq -r ".artifacts[] | select(.name | contains(\"$cliExecutableName-darwin-v$releaseVersion-$goarch\")) | .archive_download_url")

        # If a valid URL is found, return it
        if [[ -n "$artifactUrl" && "$artifactUrl" =~ ^https?://.+ ]]; then
            echo "$artifactUrl"
            return 0
        else
            # If not found, retry after a cooldown period
            retry_count=$((retry_count+1))
            sleep $cooldown
        fi
    done

    # If the maximum number of retries is exceeded, report failure
    echo "Curl request failed after $max_retries attempts."
    return 1
}

# Function to download and extract the signed macOS binaries
downloadSignedMacOSBinaries() {
    echo "Downloading Signed macOS Binaries for goarch: $goarch, release version: $releaseVersion"

    # Attempt to get the specific artifact URL
    artifactUrl=$(get_specific_artifact_url_with_retries)

    # Validate the URL
    if [[ -z "$artifactUrl" || ! "$artifactUrl" =~ ^https?://.+ ]]; then
        echo "Failed to find uploaded artifact for version: $releaseVersion and goarch: $goarch. Please validate the artifacts were successfully uploaded."
        exit 1
    fi

    # Download the artifact
    curl -L \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer $GITHUB_ACCESS_TOKEN" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "$artifactUrl" -o artifact.zip

    # Extract the artifact and clean up
    tar -xvf artifact.zip
    rm -rf artifact.zip

    # Make the binary executable
    chmod +x "$cliExecutableName"

    # Validate the binary by checking its version
    ./"$cliExecutableName" --version
}

# Start the process
downloadSignedMacOSBinaries