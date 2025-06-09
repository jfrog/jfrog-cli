# JFrog CLI Generic Package Manager Framework

## Overview

The Generic Package Manager Framework is a new feature added to JFrog CLI that allows you to integrate any package manager with JFrog Artifactory using a simple, configuration-driven approach. Instead of requiring complex native integrations for each package manager, this framework provides a unified command interface that works with any package manager through YAML configuration.

## Problem Solved

**Before:** Adding support for new package managers to JFrog CLI required:
- Extensive code changes
- Deep integration with package manager APIs  
- Complex testing and validation
- Long development cycles

**After:** Adding a new package manager now requires:
- Adding a few lines to a YAML configuration file
- Immediate availability of both native and JFrog-integrated commands
- Automatic configuration management
- Consistent user experience across all package managers

## How It Works

The framework operates through a simple command structure:

```bash
jf pkg <package-manager> <command> [args...]
```

### Supported Package Managers
- **Ruby** (`gem`) - Full JFrog integration
- **PHP** (`composer`) - Basic integration
- **Swift** (`swift`) - Basic integration  
- **Rust** (`cargo`) - Basic integration

### Command Types

1. **Native Commands** - Execute package manager's native commands
2. **JFrog Commands** - Integrated Artifactory operations:
   - `config` - Configure JFrog integration
   - `publish` - Upload packages to Artifactory with build info

## Usage Examples

### Ruby Gems
```bash
# Configure Ruby integration
jf pkg ruby config --server-id=my-server --RUBY_VIRTUAL_REPO=gems-virtual --RUBY_REPO=gems-local

# Install gems (native command)
jf pkg ruby install colorize

# Publish gem to Artifactory (JFrog command)
jf pkg ruby publish my-gem-1.0.0.gem --build-name=ruby-build --build-number=1
```

### PHP Composer 
```bash
# Configure PHP integration
jf pkg php config --server-id=my-server --PHP_VIRTUAL_REPO=composer-virtual --PHP_REPO=composer-local

# Install dependencies (native command)
jf pkg php install

# Publish package (JFrog command)
jf pkg php publish --build-name=php-build --build-number=1
```

## Architecture

### Configuration File (`buildtools/packages.yaml`)
```yaml
package_managers:
  ruby:
    executable: "gem"
    commands:                    # Native commands
      install: "gem install"
      list: "gem list"
    jfrog_commands:              # JFrog-integrated commands
      publish: "jfrog_ruby_publish"
      config: "jfrog_ruby_config"
  
  php:
    executable: "composer"
    commands:
      install: "composer install"
      update: "composer update"
    jfrog_commands:
      publish: "jfrog_php_publish"
      config: "jfrog_php_config"
```

### JFrog Configuration Files
Located in `.jfrog/projects/<package-manager>.yaml`:

```yaml
version: 1
type: ruby
resolver:
    serverID: my-server
    repo: gems-virtual
deployer:
    serverID: my-server  
    repo: gems-local
gemBuildConfig:
    gemPushToRepo: true
```

## Adding New Package Managers

Adding support for a new package manager requires only **3 simple steps**:

### Step 1: Update Configuration
Add your package manager to `buildtools/packages.yaml`:

```yaml
package_managers:
  # ... existing package managers ...
  
  python:  # Your new package manager
    executable: "pip"
    commands:
      install: "pip install"
      list: "pip list"  
      freeze: "pip freeze"
    jfrog_commands:
      publish: "jfrog_python_publish"
      config: "jfrog_python_config"
```

### Step 2: Add JFrog Command Handlers
Add cases to the `executeJFrogCommand` function in `buildtools/cli.go`:

```go
case "jfrog_python_publish":
    return executePythonPublish(args, c)
case "jfrog_python_config":
    return executeGenericConfig("Python/Pip", packageManager, args, c)
```

### Step 3: Implement Command Functions
Add the publish and config functions:

```go
func executePythonPublish(args []string, c *cli.Context) error {
    // Handle help flag
    if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
        fmt.Printf(`Usage: jf pkg python publish <package> [options]
        
Publish a Python package to JFrog Artifactory
// ... help content ...
`)
        return nil
    }
    
    // Implementation for Python package upload to Artifactory
    fmt.Printf("Publishing Python package to Artifactory...\n")
    // Add your Python-specific upload logic here
    return nil
}
```

**That's it!** Your new package manager is now available with:
- Native command execution
- JFrog configuration management  
- Help system integration
- Error handling

## Technical Details

### Command Resolution Flow
1. Parse `jf pkg <package-manager> <command>`
2. Load `packages.yaml` configuration
3. Check if command exists in `jfrog_commands` (JFrog integration)
4. If not found, check `commands` (native execution)
5. Execute appropriate handler

### Configuration Management
- Configurations stored in `.jfrog/projects/<package-manager>.yaml`
- Follows JFrog CLI configuration patterns
- Supports environment variable substitution
- Server authentication handled automatically

### Upload Implementation
- Uses JFrog Client SDK for Artifactory uploads
- Supports build info collection
- Automatic property setting for build correlation
- Standard Artifactory repository patterns

## Benefits

### For Users
- **Consistent Interface** - Same command structure across all package managers
- **Native + JFrog** - Access both native package manager features and JFrog integration
- **Easy Configuration** - Simple YAML-based setup
- **Build Integration** - Automatic build info collection

### For Developers  
- **Rapid Development** - Add new package managers in minutes, not months
- **Minimal Code** - Most functionality handled by framework
- **Standardized** - Consistent patterns across all package managers
- **Maintainable** - Configuration-driven instead of code-driven

## Future Enhancements

The framework is designed for easy extension:

- **More Package Managers** - Python, Node.js, .NET, etc.
- **Advanced Features** - Dependency scanning, vulnerability checks
- **Template System** - Pre-built configurations for common setups
- **Plugin Architecture** - Custom command extensions

## Getting Started

1. **Configure your package manager:**
   ```bash
   jf pkg <package-manager> config --server-id=<your-server>
   ```

2. **Use native commands:**
   ```bash
   jf pkg <package-manager> <native-command> [args...]
   ```

3. **Publish to Artifactory:**
   ```bash
   jf pkg <package-manager> publish <package> --build-name=<build> --build-number=<number>
   ```

---

**The Generic Package Manager Framework makes JFrog CLI extensible and future-proof, enabling rapid integration of any package manager with minimal development effort.** 