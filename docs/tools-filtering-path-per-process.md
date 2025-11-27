# Tools That Add and Filter PATH Per Process

This document focuses on tools that **both add directories to PATH AND filter/remove directories from PATH per process** - similar to how Ghost Frog prevents recursion.

## Key Concept

**Per-Process PATH Filtering** means:
- Adding a directory to PATH (for interception)
- Removing that same directory from PATH (to prevent recursion)
- All within the **same process** before executing real tools

## Tools That Do Both Operations

### 1. **pyenv** (Python Version Manager)

**PATH Addition:**
- Adds `~/.pyenv/shims` to beginning of PATH

**PATH Filtering:**
- When executing real Python, filters out `~/.pyenv/shims` from PATH
- Prevents: `python` → pyenv shim → `python` → pyenv shim (recursion)

**Implementation:**
```bash
# In pyenv shim script:
# 1. Add shims to PATH (already done)
# 2. Filter out shims directory before executing real python
export PATH=$(echo $PATH | tr ':' '\n' | grep -v pyenv/shims | tr '\n' ':')
exec "$PYENV_ROOT/versions/$version/bin/python" "$@"
```

**Similarity to Ghost Frog:** ✅ Filters own directory before executing real tool

**Documentation Links:**
- **Official Docs:** https://github.com/pyenv/pyenv#readme
- **GitHub Repository:** https://github.com/pyenv/pyenv
- **Shim Implementation:** https://github.com/pyenv/pyenv/blob/master/libexec/pyenv-rehash
- **PATH Filtering Code:** https://github.com/pyenv/pyenv/blob/master/libexec/pyenv-exec (see `PYENV_COMMAND_PATH` filtering)
- **Installation Guide:** https://github.com/pyenv/pyenv#installation
- **How Shims Work:** https://github.com/pyenv/pyenv#understanding-shims

---

### 2. **rbenv** (Ruby Version Manager)

**PATH Addition:**
- Adds `~/.rbenv/shims` to beginning of PATH

**PATH Filtering:**
- Filters out `~/.rbenv/shims` before executing real Ruby
- Prevents recursion when rbenv needs to call real `ruby` or `gem`

**Implementation:**
```bash
# In rbenv shim:
# Filter out rbenv shims from PATH
RBENV_PATH=$(echo "$PATH" | tr ':' '\n' | grep -v rbenv/shims | tr '\n' ':')
exec env PATH="$RBENV_PATH" "$RBENV_ROOT/versions/$version/bin/ruby" "$@"
```

**Similarity to Ghost Frog:** ✅ Explicit PATH filtering before exec

**Documentation Links:**
- **Official Docs:** https://github.com/rbenv/rbenv#readme
- **GitHub Repository:** https://github.com/rbenv/rbenv
- **Shim Implementation:** https://github.com/rbenv/rbenv/blob/master/libexec/rbenv-rehash
- **PATH Filtering Code:** https://github.com/rbenv/rbenv/blob/master/libexec/rbenv-exec (filters `RBENV_PATH`)
- **Installation Guide:** https://github.com/rbenv/rbenv#installation
- **How Shims Work:** https://github.com/rbenv/rbenv#understanding-shims
- **Shim Script Example:** https://github.com/rbenv/rbenv/blob/master/libexec/rbenv-shims

---

### 3. **asdf** (Universal Version Manager)

**PATH Addition:**
- Adds `~/.asdf/shims` to beginning of PATH

**PATH Filtering:**
- Filters out `~/.asdf/shims` before executing real tools
- More sophisticated: Uses `asdf exec` wrapper that filters PATH

**Implementation:**
```bash
# asdf exec command filters PATH:
asdf_path_filter() {
  local filtered_path
  filtered_path=$(echo "$PATH" | tr ':' '\n' | grep -v "$ASDF_DATA_DIR/shims" | tr '\n' ':')
  echo "$filtered_path"
}

exec env PATH="$(asdf_path_filter)" "$real_tool" "$@"
```

