#!/bin/bash

# Function to validate arguments
validateArgs() {
    # Check if both arguments are provided
    if [ $# -ne 2 ]; then
        echo "Error: Please provide exactly two arguments - fromVersion and toVersion."
        exit 1
    fi

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

# Call the function     to validate arguments
validateArgs "$@"

# Add calls to the function to replace version in file with specified filePath values
replaceVersion "utils/cliutils/cli_consts.go" "CliVersion  = \"$fromVersion\"" "$fromVersion" "$toVersion"
replaceVersion "build/npm/v2/package-lock.json" "\"version\": \"$fromVersion\"," "$fromVersion" "$toVersion"
replaceVersion "build/npm/v2/package.json"  "\"version\": \"$fromVersion\"," "$fromVersion" "$toVersion"
replaceVersion "build/npm/v2-jf/package-lock.json" "\"version\": \"$fromVersion\"," "$fromVersion" "$toVersion"
replaceVersion "build/npm/v2-jf/package.json" "\"version\": \"$fromVersion\"," "$fromVersion" "$toVersion"

# Print success message if validation and replacement pass
echo "Version bumped successfully."

git commit -m "Bump version from $fromVersion to $toVersion"
git push
echo "Version bump pushed to git."
