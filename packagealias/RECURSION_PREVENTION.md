# Recursion Prevention in Ghost Frog

## The Problem

When Ghost Frog intercepts a command like `mvn clean install`, it transforms it to `jf mvn clean install`. However, if `jf mvn` internally needs to execute the real `mvn` command, and `mvn` is still aliased to `jf`, we'd have infinite recursion:

```
mvn clean install
  → jf mvn clean install (via alias)
    → jf mvn internally calls mvn
      → jf mvn (via alias again!) ❌ INFINITE LOOP
```

## The Solution

Ghost Frog prevents recursion by **filtering the alias directory from PATH** when it detects it's running as an alias.

### How It Works

1. **Detection Phase** (`DispatchIfAlias()`):
   - When `mvn` is invoked, the shell resolves it to `/path/to/.jfrog/package-alias/bin/mvn` (our alias)
   - The `jf` binary detects it's running as an alias via `IsRunningAsAlias()`

2. **PATH Filtering** (`DisableAliasesForThisProcess()`):
   - **CRITICAL**: Before transforming the command, we remove the alias directory from PATH
   - This happens in the **same process**, so all subsequent operations use the filtered PATH

3. **Command Transformation** (`runJFMode()`):
   - Transform `os.Args` from `["mvn", "clean", "install"]` to `["jf", "mvn", "clean", "install"]`
   - Continue normal `jf` command processing in the same process

4. **Real Tool Execution**:
   - When `jf mvn` needs to execute the real `mvn` command:
     - It uses `exec.LookPath("mvn")` or `exec.Command("mvn", ...)`
     - These functions use the **current process's PATH** (which has been filtered)
     - They find the **real** `mvn` binary (e.g., `/usr/local/bin/mvn` or `/opt/homebrew/bin/mvn`)
     - No recursion occurs! ✅

### Code Flow

```go
// 1. User runs: mvn clean install
//    Shell resolves to: /path/to/.jfrog/package-alias/bin/mvn (alias)

// 2. jf binary starts, DispatchIfAlias() is called
func DispatchIfAlias() error {
    isAlias, tool := IsRunningAsAlias()  // Returns: true, "mvn"
    
    // 3. CRITICAL: Filter PATH BEFORE transforming command
    DisableAliasesForThisProcess()  // Removes alias dir from PATH
    
    // 4. Transform to jf mvn clean install
    runJFMode(tool, os.Args[1:])  // Sets os.Args = ["jf", "mvn", "clean", "install"]
    
    return nil  // Continue normal jf execution
}

// 5. jf mvn command runs (still in same process)
func MvnCmd(c *cli.Context) {
    // When it needs to execute real mvn:
    exec.LookPath("mvn")  // Uses filtered PATH → finds real mvn ✅
    exec.Command("mvn", args...)  // Executes real mvn, not alias
}
```

### Key Points

1. **Same Process**: PATH filtering happens in the same process, so all subsequent operations inherit the filtered PATH
2. **Early Filtering**: PATH is filtered **before** command transformation, ensuring safety
3. **Subprocess Inheritance**: Any subprocess spawned by `jf mvn` will inherit the filtered PATH environment variable
4. **No Recursion**: Since the alias directory is removed from PATH, `exec.LookPath("mvn")` will never find our alias

### Edge Cases Handled

#### Case 1: Direct `jf mvn` invocation
- When user runs `jf mvn` directly (not via alias):
  - `IsRunningAsAlias()` returns `false`
  - PATH is NOT filtered (not needed)
  - `jf mvn` executes normally
  - If `jf mvn` needs to call real `mvn`, it uses the original PATH
  - Since user didn't use alias, original PATH doesn't have alias directory first, so real `mvn` is found ✅

#### Case 2: Subprocess execution
- When `jf mvn` spawns a subprocess:
  - Subprocess inherits the current process's environment (including filtered PATH)
  - Subprocess will also find real `mvn`, not alias ✅

#### Case 3: Multiple levels of execution
- If `jf mvn` calls a script that calls `mvn`:
  - Script inherits filtered PATH
  - Script finds real `mvn` ✅

## Testing Recursion Prevention

To verify recursion prevention works:

```bash
# 1. Install Ghost Frog
jf package-alias install
export PATH="$HOME/.jfrog/package-alias/bin:$PATH"

# 2. Run a command that would cause recursion if not prevented
mvn --version

# 3. Check debug output (should show PATH filtering)
JFROG_CLI_LOG_LEVEL=DEBUG mvn --version 2>&1 | grep -i "path\|alias\|filter"

# 4. Verify real mvn is being used
which mvn  # Should show alias path
# But when jf mvn runs, it should use real mvn from filtered PATH
```

## Implementation Details

### `DisableAliasesForThisProcess()`

```go
func DisableAliasesForThisProcess() error {
    aliasDir, _ := GetAliasBinDir()
    oldPath := os.Getenv("PATH")
    newPath := FilterOutDirFromPATH(oldPath, aliasDir)
    return os.Setenv("PATH", newPath)  // Modifies PATH for current process
}
```

### `FilterOutDirFromPATH()`

```go
func FilterOutDirFromPATH(pathVal, rmDir string) string {
    // Removes the alias directory from PATH
    // Returns PATH without the alias directory
}
```

## Future Improvements

Potential enhancements for even better recursion prevention:

1. **Always Filter PATH**: Filter PATH even when not running as alias (defensive)
2. **Explicit Tool Path**: Store resolved tool paths to avoid PATH lookups entirely
3. **Environment Variable Flag**: Add `JFROG_ALIAS_DISABLED=true` to prevent any alias detection

## Summary

Ghost Frog prevents recursion by:
- ✅ Detecting when running as an alias
- ✅ Filtering alias directory from PATH **before** command transformation
- ✅ Using filtered PATH for all subsequent operations (same process + subprocesses)
- ✅ Ensuring `exec.LookPath()` and `exec.Command()` find real tools, not aliases

This elegant solution requires **zero changes** to existing JFrog CLI code - it works transparently!