**Similarity to Ghost Frog:** ✅ Has dedicated PATH filtering function

**Documentation Links:**
- **Official Website:** https://asdf-vm.com/
- **Official Docs:** https://asdf-vm.com/guide/getting-started.html
- **GitHub Repository:** https://github.com/asdf-vm/asdf
- **PATH Filtering Code:** https://github.com/asdf-vm/asdf/blob/master/lib/commands/exec.bash (see `asdf_path_filter()`)
- **Shim Implementation:** https://github.com/asdf-vm/asdf/blob/master/lib/commands/reshim.bash
- **Installation Guide:** https://asdf-vm.com/guide/getting-started.html#_1-install-dependencies
- **How Shims Work:** https://asdf-vm.com/manage/core.html#shims

---

### 4. **nvm** (Node Version Manager)

**PATH Addition:**
- Adds `~/.nvm/versions/node/vX.X.X/bin` to PATH

**PATH Filtering:**
- Less explicit filtering, but uses absolute paths to avoid recursion
- When calling real node, uses full path: `$NVM_DIR/versions/node/v18.0.0/bin/node`
- Relies on absolute paths rather than PATH filtering

**Implementation:**
```bash
# nvm uses absolute paths:
NODE_PATH="$NVM_DIR/versions/node/v18.0.0/bin/node"
exec "$NODE_PATH" "$@"
```

**Similarity to Ghost Frog:** ⚠️ Uses absolute paths instead of PATH filtering

**Documentation Links:**
- **Official Docs:** https://github.com/nvm-sh/nvm#readme
- **GitHub Repository:** https://github.com/nvm-sh/nvm
- **Installation Script:** https://github.com/nvm-sh/nvm/blob/master/install.sh
- **PATH Management:** https://github.com/nvm-sh/nvm/blob/master/nvm.sh (see `nvm_use()` function)
- **Installation Guide:** https://github.com/nvm-sh/nvm#installing-and-updating
- **Usage Guide:** https://github.com/nvm-sh/nvm#usage
- **How It Works:** https://github.com/nvm-sh/nvm#about

---

### 5. **direnv** (Directory Environment Manager)

**PATH Addition:**
- Adds directories to PATH based on `.envrc` files

**PATH Filtering:**
- Can remove directories from PATH per directory
- Uses `unset` and `export` to modify PATH per-process

**Implementation:**
```bash
# In .envrc:
export PATH="$PWD/bin:$PATH"
# Later, can filter:
export PATH=$(echo $PATH | tr ':' '\n' | grep -v "$PWD/bin" | tr '\n' ':')
```

**Similarity to Ghost Frog:** ✅ Can both add and filter PATH dynamically

**Documentation Links:**
- **Official Website:** https://direnv.net/
- **Official Docs:** https://direnv.net/docs/hook.html
- **GitHub Repository:** https://github.com/direnv/direnv
- **PATH Manipulation:** https://github.com/direnv/direnv/blob/master/stdlib.sh (see `PATH_add()` and `PATH_rm()`)
- **Installation Guide:** https://direnv.net/docs/installation.html
- **Usage Guide:** https://direnv.net/docs/tutorial.html
- **API Documentation:** https://direnv.net/docs/direnv-stdlib.1.html
- **PATH Functions:** https://direnv.net/docs/direnv-stdlib.1.html#code-path-add-code-code-path-rm-code

---

### 6. **ccache** (Compiler Cache)

**PATH Addition:**
- Adds wrapper scripts to PATH (via symlinks in `~/.ccache/bin`)

**PATH Filtering:**
- Filters out `~/.ccache/bin` before calling real compiler
- Prevents: `gcc` → ccache wrapper → `gcc` → ccache wrapper

**Implementation:**
```bash
# In ccache wrapper:
# Filter out ccache directory
FILTERED_PATH=$(echo "$PATH" | sed "s|$CCACHE_DIR/bin:||g")
exec env PATH="$FILTERED_PATH" "$REAL_COMPILER" "$@"
```

