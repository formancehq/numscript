{
  description = "A Nix flake for NumScript WASM compilation";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        inherit (pkgs.lib) optional optionals;
        pkgs = import nixpkgs { inherit system; };

        inputs = with pkgs; [
          go
          tinygo
          git
          just
        ] ++ optional stdenv.isLinux inotify-tools
          ++ optionals stdenv.isDarwin
            (with darwin.apple_sdk.frameworks; [ CoreFoundation CoreServices ]);

      in with pkgs; {
        devShells.default = mkShell {
          name = "numscript-wasm";
          packages = inputs;

          shellHook = ''
            echo "NumScript WASM development environment"
            echo "TinyGo version: $(tinygo version)"
            echo "Go version: $(go version)"
          '';
        };
      });
}
