#!/bin/bash

# Script Purpose: Download signed macOS binaries for a specific version and architecture.
# The name of the CLI executable to be processed - jfrog or jf
cliExecutableName=$1

# The version of the release being processed
releaseVersion=$2

# The architecture of the macOS binary to be downloaded - amd64 or arm64
goarch=$3

# GitHub Access Token for authentication
GITHUB_ACCESS_TOKEN=$3

# Function to retrieve the specific artifact URL with retries
get_specific_artifact_url_with_retries() {
    local max_retries=4
    # Cooldown in seconds between retries
    local cooldown=15
    local retry_count=0

    while [ $retry_count -lt $max_retries ]; do
        # Fetch the list of artifacts from GitHub
        response=$(curl -L --retry 3 \
            -H "Accept: application/vnd.github+json" \
            -H "Authorization: Bearer $GITHUB_ACCESS_TOKEN" \
            -H "X-GitHub-Api-Version: 2022-11-28" \
            -s https://api.github.com/repos/eyaldelarea/jfrog-cli/actions/artifacts)

        # Parse the response to find the URL of the desired artifact
        artifactUrl=$(echo "$response" | jq -r ".artifacts[] | select(.name | contains(\"$cliExecutableName-darwin-v$releaseVersion-$goarch\")) | .archive_download_url")

        # If a valid URL is found, return it
        if [[ "$artifactUrl" =~ ^https?://.+ ]]; then
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

    echo "Downloading signed executable from $artifactUrl"
    # Download the artifact
    curl -L \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer $GITHUB_ACCESS_TOKEN" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "$artifactUrl" -o artifact.zip

    echo "installing zip..."
    apt-get update
    apt-get install unzip

    # Extract the artifact and clean up
    unzip artifact.zip
    rm -rf artifact.zip

    # Check if the executable exists
    if [ ! -f "$cliExecutableName" ]; then
        echo "Error: Executable $cliExecutableName not found."
        exit 1
    fi
}

# Start the process
downloadSignedMacOSBinaries