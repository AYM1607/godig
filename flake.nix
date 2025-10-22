{
  description = "Godig tunneling service flake";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    flake-utils = {
      url = "github:numtide/flake-utils";
      inputs.systems.follows = "systems";
    };
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        godig-server = pkgs.callPackage ./nix/godig-server.nix { };
        godig-service = pkgs.callPackage ./nix/godig-service.nix { };
      in
      {
        packages = {
          default = godig-service;
          server = godig-server;
          service = godig-service;
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gotools
            gopls

            flyctl
          ];
        };
      }
    );
}