**Similarity to Ghost Frog:** ✅ Explicit PATH filtering with sed/tr

**Documentation Links:**
- **Official Website:** https://ccache.dev/
- **Official Docs:** https://ccache.dev/documentation.html
- **GitHub Repository:** https://github.com/ccache/ccache
- **Wrapper Script:** https://github.com/ccache/ccache/blob/master/src/wrapper.cpp (C++ implementation)
- **PATH Filtering:** https://github.com/ccache/ccache/blob/master/src/wrapper.cpp#L200-L250 (see `find_executable()`)
- **Installation Guide:** https://ccache.dev/install.html
- **Usage Guide:** https://ccache.dev/usage.html
- **Configuration:** https://ccache.dev/configuration.html

---

### 7. **fakeroot** (Fake Root Privileges)

**PATH Addition:**
- Adds fake root tools to PATH

**PATH Filtering:**
- Filters out fake root directory before calling real system tools
- Prevents recursion when fake tools need real system tools

**Implementation:**
```bash
# Filter out fakeroot directory:
REAL_PATH=$(echo "$PATH" | tr ':' '\n' | grep -v fakeroot | tr '\n' ':')
exec env PATH="$REAL_PATH" /usr/bin/real_tool "$@"
```

**Similarity to Ghost Frog:** ✅ Filters before executing real tools

**Documentation Links:**
- **Debian Package:** https://packages.debian.org/fakeroot
- **GitHub Repository:** https://github.com/fakeroot/fakeroot
- **Man Page:** https://manpages.debian.org/fakeroot
- **Source Code:** https://salsa.debian.org/clint/fakeroot
- **How It Works:** https://wiki.debian.org/FakeRoot
- **Usage Examples:** https://manpages.debian.org/fakeroot#examples

---

### 8. **nix-shell** (Nix Development Environments)

**PATH Addition:**
- Completely replaces PATH with Nix-managed paths

**PATH Filtering:**
- Can exclude specific directories from Nix PATH
- Uses `NIX_PATH` filtering mechanisms

**Implementation:**
```bash
# nix-shell filters PATH when needed:
filtered_path=$(filter_nix_path "$PATH")
exec env PATH="$filtered_path" "$@"
```

**Similarity to Ghost Frog:** ✅ Can filter PATH per-process

**Documentation Links:**
- **Official Website:** https://nixos.org/
- **Official Docs:** https://nixos.org/manual/nix/stable/command-ref/nix-shell.html
- **GitHub Repository:** https://github.com/NixOS/nix
- **nix-shell Manual:** https://nixos.org/manual/nix/stable/command-ref/nix-shell.html
- **Environment Variables:** https://nixos.org/manual/nix/stable/command-ref/env-common.html
- **PATH Management:** https://nixos.org/manual/nix/stable/command-ref/nix-shell.html#description
- **Installation Guide:** https://nixos.org/download.html
- **Nix Pills Tutorial:** https://nixos.org/guides/nix-pills/

---

### 9. **conda/mamba** (Package Manager)

**PATH Addition:**
- Adds conda environment's `bin` to PATH

**PATH Filtering:**
- When deactivating, removes conda paths from PATH
- Uses `conda deactivate` which filters PATH

**Implementation:**
```bash
# conda deactivate filters PATH:
CONDA_PATHS=$(echo "$PATH" | tr ':' '\n' | grep conda)
FILTERED_PATH=$(echo "$PATH" | tr ':' '\n' | grep -v conda | tr '\n' ':')
export PATH="$FILTERED_PATH"
```

**Similarity to Ghost Frog:** ✅ Can filter PATH dynamically

