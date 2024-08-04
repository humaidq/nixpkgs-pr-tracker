# SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
# SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
{
  description = "Nixpkgs PR Tracker";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      perSystem = {pkgs, ...}: {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            bashInteractive
            go
            reuse
            just
            git
            pnpm
            nodejs_22
          ];
        };
        packages.default = pkgs.buildGoModule {
          name = "nix-tracker";
          src = ./.;
          buildInputs = [
            pkgs.git
          ];
          # Because vendor file exists, need to set to null
          vendorHash = null;
          meta = with pkgs.lib; {
            description = "Nixpkgs PR Tracker";
            homepage = "https://github.com/humaidq/nix-tracker";
            license = licenses.agpl3;
          };
        };
      };

      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];
    };
}
