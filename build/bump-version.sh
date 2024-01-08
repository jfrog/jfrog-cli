#!/bin/bash

# Function to get fromVersion from a file
populateFromVersion() {
    build/build.sh
    fromVersion=$(./jf -v | tr -d 's/jf version//g' | tr -d '\n')
}

# Function to validate arguments
validateArg() {
    # Check if both arguments are provided
    if [ $# -ne 1 ]; then
        echo "Error: Please provide exactly one argument - the version to bump."
        exit 1
    fi
}

validateVersions() {
    # Extract arguments
    fromVersion=$1
    toVersion=$2

    # Check if arguments are non-empty
    if [ -z "$fromVersion" ] || [ -z "$toVersion" ]; then
        echo "Error: Both fromVersion and toVersion must have non-empty values."
        exit 1
    fi

    # Check if arguments are not identical
    if [ "$fromVersion" = "$toVersion" ]; then
        echo "Error: fromVersion and toVersion must have different values."
        exit 1
    fi

    echo "Bumping version from $fromVersion to $toVersion"
}

createBranch() {
  branchName="bump-ver-from-$fromVersion-to-$toVersion"
  git remote rm upstream
  git remote add upstream https://github.com/jfrog/jfrog-cli.git
  git checkout dev
  git fetch upstream dev
  git pull upstream dev
  git push
  git checkout -b "$branchName"
}

# Function to replace version in file
replaceVersion() {
    local filePath=$1
    local line=$2
    local fromVersion=$3
    local toVersion=$4

    # Check if the file exists
    if [ ! -e "$filePath" ]; then
        echo "Error: File '$filePath' not found."
        exit 1
    fi

    # Use awk to replace the value if the line is found
    awk -v line="$line" -v from="$fromVersion" -v to="$toVersion" '
        index($0, line) {
            gsub(from, to);
            found=1;
        }
        { print }
        END {
            if (found != 1) {
                print "Error: The specified line ('" line "') does not exist in the file ('" filePath "').";
                exit 1;
            }
        }
    ' "$filePath" > "$filePath.tmp" && mv "$filePath.tmp" "$filePath"

    # Validate if the file was modified using git
    if git diff --exit-code "$filePath" > /dev/null; then
        echo "Error: File '$filePath' was not modified."
        exit 1
    fi

    git add "$filePath"
}

## Validate the argument was received.
validateArg "$@"

## Read the script argument into the toVersion variable
toVersion=$1

## Call the function to populate the fromVersion argument from the current version of the local JFrog CLI binary
populateFromVersion

## Call the function to validate arguments
validateVersions "$fromVersion" "$toVersion"

## Create a new branch
createBranch

## Add calls to the function to replace version in file with specified filePath values
replaceVersion "utils/cliutils/cli_consts.go" "CliVersion  = \"$fromVersion\"" "$fromVersion" "$toVersion"
replaceVersion "build/npm/v2/package-lock.json" "\"version\": \"$fromVersion\"," "$fromVersion" "$toVersion"
replaceVersion "build/npm/v2/package.json"  "\"version\": \"$fromVersion\"," "$fromVersion" "$toVersion"
replaceVersion "build/npm/v2-jf/package-lock.json" "\"version\": \"$fromVersion\"," "$fromVersion" "$toVersion"
replaceVersion "build/npm/v2-jf/package.json" "\"version\": \"$fromVersion\"," "$fromVersion" "$toVersion"

## Print success message if validation and replacement pass
echo "Version bumped successfully."

## Push the new branch, with the version bump
git commit -m "Bump version from $fromVersion to $toVersion"
git push --set-upstream origin "$branchName"
