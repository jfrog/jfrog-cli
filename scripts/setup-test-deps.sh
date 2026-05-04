#!/usr/bin/env bash
#
# Install test-suite-specific dependencies.
# Called by jfrog-cli-workflows before running go test.
#
# Usage: ./scripts/setup-test-deps.sh <suite-name>
#
# This script is the single source of truth for which tools each test
# suite requires. When a new suite is added or a tool version changes,
# update this script — downstream consumers (jfrog-cli-workflows) never
# need to change.

set -euo pipefail

SUITE="${1:?Usage: $0 <suite-name>}"

install_node() {
  local version="${1:-20}"
  if command -v node &>/dev/null && [[ "$(node -v)" == v${version}* ]]; then
    echo "Node.js ${version} already installed"
    return
  fi
  curl -fsSL "https://deb.nodesource.com/setup_${version}.x" | sudo -E bash -
  sudo apt-get install -y nodejs
  echo "Installed Node.js $(node -v)"
}

install_java() {
  local version="${1:-11}"
  sudo apt-get update -qq
  sudo apt-get install -y "temurin-${version}-jdk" 2>/dev/null \
    || sudo apt-get install -y "openjdk-${version}-jdk"
  echo "Installed Java $(java -version 2>&1 | head -1)"
}

install_maven() {
  local version="${1:-3.8.8}"
  curl -fsSL "https://archive.apache.org/dist/maven/maven-3/${version}/binaries/apache-maven-${version}-bin.tar.gz" \
    | sudo tar xz -C /opt
  echo "/opt/apache-maven-${version}/bin" >> "$GITHUB_PATH"
  echo "Installed Maven ${version}"
}

install_gradle() {
  local version="${1:-8.3}"
  curl -fsSL "https://services.gradle.org/distributions/gradle-${version}-bin.zip" -o /tmp/gradle.zip
  sudo unzip -q -d /opt /tmp/gradle.zip
  echo "/opt/gradle-${version}/bin" >> "$GITHUB_PATH"
  echo "Installed Gradle ${version}"
}

install_python() {
  local version="${1:-3.11}"
  if command -v python3 &>/dev/null; then
    echo "Python already available: $(python3 --version)"
    return
  fi
  sudo apt-get update -qq
  sudo apt-get install -y python3 python3-pip
}

install_conan() {
  local version="${1:-2.10.2}"
  sudo apt-get update -qq && sudo apt-get install -y gcc g++
  pip3 install "conan==${version}"
  conan profile detect --force
  echo "Installed Conan ${version}"
}

install_helm() {
  curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
  echo "Installed Helm $(helm version --short)"
}

install_dotnet_and_nuget() {
  sudo apt-get update -qq
  sudo apt-get install -y apt-transport-https dirmngr gnupg ca-certificates
  sudo apt-key adv --recv-keys --keyserver hkp://keyserver.ubuntu.com:80 \
    3FA7E0328081BFF6A14DA29AA6A19B38D3D831EF
  echo "deb https://download.mono-project.com/repo/ubuntu stable-focal main" \
    | sudo tee /etc/apt/sources.list.d/mono-official-stable.list
  sudo apt-get update -qq
  sudo apt-get install -y mono-complete

  sudo mkdir -p /usr/share/dotnet && sudo chmod 777 /usr/share/dotnet
  echo "DOTNET_INSTALL_DIR=/usr/share/dotnet" >> "$GITHUB_ENV"

  # .NET SDK
  curl -fsSL https://dot.net/v1/dotnet-install.sh | bash -s -- --channel 8.0 --install-dir /usr/share/dotnet
  echo "/usr/share/dotnet" >> "$GITHUB_PATH"

  # NuGet CLI
  sudo curl -fsSL "https://dist.nuget.org/win-x86-commandline/v6.12.1/nuget.exe" \
    -o /usr/local/bin/nuget.exe
  printf '#!/bin/bash\nmono /usr/local/bin/nuget.exe "$@"\n' | sudo tee /usr/local/bin/nuget > /dev/null
  sudo chmod +x /usr/local/bin/nuget
  echo "Installed Mono, .NET SDK 8.x, NuGet CLI"
}

install_pnpm() {
  npm install -g pnpm@10
  echo "Installed pnpm $(pnpm --version)"
}

# ── Per-suite dependency map ────────────────────────────────────────

echo "=== Setting up dependencies for suite: ${SUITE} ==="

case "${SUITE}" in
  artifactory|artifactoryProject)
    # Go-only, no extra deps
    ;;
  access)
    # Go-only
    ;;
  npm)
    install_node 16
    npm install -g yarn
    ;;
  pnpm)
    install_node 20
    install_pnpm
    ;;
  maven)
    install_maven 3.8.8
    ;;
  gradle)
    install_java 11
    install_gradle 8.3
    echo "GRADLE_OPTS=-Dorg.gradle.daemon=false" >> "$GITHUB_ENV"
    ;;
  conan)
    install_python
    install_conan 2.10.2
    ;;
  pip)
    install_python 3.11
    pip install twine
    ;;
  pipenv)
    install_python 3.11
    pip install pipenv==2026.2.2
    ;;
  nuget)
    install_dotnet_and_nuget
    ;;
  helm)
    install_helm
    ;;
  plugins)
    # Go-only
    ;;
  lifecycle)
    # Go-only
    ;;
  huggingface)
    install_python 3.11
    pip install huggingface_hub
    ;;
  distribution)
    # Go-only (uses platform secrets, not local RT)
    ;;
  go)
    # Go-only
    ;;
  *)
    echo "WARNING: Unknown suite '${SUITE}' — no extra deps installed."
    echo "If this suite needs tools, add a case to scripts/setup-test-deps.sh"
    ;;
esac

echo "=== Dependency setup complete for: ${SUITE} ==="
