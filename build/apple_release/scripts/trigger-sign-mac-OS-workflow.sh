#!/bin/bash


# This script triggers a GitHub Actions workflow to sign and notarize macOS binaries.

cliExecutableName=$1 # The name of the CLI executable to be processed

releaseVersion=$2 # The version of the release being processed

# Notice that the GITHUB_ACCESS_TOKEN is not defined in this script.
# It should be set as an environment variable before running the script.


# Trigger
curl -L \
   --retry 3 \
   -X POST \
   -H "Accept: application/vnd.github+json" \
   -H "Authorization: Bearer $GITHUB_ACCESS_TOKEN" \
   -H "X-GitHub-Api-Version: 2022-11-28" \
   https://api.github.com/repos/eyalDelarea/jfrog-cli/actions/workflows/prepareDarwinBinariesForRelease.yml/dispatches \
   -d "{\"ref\":\"sign_apple_binary\",\"inputs\":{\"releaseVersion\":\"$releaseVersion\",\"binaryFileName\":\"$cliExecutableName\"}}"