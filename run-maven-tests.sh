#!/bin/bash

# Maven Tests Local Runner Script
# This script helps you run Maven integration tests locally by connecting to an existing Artifactory

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Maven Tests Local Runner${NC}"
echo -e "${GREEN}========================================${NC}"

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"

# Check Maven
if ! command -v mvn &> /dev/null; then
    echo -e "${RED}❌ Maven is not installed. Please install Maven first.${NC}"
    exit 1
fi
MAVEN_VERSION=$(mvn -version | head -n 1)
echo -e "${GREEN}✓ Maven found: $MAVEN_VERSION${NC}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed. Please install Go first.${NC}"
    exit 1
fi
GO_VERSION=$(go version)
echo -e "${GREEN}✓ Go found: $GO_VERSION${NC}"

# Configuration - Update these with your Artifactory details
echo -e "\n${YELLOW}========================================${NC}"
echo -e "${YELLOW}Artifactory Configuration${NC}"
echo -e "${YELLOW}========================================${NC}"

# Default values - you can override these by setting environment variables
JFROG_URL="${JFROG_URL:-http://localhost:8081/}"
JFROG_USER="${JFROG_USER:-admin}"
JFROG_PASSWORD="${JFROG_PASSWORD:-password}"
JFROG_ACCESS_TOKEN="${JFROG_ACCESS_TOKEN:-}"

# Prompt for configuration if not set
echo -e "\nCurrent configuration:"
echo -e "  URL: ${GREEN}${JFROG_URL}${NC}"
echo -e "  User: ${GREEN}${JFROG_USER}${NC}"

if [ -n "$JFROG_ACCESS_TOKEN" ]; then
    echo -e "  Auth: ${GREEN}Access Token${NC}"
else
    echo -e "  Auth: ${GREEN}Username/Password${NC}"
fi

echo -e "\n${YELLOW}To change configuration, set these environment variables:${NC}"
echo -e "  export JFROG_URL='https://your-artifactory.jfrog.io/'"
echo -e "  export JFROG_USER='your-username'"
echo -e "  export JFROG_PASSWORD='your-password'"
echo -e "  # OR use access token:"
echo -e "  export JFROG_ACCESS_TOKEN='your-access-token'"

echo -e "\n${YELLOW}Press Enter to continue with current configuration, or Ctrl+C to exit...${NC}"
read -r

# Build test command
echo -e "\n${YELLOW}========================================${NC}"
echo -e "${YELLOW}Running Maven Tests${NC}"
echo -e "${YELLOW}========================================${NC}"

TEST_CMD="go test -v github.com/jfrog/jfrog-cli --timeout 0 --test.maven"
TEST_CMD="$TEST_CMD -jfrog.url='${JFROG_URL}'"
TEST_CMD="$TEST_CMD -jfrog.user='${JFROG_USER}'"

if [ -n "$JFROG_ACCESS_TOKEN" ]; then
    TEST_CMD="$TEST_CMD -jfrog.adminToken='${JFROG_ACCESS_TOKEN}'"
else
    TEST_CMD="$TEST_CMD -jfrog.password='${JFROG_PASSWORD}'"
fi

echo -e "\n${GREEN}Executing test command...${NC}"
echo -e "${YELLOW}Note: Tests will create repositories (cli-mvn1, cli-mvn2, cli-mvn-remote) in your Artifactory${NC}\n"

# Run the tests
eval $TEST_CMD

TEST_RESULT=$?

echo -e "\n${YELLOW}========================================${NC}"
if [ $TEST_RESULT -eq 0 ]; then
    echo -e "${GREEN}✓ Tests completed successfully!${NC}"
else
    echo -e "${RED}✗ Tests failed with exit code: $TEST_RESULT${NC}"
fi
echo -e "${YELLOW}========================================${NC}"

exit $TEST_RESULT









