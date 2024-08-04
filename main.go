// SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
// SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
package main

import (
	"os"

	"github.com/charmbracelet/log"
)

var (
	port        string
	githubToken string
)

func loadEnv() {
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	} else {
		port = "8082"
	}

	if os.Getenv("GITHUB_TOKEN") != "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
	} else {
		log.Warn("GITHUB_TOKEN is not set, you may hit the API rate limit...")
	}
}

func main() {
	loadEnv()
	go setupCache()
	runWebServer()
}
