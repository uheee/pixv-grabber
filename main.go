package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "pixiv grabber",
		Usage: "save your favorites!",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "host",
				Aliases:     []string{"H"},
				Required:    false,
				DefaultText: "pixiv.net",
			},
			&cli.StringFlag{
				Name:     "user",
				Aliases:  []string{"U"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "version",
				Aliases:  []string{"V"},
				Required: false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "check",
				Aliases: []string{"c"},
				Usage:   "check parameters & network available",
				Action:  checkCommand,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