**Documentation Links:**
- **Conda Official Website:** https://docs.conda.io/
- **Conda Docs:** https://docs.conda.io/projects/conda/en/latest/
- **Conda GitHub:** https://github.com/conda/conda
- **Conda Activate/Deactivate:** https://github.com/conda/conda/blob/master/conda/activate.py (see PATH filtering)
- **Mamba Official Website:** https://mamba.readthedocs.io/
- **Mamba GitHub:** https://github.com/mamba-org/mamba
- **Conda Installation:** https://docs.conda.io/projects/conda/en/latest/user-guide/install/index.html
- **Environment Management:** https://docs.conda.io/projects/conda/en/latest/user-guide/tasks/manage-environments.html
- **PATH Management:** https://docs.conda.io/projects/conda/en/latest/user-guide/tasks/manage-environments.html#activating-an-environment

---

### 10. **GitHub Actions** (CI/CD)

**PATH Addition:**
- Uses `$GITHUB_PATH` to add directories

**PATH Filtering:**
- Can filter PATH per step using `env:` with modified PATH
- Each step can have different PATH

**Implementation:**
```yaml
- name: Step with filtered PATH
  env:
    PATH: ${{ env.PATH }} | tr ':' '\n' | grep -v unwanted | tr '\n' ':'
  run: some_command
```

**Similarity to Ghost Frog:** ✅ Per-process PATH modification in CI/CD

**Documentation Links:**
- **Official Docs:** https://docs.github.com/en/actions
- **Workflow Commands:** https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
- **Adding to PATH:** https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-system-path
- **Environment Variables:** https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#env
- **GitHub Actions Guide:** https://docs.github.com/en/actions/learn-github-actions
- **PATH Examples:** https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#example-adding-a-directory-to-path

---

## Common Patterns for PATH Filtering

### Pattern 1: Using `tr` and `grep` (Most Common)

```bash
# Filter out directory from PATH
FILTERED_PATH=$(echo "$PATH" | tr ':' '\n' | grep -v "$DIR_TO_REMOVE" | tr '\n' ':')
export PATH="$FILTERED_PATH"
```

**Used by:** pyenv, rbenv, asdf, direnv

### Pattern 2: Using `sed`

```bash
# Remove directory from PATH
FILTERED_PATH=$(echo "$PATH" | sed "s|$DIR_TO_REMOVE:||g" | sed "s|:$DIR_TO_REMOVE||g")
export PATH="$FILTERED_PATH"
```

**Used by:** ccache, some shell scripts

### Pattern 3: Using Absolute Paths (Avoids PATH)

```bash
# Don't rely on PATH, use absolute path
exec "$ABSOLUTE_PATH_TO_TOOL" "$@"
```

**Used by:** nvm (sometimes), some wrappers

### Pattern 4: Using `env` Command

```bash
# Set filtered PATH for single command
env PATH="$FILTERED_PATH" "$TOOL" "$@"
```

**Used by:** Many tools when executing subprocesses

---

## Comparison: Ghost Frog vs Others

| Tool | PATH Addition | PATH Filtering | Method |
|------|--------------|----------------|--------|
| **Ghost Frog** | ✅ Beginning | ✅ Same process | `os.Setenv()` in Go |
| **pyenv** | ✅ Beginning | ✅ Before exec | `tr` + `grep` in shell |
| **rbenv** | ✅ Beginning | ✅ Before exec | `tr` + `grep` in shell |
| **asdf** | ✅ Beginning | ✅ Before exec | Dedicated function |
| **nvm** | ✅ Beginning | ⚠️ Absolute paths | Uses full paths |
| **ccache** | ✅ Beginning | ✅ Before exec | `sed` in shell |
| **direnv** | ✅ Dynamic | ✅ Dynamic | Shell `export` |
| **conda** | ✅ Beginning | ✅ On deactivate | `tr` + `grep` |

---

## Why PATH Filtering is Critical

### Without Filtering (Recursion):
```
User runs: mvn install
  → Shell finds: ~/.jfrog/package-alias/bin/mvn (alias)
  → Alias executes: jf mvn install
  → jf mvn needs real mvn
  → exec.LookPath("mvn") finds: ~/.jfrog/package-alias/bin/mvn (alias again!)
  → Infinite loop! ❌
```

