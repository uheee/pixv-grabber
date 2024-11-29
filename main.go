package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/uheee/pixiv-grabber/subcommand"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "PIXIV Grabber",
		Usage: "Fetch and download PIXIV content",
		Commands: []*cli.Command{
			subcommand.SyncCommand(),
			subcommand.CheckCommand(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Log(context.Background(), -10, "fatal error", "error", err)
	}
}
