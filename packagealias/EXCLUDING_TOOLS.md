# Excluding Tools from Ghost Frog Interception

Ghost Frog allows you to exclude specific tools from interception, so they run natively without being routed through JFrog CLI.

## Quick Start

### Exclude a Tool

```bash
# Exclude go from Ghost Frog interception
jf package-alias exclude go

# Now 'go' commands run natively
go build
go test
```

### Include a Tool Back

```bash
# Re-enable Ghost Frog interception for go
jf package-alias include go

# Now 'go' commands are intercepted again
go build  # → runs as: jf go build
```

## Use Cases

### 1. Tool Conflicts

Some tools might have conflicts when run through JFrog CLI:

```bash
# Exclude problematic tool
jf package-alias exclude docker

# Use native docker
docker build -t myapp .
```

### 2. Development vs Production

Exclude tools during local development, but keep them intercepted in CI/CD:

```bash
# Local development - use native tools
jf package-alias exclude mvn
jf package-alias exclude npm

# CI/CD - tools are intercepted (aliases are installed fresh)
```

### 3. Performance

For tools where JFrog CLI overhead isn't needed:

```bash
# Exclude fast tools that don't need build info
jf package-alias exclude gem
jf package-alias exclude bundle
```

## Commands

### `jf package-alias exclude <tool>`

Excludes a tool from Ghost Frog interception. The tool will run natively.

**Supported tools:**
- `mvn`, `gradle` (Java)
- `npm`, `yarn`, `pnpm` (Node.js)
- `go` (Go)
- `pip`, `pipenv`, `poetry` (Python)
- `dotnet`, `nuget` (.NET)
- `docker` (Containers)
- `gem`, `bundle` (Ruby)

**Example:**
```bash
$ jf package-alias exclude go
Tool 'go' is now configured to: run natively (excluded from interception)
Mode: pass
When you run 'go', it will execute the native tool directly without JFrog CLI interception.
```

### `jf package-alias include <tool>`

Re-enables Ghost Frog interception for a tool.

**Example:**
```bash
$ jf package-alias include go
Tool 'go' is now configured to: intercepted by JFrog CLI
Mode: jf
When you run 'go', it will be intercepted and run as 'jf go'.
```

## Checking Status

Use `jf package-alias status` to see which tools are excluded:

```bash
$ jf package-alias status
...
Tool Configuration:
  mvn        mode=jf    alias=✓ real=✓
  npm        mode=jf    alias=✓ real=✓
  go         mode=pass  alias=✓ real=✓  ← excluded
  pip        mode=jf    alias=✓ real=✓
```

**Mode meanings:**
- `mode=jf` - Intercepted by Ghost Frog (runs as `jf <tool>`)
- `mode=pass` - Excluded (runs natively)
- `mode=env` - Reserved for future use

## How It Works

When you exclude a tool:

1. **Configuration is saved** to `~/.jfrog/package-alias/config.yaml`
2. **Alias symlink remains** - the symlink still exists in `~/.jfrog/package-alias/bin/`
3. **Mode is set to `pass`** - When the alias is invoked, Ghost Frog detects the `pass` mode and runs the native tool directly

### Example Flow

```bash
# User runs: go build
# Shell resolves to: ~/.jfrog/package-alias/bin/go (alias)

# Ghost Frog detects:
# - Running as alias: yes
# - Tool mode: pass (excluded)
# - Action: Run native go directly

# Result: Native go build executes
```

## Configuration File

Tool exclusions are stored in `~/.jfrog/package-alias/config.yaml`:

```yaml
{
  "tool_modes": {
    "go": "pass",
    "docker": "pass"
  },
  "enabled": true
}
```

You can manually edit this file if needed, but using the CLI commands is recommended.

## Examples

### Exclude Multiple Tools

```bash
jf package-alias exclude go
jf package-alias exclude docker
jf package-alias exclude gem
```

### Exclude All Python Tools

```bash
jf package-alias exclude pip
jf package-alias exclude pipenv
jf package-alias exclude poetry
```

### Re-enable Everything

```bash
for tool in mvn gradle npm yarn pnpm go pip pipenv poetry dotnet nuget docker gem bundle; do
  jf package-alias include $tool
done
```

## Troubleshooting

### Tool Still Being Intercepted

1. **Check status:**
   ```bash
   jf package-alias status
   ```

2. **Verify exclusion:**
   ```bash
   jf package-alias exclude <tool>
   ```

3. **Check PATH order:**
   - Ensure alias directory is first in PATH
   - Run `which <tool>` to verify

### Tool Not Found After Exclusion

If excluding a tool causes "command not found":

1. **Verify real tool exists:**
   ```bash
   # Temporarily disable aliases
   jf package-alias disable
   
   # Check if tool exists
   which <tool>
   ```

2. **Re-enable aliases:**
   ```bash
   jf package-alias enable
   ```

## Best Practices

1. **Exclude sparingly** - Only exclude tools that have specific issues or requirements
2. **Document exclusions** - Note why tools are excluded in your team documentation
3. **Test in CI/CD** - Ensure excluded tools work correctly in your pipelines
4. **Use status command** - Regularly check `jf package-alias status` to see current configuration

## Related Commands

- `jf package-alias status` - View current tool configurations
- `jf package-alias enable` - Enable Ghost Frog globally
- `jf package-alias disable` - Disable Ghost Frog globally
- `jf package-alias install` - Install/update aliases

