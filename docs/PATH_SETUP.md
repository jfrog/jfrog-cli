# How to Update PATH for Ghost Frog

After running `jf package-alias install`, you need to add the alias directory to your PATH so that package manager commands are intercepted.

## 🐧 Linux / macOS

### Option 1: Add to Shell Configuration File (Recommended)

**For Bash** (`~/.bashrc` or `~/.bash_profile`):
```bash
export PATH="$HOME/.jfrog/package-alias/bin:$PATH"
```

**For Zsh** (`~/.zshrc`):
```zsh
export PATH="$HOME/.jfrog/package-alias/bin:$PATH"
```

**Steps:**
1. Open your shell config file:
   ```bash
   # For bash
   nano ~/.bashrc
   # or
   nano ~/.bash_profile
   
   # For zsh
   nano ~/.zshrc
   ```

2. Add this line at the end:
   ```bash
   export PATH="$HOME/.jfrog/package-alias/bin:$PATH"
   ```

3. Save and reload:
   ```bash
   source ~/.bashrc    # or source ~/.zshrc
   # or simply open a new terminal
   ```

4. Verify:
   ```bash
   which npm
   # Should show: /home/username/.jfrog/package-alias/bin/npm
   
   jf package-alias status
   # Should show: PATH: Configured ✓
   ```

### Option 2: Temporary (Current Session Only)

```bash
export PATH="$HOME/.jfrog/package-alias/bin:$PATH"
hash -r  # Clear shell command cache
```

### Option 3: Using jf package-alias status

The `jf package-alias status` command will show you the exact command to add:

```bash
$ jf package-alias status
...
PATH: Not configured
Add to PATH: export PATH="/Users/username/.jfrog/package-alias/bin:$PATH"
```

## 🪟 Windows

### Option 1: PowerShell Profile (Recommended)

1. Open PowerShell and check if profile exists:
   ```powershell
   Test-Path $PROFILE
   ```

2. If it doesn't exist, create it:
   ```powershell
   New-Item -Path $PROFILE -Type File -Force
   ```

3. Edit the profile:
   ```powershell
   notepad $PROFILE
   ```

4. Add this line:
   ```powershell
   $env:Path = "$env:USERPROFILE\.jfrog\package-alias\bin;$env:Path"
   ```

5. Reload PowerShell or run:
   ```powershell
   . $PROFILE
   ```

### Option 2: System Environment Variables (Permanent)

1. Open System Properties:
   - Press `Win + R`
   - Type `sysdm.cpl` and press Enter
   - Go to "Advanced" tab
   - Click "Environment Variables"

2. Under "User variables", select "Path" and click "Edit"

3. Click "New" and add:
   ```
   %USERPROFILE%\.jfrog\package-alias\bin
   ```

4. Click "OK" on all dialogs

5. Restart your terminal/PowerShell

### Option 3: Command Prompt (Temporary)

```cmd
set PATH=%USERPROFILE%\.jfrog\package-alias\bin;%PATH%
```

## ☁️ CI/CD Environments

### GitHub Actions

The Ghost Frog GitHub Action automatically adds to PATH. If doing it manually:

```yaml
- name: Install Ghost Frog
  run: |
    jf package-alias install
    echo "$HOME/.jfrog/package-alias/bin" >> $GITHUB_PATH
```

### Jenkins

Add to your pipeline:
```groovy
steps {
    sh '''
        jf package-alias install
        export PATH="$HOME/.jfrog/package-alias/bin:$PATH"
        npm install  # Will be intercepted
    '''
}
```

### GitLab CI

```yaml
before_script:
  - jf package-alias install
  - export PATH="$HOME/.jfrog/package-alias/bin:$PATH"
```

### Docker

In your Dockerfile:
```dockerfile
RUN jf package-alias install
ENV PATH="/root/.jfrog/package-alias/bin:${PATH}"
```

Or in docker-compose.yml:
```yaml
environment:
  - PATH=/root/.jfrog/package-alias/bin:${PATH}
```

## ✅ Verification

After updating PATH, verify it's working:

```bash
# Check if alias directory is in PATH
echo $PATH | grep -q ".jfrog/package-alias" && echo "✓ PATH configured" || echo "✗ PATH not configured"

# Check which npm will be used
which npm
# Should show: /home/username/.jfrog/package-alias/bin/npm

# Check Ghost Frog status
jf package-alias status
# Should show: PATH: Configured ✓

# Test interception (with debug)
JFROG_CLI_LOG_LEVEL=DEBUG npm --version
# Should show: "Detected running as alias: npm"
```

## 🔧 Troubleshooting

### PATH not persisting

**Problem**: PATH resets after closing terminal

**Solution**: Make sure you added it to the correct shell config file:
- Bash: `~/.bashrc` or `~/.bash_profile`
- Zsh: `~/.zshrc`
- Fish: `~/.config/fish/config.fish`

### Command not found

**Problem**: `npm: command not found` after adding to PATH

**Solution**: 
1. Verify the alias directory exists:
   ```bash
   ls -la $HOME/.jfrog/package-alias/bin/
   ```

2. Check PATH includes it:
   ```bash
   echo $PATH | tr ':' '\n' | grep jfrog
   ```

3. Clear shell cache:
   ```bash
   hash -r  # bash/zsh
   ```

### Multiple npm installations

**Problem**: Wrong npm is being used

**Solution**: Ghost Frog aliases should be FIRST in PATH:
```bash
# Correct order (Ghost Frog first)
export PATH="$HOME/.jfrog/package-alias/bin:$PATH"

# Wrong order (system npm first)
export PATH="$PATH:$HOME/.jfrog/package-alias/bin"
```

## 📝 Quick Reference

| Environment | Command |
|------------|---------|
| **Bash/Zsh** | `export PATH="$HOME/.jfrog/package-alias/bin:$PATH"` |
| **PowerShell** | `$env:Path = "$env:USERPROFILE\.jfrog\package-alias\bin;$env:Path"` |
| **GitHub Actions** | `echo "$HOME/.jfrog/package-alias/bin" >> $GITHUB_PATH` |
| **Docker** | `ENV PATH="/root/.jfrog/package-alias/bin:${PATH}"` |

## 🎯 Best Practices

1. **Always add Ghost Frog directory FIRST** in PATH to ensure interception
2. **Use `jf package-alias status`** to verify PATH configuration
3. **Test with `which <command>`** to confirm interception
4. **In CI/CD**, use the GitHub Action which handles PATH automatically
