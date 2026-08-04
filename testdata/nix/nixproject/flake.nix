{
  description = "Test flake project for JFrog CLI Nix integration tests";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.hello;
        devShells.default = pkgs.mkShell {
          buildInputs = [ pkgs.hello ];
        };
      }
    );
}
