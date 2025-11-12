# Ghost Frog GitHub Action

Transparently intercept package manager commands in your CI/CD pipelines without changing any code!

## 🚀 Quick Start

```yaml
- uses: jfrog/jfrog-cli/ghost-frog-action@main
  with:
    jfrog-url: ${{ secrets.JFROG_URL }}
    jfrog-access-token: ${{ secrets.JFROG_ACCESS_TOKEN }}

# Now all your package manager commands automatically use JFrog Artifactory!
- run: npm install      # → runs as: jf npm install
- run: mvn package      # → runs as: jf mvn package  
- run: pip install -r requirements.txt  # → runs as: jf pip install
```

## 🎯 Benefits

- **Zero Code Changes**: Keep your existing build scripts unchanged
- **Transparent Integration**: Package managers automatically route through JFrog
- **Universal Support**: Works with npm, Maven, Gradle, pip, go, docker, and more
- **Build Info**: Automatically collect build information
- **Security Scanning**: Enable vulnerability scanning without modifications

## 📋 Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `jfrog-url` | JFrog Platform URL | No | - |
| `jfrog-access-token` | JFrog access token (use secrets!) | No | - |
| `jf-version` | JFrog CLI version to install | No | `latest` |
| `enable-aliases` | Enable package manager aliases | No | `true` |

## 📚 Examples

### Basic Usage

```yaml
name: Build with Ghost Frog
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: jfrog/jfrog-cli/ghost-frog-action@main
        with:
          jfrog-url: ${{ secrets.JFROG_URL }}
          jfrog-access-token: ${{ secrets.JFROG_ACCESS_TOKEN }}
      
      # Your existing build commands work unchanged!
      - run: npm install
      - run: npm test
      - run: npm run build
```

### Multi-Language Project

```yaml
name: Multi-Language Build
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: jfrog/jfrog-cli/ghost-frog-action@main
        with:
          jfrog-url: ${{ secrets.JFROG_URL }}
          jfrog-access-token: ${{ secrets.JFROG_ACCESS_TOKEN }}
      
      # All package managers are intercepted!
      - name: Build Frontend
        run: |
          cd frontend
          npm install
          npm run build
      
      - name: Build Backend
        run: |
          cd backend
          mvn clean package
      
      - name: Build ML Service
        run: |
          cd ml-service
          pip install -r requirements.txt
          python setup.py build
```

### Without JFrog Configuration (Local Testing)

```yaml
- uses: jfrog/jfrog-cli/ghost-frog-action@main
  # No configuration needed - commands will run but without Artifactory integration

- run: npm install  # Works normally, ready for Artifactory when configured
```

## 🔧 How It Works

1. **Installs JFrog CLI**: Downloads and installs the specified version
2. **Configures Connection**: Sets up connection to your JFrog instance (if provided)
3. **Creates Aliases**: Creates symlinks for all supported package managers
4. **Updates PATH**: Adds the alias directory to PATH for transparent interception
5. **Ready to Go**: All subsequent package manager commands are automatically intercepted

## 🛡️ Security

- Always use GitHub Secrets for `jfrog-access-token`
- Never commit credentials to your repository
- Use minimal required permissions for the access token

## 📦 Supported Package Managers

- **npm** / **yarn** / **pnpm** - Node.js
- **mvn** / **gradle** - Java
- **pip** / **pipenv** / **poetry** - Python
- **go** - Go
- **dotnet** / **nuget** - .NET
- **docker** / **podman** - Containers
- **gem** / **bundle** - Ruby

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
