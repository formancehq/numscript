{
  description = "A Nix-flake-based Go 1.26 development environment";

  inputs = {
    nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.2511";
    nixpkgs-unstable.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1";

    go-tooling = {
      url = "github:NixOS/nixpkgs/8cec5ef1131078ea76689e84c362b0ed6a1fb6a1";
      flake = false;
    };

    nur = {
      url = "github:nix-community/NUR";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, nixpkgs-unstable, nur, go-tooling }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forEachSupportedSystem = f:
        nixpkgs.lib.genAttrs supportedSystems (system:
          let
            pkgs = import nixpkgs {
              inherit system;
              overlays = [ nur.overlays.default ];
              config.allowUnfreePredicate = pkg: builtins.elem (nixpkgs.lib.getName pkg) [
                "goreleaser-pro"
              ];
            };
            pkgs-unstable = import nixpkgs-unstable {
              inherit system;
            };
          in
          f { pkgs = pkgs; pkgs-unstable = pkgs-unstable; system = system; }
        );
    in
    {
      devShells = forEachSupportedSystem ({ pkgs, pkgs-unstable, system }:
        let
          go = pkgs.callPackage "${go-tooling}/pkgs/development/compilers/go/1.26.nix" {
            buildGo126Module = buildGoModule;
          };
          buildGoModule = pkgs.buildGoModule.override { inherit go; };
          go127 = pkgs.callPackage "${go-tooling}/pkgs/development/compilers/go/1.27.nix" {
            buildGo127Module = buildGo127Module;
          };
          buildGo127Module = pkgs.buildGoModule.override { go = go127; };
          golangci-lint = pkgs.callPackage "${go-tooling}/pkgs/by-name/go/golangci-lint/package.nix" {
            inherit buildGo127Module;
          };
          stablePackages = with pkgs; [
            antlr
            ginkgo
            goperf
            gotools
            just
          ];
          goPackages = [ go golangci-lint ];
          otherPackages = [
            pkgs.nur.repos.goreleaser.goreleaser-pro
          ];
        in
        {
          default = pkgs.mkShell {
            packages = stablePackages ++ goPackages ++ otherPackages;
          };
        }
      );
    };
}
