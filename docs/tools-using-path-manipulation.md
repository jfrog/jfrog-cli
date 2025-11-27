# Tools Using PATH Manipulation

This document lists tools that manipulate the `PATH` environment variable to achieve their functionality, similar to how Ghost Frog works.

## Version Managers

### 1. **nvm (Node Version Manager)**
- **Purpose:** Manage multiple Node.js versions
- **PATH Manipulation:** Adds `~/.nvm/versions/node/vX.X.X/bin` to PATH
- **How it works:**
  ```bash
  nvm use 18.0.0
  # Adds: ~/.nvm/versions/node/v18.0.0/bin to PATH
  # Now 'node' resolves to the selected version
  ```
- **Similarity to Ghost Frog:** Creates symlinks/wrappers in a directory added to PATH

### 2. **pyenv (Python Version Manager)**
- **Purpose:** Manage multiple Python versions
- **PATH Manipulation:** Adds `~/.pyenv/shims` to PATH
- **How it works:**
  ```bash
  pyenv install 3.11.0
  pyenv global 3.11.0
  # Adds: ~/.pyenv/shims to PATH
  # 'python' and 'pip' resolve to shims that route to selected version
  ```
- **Similarity to Ghost Frog:** Uses shim scripts in PATH to intercept commands

### 3. **rbenv (Ruby Version Manager)**
- **Purpose:** Manage multiple Ruby versions
- **PATH Manipulation:** Adds `~/.rbenv/shims` to PATH
- **How it works:** Similar to pyenv, uses shim scripts
- **Similarity to Ghost Frog:** Intercepts Ruby commands via PATH manipulation

### 4. **jenv (Java Version Manager)**
- **Purpose:** Manage multiple Java versions
- **PATH Manipulation:** Adds `~/.jenv/shims` to PATH
- **How it works:** Wraps Java executables with version selection logic

### 5. **gvm (Go Version Manager)**
- **Purpose:** Manage multiple Go versions
- **PATH Manipulation:** Adds `~/.gvm/gos/goX.X.X/bin` to PATH
- **How it works:** Switches PATH to point to different Go installations

## Package Managers & Wrappers

### 6. **npm (Node Package Manager)**
- **Purpose:** Install Node.js packages
- **PATH Manipulation:** Adds `node_modules/.bin` to PATH when running scripts
- **How it works:**
  ```bash
  npm install
  # Adds: ./node_modules/.bin to PATH
  # Local binaries become available
  ```
- **Similarity to Ghost Frog:** Adds local bin directory to PATH

### 7. **pipx**
- **Purpose:** Install Python applications in isolated environments
- **PATH Manipulation:** Adds `~/.local/bin` to PATH
- **How it works:** Installs packages in isolated venvs, creates symlinks in `~/.local/bin`

### 8. **cargo (Rust)**
- **Purpose:** Rust package manager
- **PATH Manipulation:** Adds `~/.cargo/bin` to PATH
- **How it works:** Installs Rust binaries to `~/.cargo/bin`

### 9. **Homebrew (macOS)**
- **Purpose:** Package manager for macOS
- **PATH Manipulation:** Adds `/opt/homebrew/bin` or `/usr/local/bin` to PATH
- **How it works:** Installs packages to a directory that's added to PATH

## Development Environment Tools

### 10. **direnv**
- **Purpose:** Load/unload environment variables per directory
- **PATH Manipulation:** Dynamically modifies PATH based on `.envrc` files
- **How it works:**
  ```bash
  # In project directory with .envrc:
  export PATH="$PWD/bin:$PATH"
  # direnv automatically loads this when you cd into directory
  ```
- **Similarity to Ghost Frog:** Modifies PATH per-process

### 11. **asdf (Version Manager)**
- **Purpose:** Manage multiple runtime versions (universal version manager)
- **PATH Manipulation:** Adds `~/.asdf/shims` to PATH
- **How it works:** Creates shims for all managed tools (node, python, ruby, etc.)
- **Similarity to Ghost Frog:** Uses shim directory in PATH to intercept commands

### 12. **nix-shell**
- **Purpose:** Reproducible development environments
- **PATH Manipulation:** Completely replaces PATH with Nix-managed paths
- **How it works:** Creates isolated environments with specific tool versions

### 13. **conda/mamba**
- **Purpose:** Package and environment management
- **PATH Manipulation:** Adds conda environment's `bin` directory to PATH
- **How it works:**
  ```bash
  conda activate myenv
  # Adds: ~/anaconda3/envs/myenv/bin to PATH
  ```

