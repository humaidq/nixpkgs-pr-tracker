<!-- SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id> -->
<!-- SPDX-License-Identifier: CC0-1.0 -->
<div align="center">
    <h1>Nixpkgs PR Tracker</h1>
    <p>Tracks a pull requests across nixpkgs release branches</p>
</div>

[![built with nix](https://builtwithnix.org/badge.svg)](https://builtwithnix.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/humaidq/nixpkgs-pr-tracker)](https://goreportcard.com/report/github.com/humaidq/nixpkgs-pr-tracker)

## Description

Nixpkgs PR Tracker is a web server that tracks the propagation of a pull
request across release branches on the nixpkgs repository.

The web application provides an interactive graph that allows you to visualise
the pull request easily.

The server may use significant amount of memory and CPU while building the
caches, but this allows checking statuses instantly.

## Usage

This project includes a [Nix] development shell, which pulls in the required
version of the dependencies. It also includes the application as a Nix package.

### With Nix (recommended)

To run the application:

```
PORT=8080 \
GITHUB_TOKEN=token \
nix run
```

To load a development shell:

```
nix develop
```

The development shell would automatically be loaded if you have [nix-direnv]
configured on your machine.

## License

The main application is released under the AGPL-3.0 license, find the license
at [./LICENSE](./LICENSE) and full texts at [./LICENSES](./LICENSES).

This repository contains modified work based on
[pr-tracker](https://git.qyliss.net/pr-tracker/) by Alyssa Ross, initially
modified at 2024-08-03. The work used is only the branch and link maps, which
has been re-implemented in Go.

[Nix]: https://zero-to-nix.com/start/install)
[nix-direnv]: https://github.com/nix-community/nix-direnv
