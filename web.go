// SPDX-FileCopyrightText: 2024 Humaid Alqasimi <https://huma.id>
// SPDX-License-Identifier: AGPL-3.0-or-later WITH GPL-3.0-linking-exception
package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/flamego/flamego"
	"github.com/humaidq/nixpkgs-pr-tracker/app/dist"
)

func runWebServer() {
	f := flamego.Classic()
	f.Use(flamego.Static(
		flamego.StaticOptions{
			FileSystem: http.FS(dist.Embed),
		},
	))
	f.Use(flamego.Renderer())

	f.Get("/pr", func(c flamego.Context, r flamego.Render, logger *log.Logger) {
		if c.Query("id") != "" {
			prId, err := strconv.Atoi(c.Query("id"))
			if err != nil {
				logger.Error("failed to parse PR ID as int", "error", err)
				// TODO properly respond with error
				return
			}
			pr, err := GetBranchesForPR(prId)
			if err != nil {
				logger.Error("Failed to get branches for PR", "pr", prId, "error", err)
				// TODO properly respond with error
				return
			}
			r.JSON(http.StatusOK, pr)
		}
	})

	log.Print("Starting web server", "port", port)
	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", port),
		Handler:      f,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
