# SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
# SPDX-License-Identifier: CC0-1.0

# build the full web application and server
build: build-react
	go build


# build the react app
build-react:
	cd app && pnpm i && pnpm build

run: build
	./nixpkgs-pr-tracker
