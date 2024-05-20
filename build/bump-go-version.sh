#!/bin/bash

# Check if a Go version is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <go-version>"
  exit 1
fi

GO_VERSION_FULL=$1
GO_VERSION_SHORT=${GO_VERSION_FULL%.*}

# Define the root directory of the project
ROOT_DIR=$(pwd)

validateVersionFormat() {
  if ! [[ $1 =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Invalid version format. Please provide a version in the format x.x.x"
    exit 1
  fi
}
# Function to update Go version in the GitHub Actions workflow files
updateWorkflows() {
  for file in "$ROOT_DIR"/.github/workflows/*.yml; do
    sed -i "" "s|go-version:.*|go-version: ${GO_VERSION_SHORT}\.x |" "$file"
  done
}

updateJenkinsFile() {
  # updates the jenkins tool specific go version
  sed -i "" "s|def goRoot = tool 'go-.*'|def goRoot = tool 'go-$GO_VERSION_FULL'|" "$ROOT_DIR"/Jenkinsfile
}

updateReleaseDockerImages() {
  # update slim image
  sed -i "" "s|/jfrog-docker/golang:.*-alpine.*|/jfrog-docker/golang:$GO_VERSION_SHORT-alpine as builder |" "$ROOT_DIR""/build/docker/slim/Dockerfile"
  # update full image
  sed -i "" "s|/jfrog-docker/golang:.*|/jfrog-docker/golang:$GO_VERSION_SHORT as builder |" "$ROOT_DIR""/build/docker/full/Dockerfile"
}

# updates readme go badge
updateReadme() {
  sed -i "" "s|https://tip.golang.org/doc/go.*|https://tip.golang.org/doc/go${GO_VERSION_SHORT})|" "$ROOT_DIR"/README.md
}

updateGoModVersion() {
  sed -i "" "s|^go 1.*|go ${GO_VERSION_SHORT}|" "$ROOT_DIR"/go.mod
  go mod tidy
}

gitCommit(){
  git add .
  git commit -m "Bump Go version to $GO_VERSION_FULL"
}

sanityCheck(){
  echo "Running fmt & vet ..."
  go fmt ./...
  go vet ./...
}

# Validate input
validateVersionFormat "$GO_VERSION_FULL"
# Update files
updateWorkflows
updateJenkinsFile
updateReleaseDockerImages
updateReadme
updateGoModVersion
sanityCheck
# Commit
#gitCommit


echo -e "
  Successfully Updated Go version to $GO_VERSION_FULL!
  NEXT_STEPS:
    1.Please make sure to install Go $GO_VERSION_FULL tool in Jenkins to allow CLI release
    2.Update the jfrog-ecosystem-integration-envs repository with the new Go version
    https://github.com/jfrog/jfrog-ecosystem-integration-env
    3.
    "

