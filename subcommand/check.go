package subcommand

import (
	"github.com/urfave/cli/v2"
)

func CheckCommand() *cli.Command {
	return &cli.Command{
		Name:   "check",
		Usage:  "check local asset integrity ",
		Action: checkAction,
	}
}

func checkAction(c *cli.Context) error {
	return nil
}
