# SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
# SPDX-FileCopyrightText: Copyright (c) 2020-2021 Eelco Dolstra and the flake-compat contributors
# SPDX-License-Identifier: MIT
# Originally from: https://github.com/nix-community/flake-compat
# This file provides backward compatibility to nix < 2.4 clients
{system ? builtins.currentSystem}: let
  lock = builtins.fromJSON (builtins.readFile ./flake.lock);

  root = lock.nodes.${lock.root};
  inherit (lock.nodes.${root.inputs.flake-compat}.locked) owner repo rev narHash;

  flake-compat = fetchTarball {
    url = "https://github.com/${owner}/${repo}/archive/${rev}.tar.gz";
    sha256 = narHash;
  };

  flake = import flake-compat {
    inherit system;
    src = ./.;
  };
in
  flake.defaultNix
