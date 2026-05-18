{ pkgs ? import <nixpkgs> {} }:

pkgs.stdenv.mkDerivation {
  pname = "my-test-app";
  version = "1.0.0";

  # Simple script that depends on curl and jq at runtime
  dontUnpack = true;

  buildInputs = [ pkgs.curl pkgs.jq ];

  installPhase = ''
    mkdir -p $out/bin
    cat > $out/bin/my-test-app << 'SCRIPT'
    #!/usr/bin/env bash
    echo "Test app running"
    curl --version | head -1
    jq --version
    SCRIPT
    chmod +x $out/bin/my-test-app

    # Wrap with runtime deps on PATH so Nix tracks them as references
    mkdir -p $out/bin/.wrapped
    mv $out/bin/my-test-app $out/bin/.wrapped/my-test-app
    cat > $out/bin/my-test-app << EOF
    #!/usr/bin/env bash
    export PATH="${pkgs.curl}/bin:${pkgs.jq}/bin:\$PATH"
    exec $out/bin/.wrapped/my-test-app "\$@"
    EOF
    chmod +x $out/bin/my-test-app
  '';

  meta = {
    description = "Test app with runtime deps (curl, jq) for JFrog CLI integration tests";
  };
}
