#!/bin/bash


# This script triggers a GitHub Actions workflow to sign and notarize macOS binaries.

cliExecutableName=$1 # The name of the CLI executable to be processed

releaseVersion=$2 # The version of the release being processed

GITHUB_ACCESS_TOKEN=$3 # GitHub Access Token for authentication

# Trigger
curl -L \
   --retry 3 \
   -X POST \
   -H "Accept: application/vnd.github+json" \
   -H "Authorization: Bearer $GITHUB_ACCESS_TOKEN" \
   -H "X-GitHub-Api-Version: 2022-11-28" \
   https://api.github.com/repos/jfrog/jfrog-cli/actions/workflows/prepareDarwinBinariesForRelease.yml/dispatches \
   -d "{\"ref\":\"v2\",\"inputs\":{\"releaseVersion\":\"$releaseVersion\",\"binaryFileName\":\"$cliExecutableName\"}}"