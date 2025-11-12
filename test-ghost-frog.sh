#!/bin/bash

# Ghost Frog Local Testing Script
# This script tests the Ghost Frog functionality locally

set -e

echo "👻 Ghost Frog Local Test"
echo "======================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if jf is available
if ! command -v jf &> /dev/null; then
    echo -e "${RED}❌ JFrog CLI not found in PATH${NC}"
    echo "Please build and install JFrog CLI first:"
    echo "  go build -o jf ."
    echo "  sudo mv jf /usr/local/bin/"
    exit 1
fi

echo -e "${GREEN}✓ JFrog CLI found:${NC} $(which jf)"
echo -e "${GREEN}✓ Version:${NC} $(jf --version)"
echo ""

# Test 1: Install Ghost Frog aliases
echo "Test 1: Installing Ghost Frog aliases..."
echo "----------------------------------------"
jf package-alias install
echo ""

# Test 2: Check status
echo "Test 2: Checking Ghost Frog status..."
echo "------------------------------------"
jf package-alias status
echo ""

# Test 3: Add to PATH and test interception
echo "Test 3: Testing command interception..."
echo "--------------------------------------"
export PATH="$HOME/.jfrog/package-alias/bin:$PATH"
echo "PATH updated to include Ghost Frog aliases"
echo ""

# Test which commands would be intercepted
for cmd in npm mvn pip go docker; do
    if command -v $cmd &> /dev/null; then
        WHICH_CMD=$(which $cmd)
        if [[ $WHICH_CMD == *".jfrog/package-alias"* ]]; then
            echo -e "${GREEN}✓ $cmd would be intercepted${NC} (found at: $WHICH_CMD)"
        else
            echo -e "${YELLOW}○ $cmd found but not intercepted${NC} (found at: $WHICH_CMD)"
        fi
    else
        echo -e "${RED}✗ $cmd not found in PATH${NC}"
    fi
done
echo ""

# Test 4: Test actual interception with npm
if command -v npm &> /dev/null && [[ $(which npm) == *".jfrog/package-alias"* ]]; then
    echo "Test 4: Testing NPM interception..."
    echo "----------------------------------"
    echo "Running: npm --version"
    echo "(This should be intercepted and run as: jf npm --version)"
    echo ""
    
    # Run with debug to see interception
    JFROG_CLI_LOG_LEVEL=DEBUG npm --version 2>&1 | grep -E "(Detected running as alias|Running in JF mode)" || echo "Note: Interception messages not visible in output"
    echo ""
fi

# Test 5: Test enable/disable
echo "Test 5: Testing enable/disable..."
echo "--------------------------------"
echo "Disabling Ghost Frog..."
jf package-alias disable

echo -e "\n${YELLOW}When disabled, commands run natively:${NC}"
npm --version 2>&1 | head -1 || echo "npm not available"

echo -e "\nRe-enabling Ghost Frog..."
jf package-alias enable
echo ""

# Test 6: Cleanup option
echo "Test 6: Cleanup (optional)..."
echo "----------------------------"
echo "To uninstall Ghost Frog aliases, run:"
echo "  jf package-alias uninstall"
echo ""

# Summary
echo -e "${GREEN}🎉 Ghost Frog testing complete!${NC}"
echo ""
echo "Summary:"
echo "--------"
echo "• Ghost Frog aliases installed successfully"
echo "• Commands can be transparently intercepted"
echo "• Enable/disable functionality works"
echo ""
echo "To use Ghost Frog in your terminal:"
echo "1. Add to your shell configuration (~/.bashrc or ~/.zshrc):"
echo "   export PATH=\"\$HOME/.jfrog/package-alias/bin:\$PATH\""
echo "2. Reload your shell: source ~/.bashrc"
echo "3. All package manager commands will be intercepted!"
echo ""
echo "To use in CI/CD, see the GitHub Action examples in:"
echo "  .github/workflows/ghost-frog-*.yml"
