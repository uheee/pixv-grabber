package main

import (
	"errors"
	"github.com/urfave/cli/v2"
)

func checkCommand(context *cli.Context) error {
	user := context.String("user")
	if user == "" {
		return errors.New("user must be specified")
	}
	return nil
}