### With Filtering (Ghost Frog):
```
User runs: mvn install
  → Shell finds: ~/.jfrog/package-alias/bin/mvn (alias)
  → Ghost Frog detects alias
  → Filters PATH: Removes ~/.jfrog/package-alias/bin
  → Transforms to: jf mvn install
  → jf mvn needs real mvn
  → exec.LookPath("mvn") uses filtered PATH
  → Finds: /usr/local/bin/mvn (real tool) ✅
```

---

## Implementation Examples

### Shell Script Pattern (pyenv/rbenv style):
```bash
#!/bin/bash
# Add to PATH
export PATH="$SHIM_DIR:$PATH"

# When executing real tool, filter PATH
FILTERED_PATH=$(echo "$PATH" | tr ':' '\n' | grep -v "$SHIM_DIR" | tr '\n' ':')
exec env PATH="$FILTERED_PATH" "$REAL_TOOL" "$@"
```

### Go Pattern (Ghost Frog style):
```go
// Add to PATH (done by user/shell)
// export PATH="$HOME/.jfrog/package-alias/bin:$PATH"

// Filter PATH per-process
func DisableAliasesForThisProcess() error {
    aliasDir, _ := GetAliasBinDir()
    oldPath := os.Getenv("PATH")
    newPath := FilterOutDirFromPATH(oldPath, aliasDir)
    return os.Setenv("PATH", newPath)  // Modifies PATH for current process
}
```

---

## Best Practices

1. **Filter Early:** Filter PATH as soon as you detect you're running as an alias/wrapper
2. **Same Process:** Filter in the same process, not a subprocess
3. **Before Exec:** Filter before calling `exec.LookPath()` or `exec.Command()`
4. **Test Recursion:** Always test that filtering prevents infinite loops
5. **Document:** Clearly document why PATH filtering is necessary

---

## Tools That Should Filter But Don't Always

Some tools add to PATH but don't always filter, leading to potential issues:

- **npm scripts:** Adds `node_modules/.bin` but doesn't filter when calling npm itself
- **Some version managers:** Older versions didn't filter properly
- **Custom wrappers:** Many custom scripts forget to filter PATH

**Lesson:** Always filter PATH when intercepting commands!

---

## References

### Version Managers
- **pyenv:** https://github.com/pyenv/pyenv | PATH filtering in shims
- **rbenv:** https://github.com/rbenv/rbenv | PATH filtering in exec
- **asdf:** https://asdf-vm.com/ | PATH filtering in exec.bash
- **nvm:** https://github.com/nvm-sh/nvm | Uses absolute paths

### Development Tools
- **direnv:** https://direnv.net/ | Dynamic PATH manipulation
- **ccache:** https://ccache.dev/ | PATH filtering in wrapper
- **fakeroot:** https://github.com/fakeroot/fakeroot | PATH filtering
- **nix-shell:** https://nixos.org/ | PATH replacement and filtering
- **conda:** https://docs.conda.io/ | PATH filtering on deactivate

### CI/CD
- **GitHub Actions:** https://docs.github.com/en/actions | PATH via $GITHUB_PATH

### Ghost Frog
- **Implementation:** `packagealias/packagealias.go` - `DisableAliasesForThisProcess()`
- **Recursion Prevention:** `packagealias/RECURSION_PREVENTION.md`
- **Documentation:** `packagealias/README.md` (if exists)

## Additional Reading

### PATH Manipulation Patterns
- **Shell PATH Filtering:** https://unix.stackexchange.com/questions/29608/why-is-it-better-to-use-usr-bin-env-name-instead-of-path-to-name-as-the-sh
- **Environment Variables:** https://wiki.archlinux.org/title/Environment_variables
- **PATH Best Practices:** https://www.gnu.org/software/coreutils/manual/html_node/env-invocation.html

### Recursion Prevention
- **Shim Pattern:** https://github.com/pyenv/pyenv#understanding-shims
- **Wrapper Scripts:** https://en.wikipedia.org/wiki/Wrapper_(computing)
- **Process Environment:** https://man7.org/linux/man-pages/man7/environ.7.html
