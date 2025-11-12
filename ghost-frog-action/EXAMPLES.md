# Ghost Frog GitHub Action Examples

## 🎯 Available Demo Workflows

### 1. [Basic NPM Demo](../.github/workflows/ghost-frog-demo.yml)
Shows how npm commands are transparently intercepted by Ghost Frog.

```yaml
# Highlights:
- npm install becomes → jf npm install
- Shows debug output to see interception in action
- Demonstrates enable/disable functionality
```

### 2. [Multi-Tool Demo](../.github/workflows/ghost-frog-multi-tool.yml)
Comprehensive demo showing NPM, Maven, and Python in a single workflow.

```yaml
# Highlights:
- Multiple package managers in one workflow
- Real-world project structures
- Shows universal Ghost Frog configuration
```

### 3. [Simple Usage Example](../.github/workflows/example-ghost-frog-usage.yml)
The simplest possible integration - just add the action and go!

```yaml
# Highlights:
- Minimal setup required
- Before/after comparison
- Focus on ease of adoption
```

### 4. [Matrix Build Demo](../.github/workflows/ghost-frog-matrix-demo.yml)
Advanced example with multiple language versions in parallel.

```yaml
# Highlights:
- Node.js 16 & 18
- Python 3.9 & 3.11  
- Java 11 & 17
- All using the same Ghost Frog setup!
```

## 🚀 Quick Start Template

Copy this into your `.github/workflows/build.yml`:

```yaml
name: Build with Ghost Frog
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      # Add Ghost Frog - that's it!
      - uses: jfrog/jfrog-cli/ghost-frog-action@main
        with:
          jfrog-url: ${{ secrets.JFROG_URL }}
          jfrog-access-token: ${{ secrets.JFROG_ACCESS_TOKEN }}
      
      # Your existing build steps work unchanged
      - run: npm install
      - run: npm test
      - run: npm run build
```

## 💡 Integration Patterns

### Pattern 1: Add to Existing Workflow
```yaml
# Just add this step before your build commands
- uses: jfrog/jfrog-cli/ghost-frog-action@main
  with:
    jfrog-url: ${{ secrets.JFROG_URL }}
    jfrog-access-token: ${{ secrets.JFROG_ACCESS_TOKEN }}
```

### Pattern 2: Conditional Integration
```yaml
- uses: jfrog/jfrog-cli/ghost-frog-action@main
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
  with:
    jfrog-url: ${{ secrets.JFROG_URL }}
    jfrog-access-token: ${{ secrets.JFROG_ACCESS_TOKEN }}
```

### Pattern 3: Development vs Production
```yaml
- uses: jfrog/jfrog-cli/ghost-frog-action@main
  with:
    jfrog-url: ${{ github.ref == 'refs/heads/main' && secrets.PROD_JFROG_URL || secrets.DEV_JFROG_URL }}
    jfrog-access-token: ${{ github.ref == 'refs/heads/main' && secrets.PROD_TOKEN || secrets.DEV_TOKEN }}
```

## 🔒 Security Best Practices

1. **Always use GitHub Secrets** for credentials:
   ```yaml
   jfrog-access-token: ${{ secrets.JFROG_ACCESS_TOKEN }}  # ✅ Good
   jfrog-access-token: "my-token-123"                     # ❌ Never do this!
   ```

2. **Use minimal permissions** for access tokens

3. **Rotate tokens regularly** using GitHub's secret scanning

## 📚 More Resources

- [Ghost Frog Action README](README.md)
- [JFrog CLI Documentation](https://www.jfrog.com/confluence/display/CLI/JFrog+CLI)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
