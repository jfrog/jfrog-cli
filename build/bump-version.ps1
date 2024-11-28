# Check if exactly one argument is provided
if ($args.Count -ne 1) {
    Write-Error "Error: Please provide exactly one argument - the version to bump."
    exit 1
}

# Assign the argument to the toVersion variable
$toVersion = $args[0]

# Debugging output for toVersion
Write-Host "Running bump-version.ps1 with toVersion: '$toVersion'"

if (-not $toVersion) {
    Write-Error "Error: toVersion is not set. Exiting."
    exit 1
}

# Function to get fromVersion from a file
function Populate-FromVersion {
    & build/build.bat
    $versionString = & ./jf -v
    Write-Host "Debug: versionString='$versionString'"

    # Extract the version from the output (e.g., 'jf version 2.71.5')
    if ($versionString -match "jf version ([0-9]+\.[0-9]+\.[0-9]+)") {
        $global:fromVersion = $matches[1]
    } else {
        Write-Error "Error: Failed to extract fromVersion. Unexpected format: '$versionString'."
        exit 1
    }

    Write-Host "Debug: fromVersion='$global:fromVersion'"
}

# Function to validate versions
function Validate-Versions {
    param (
        [string]$fromVersion,
        [string]$toVersion
    )

    Write-Host "Debug: Validating versions - fromVersion='$fromVersion', toVersion='$toVersion'"

    if (-not $fromVersion -or -not $toVersion) {
        Write-Error "Error: Both fromVersion and toVersion must have non-empty values."
        exit 1
    }

    if ($fromVersion -eq $toVersion) {
        Write-Error "Error: fromVersion and toVersion must have different values."
        exit 1
    }

    Write-Host "Bumping version from $fromVersion to $toVersion"
}

# Function to create a new Git branch
function Create-Branch {
    $branchName = "bump-ver-from-$fromVersion-to-$toVersion"

    Write-Host "Creating and switching to a new branch: $branchName"
    git remote rm upstream 2>$null
    git remote add upstream https://github.com/jfrog/jfrog-cli.git
    git checkout dev
    git fetch upstream dev
    git reset --hard upstream/dev
    git push origin dev
    git checkout -b $branchName
}

# Function to replace a version in a file
function Replace-Version {
    param (
        [string]$filePath,
        [string]$linePattern,
        [string]$fromVersion,
        [string]$toVersion
    )

    if (-not (Test-Path $filePath)) {
        Write-Error "Error: File '$filePath' not found."
        exit 1
    }

    Write-Host "Replacing version in file: $filePath"
    Write-Host "Looking for pattern: $linePattern"

    # Read the content, replace the line, and write it back
    $content = Get-Content -Path $filePath
    $found = $false
    $newContent = $content | ForEach-Object {
        if ($_ -match $linePattern) {
            $found = $true
            $_ -replace [regex]::Escape($fromVersion), $toVersion
        } else {
            $_
        }
    }

    if (-not $found) {
        Write-Error "Error: Pattern '$linePattern' not found in file '$filePath'."
        exit 1
    }

    Set-Content -Path $filePath -Value $newContent
    git add $filePath
}

# Main Script
Populate-FromVersion
Validate-Versions -fromVersion $fromVersion -toVersion $toVersion
Create-Branch

# Replace versions in specified files
Replace-Version -filePath "utils/cliutils/cli_consts.go" -linePattern 'CliVersion  = "' -fromVersion $fromVersion -toVersion $toVersion
Replace-Version -filePath "build/npm/v2/package-lock.json" -linePattern '"version": "' -fromVersion $fromVersion -toVersion $toVersion
Replace-Version -filePath "build/npm/v2/package.json" -linePattern '"version": "' -fromVersion $fromVersion -toVersion $toVersion
Replace-Version -filePath "build/npm/v2-jf/package-lock.json" -linePattern '"version": "' -fromVersion $fromVersion -toVersion $toVersion
Replace-Version -filePath "build/npm/v2-jf/package.json" -linePattern '"version": "' -fromVersion $fromVersion -toVersion $toVersion

# Commit and push changes
Write-Host "Committing and pushing changes"

# Commit the changes with a descriptive message
git commit -m "Bump version from $fromVersion to $toVersion"

# Push the branch to the remote repository
$branchName = "bump-ver-from-$fromVersion-to-$toVersion"
git push --set-upstream origin $branchName

# Success message
Write-Host ('Version bump successfully completed and pushed to branch: bump-ver-from-' + $fromVersion + '-to-' + $toVersion)