## CI/CD & Build Tools

### 14. **GitHub Actions**
- **Purpose:** CI/CD automation
- **PATH Manipulation:** Uses `$GITHUB_PATH` to add directories
- **How it works:**
  ```yaml
  - run: echo "$HOME/.local/bin" >> $GITHUB_PATH
  # Adds directory to PATH for subsequent steps
  ```
- **Similarity to Ghost Frog:** Used in CI/CD environments

### 15. **Jenkins**
- **Purpose:** CI/CD automation
- **PATH Manipulation:** Modifies PATH per build
- **How it works:** Sets PATH in build environment

### 16. **Docker**
- **Purpose:** Containerization
- **PATH Manipulation:** Sets PATH in container images
- **How it works:**
  ```dockerfile
  ENV PATH="/app/bin:${PATH}"
  ```

## Security & Audit Tools

### 17. **asdf-vm with security plugins**
- **Purpose:** Version management with security scanning
- **PATH Manipulation:** Similar to asdf, but adds security wrappers

### 18. **Snyk CLI**
- **Purpose:** Security vulnerability scanning
- **PATH Manipulation:** Can wrap package managers to inject scanning
- **Similarity to Ghost Frog:** Intercepts package manager commands

### 19. **WhiteSource/Mend**
- **Purpose:** Open source security management
- **PATH Manipulation:** Some versions use PATH manipulation to intercept builds

## Proxy & Interception Tools

### 20. **proxychains**
- **Purpose:** Route network traffic through proxies
- **PATH Manipulation:** Uses LD_PRELOAD, but can manipulate PATH for tool discovery

### 21. **fakeroot**
- **Purpose:** Fake root privileges for builds
- **PATH Manipulation:** Adds fake root tools to PATH

### 22. **ccache**
- **Purpose:** Compiler cache
- **PATH Manipulation:** Adds wrapper scripts to PATH that intercept compiler calls

## Language-Specific Tools

### 23. **virtualenv/venv (Python)**
- **Purpose:** Python virtual environments
- **PATH Manipulation:** Adds venv's `bin` directory to PATH
- **How it works:**
  ```bash
  source venv/bin/activate
  # Adds: ./venv/bin to PATH
  ```

### 24. **rvm (Ruby Version Manager)**
- **Purpose:** Manage Ruby versions
- **PATH Manipulation:** Adds `~/.rvm/bin` and version-specific paths
- **Note:** Less popular now, rbenv is preferred

### 25. **sdkman**
- **Purpose:** SDK manager for JVM-based tools
- **PATH Manipulation:** Adds `~/.sdkman/candidates/*/current/bin` to PATH
- **How it works:** Manages Java, Maven, Gradle, etc. via PATH switching

## Comparison with Ghost Frog

### Similarities:
1. **PATH Precedence:** All add directories to the **beginning** of PATH
2. **Symlink/Shim Pattern:** Most use symlinks or wrapper scripts
3. **Transparent Operation:** Work invisibly without modifying user commands
4. **Per-Process:** PATH changes affect current process and subprocesses

### Differences:
1. **Purpose:** Most manage versions, Ghost Frog manages Artifactory integration
2. **Scope:** Ghost Frog intercepts multiple tools, others focus on one tool
3. **Configuration:** Ghost Frog has enable/disable and per-tool exclusions
4. **CI/CD Focus:** Ghost Frog is designed specifically for CI/CD adoption

## Best Practices from These Tools

1. **Early PATH Addition:** Add directories to **beginning** of PATH (highest priority)
2. **Shim Pattern:** Use lightweight wrapper scripts/symlinks
3. **Recursion Prevention:** Filter own directory from PATH when executing real tools
4. **Documentation:** Clear instructions on PATH setup
5. **Fallback:** Graceful handling when tools aren't found

## Lessons for Ghost Frog

- ✅ **PATH filtering** (like Ghost Frog does) is critical to prevent recursion
- ✅ **Shim directory pattern** is proven and widely used
- ✅ **Transparent operation** is what users expect
- ✅ **Per-tool configuration** (like Ghost Frog's exclude/include) adds flexibility
- ✅ **CI/CD integration** (like GitHub Actions) is essential for adoption

## References

- nvm: https://github.com/nvm-sh/nvm
- pyenv: https://github.com/pyenv/pyenv
- asdf: https://asdf-vm.com/
- direnv: https://direnv.net/
- GitHub Actions PATH: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-system-path

