package cmd

import (
	// "fmt"
	"github.com/urfave/cli"
)

var Dkim = cli.Command{
	Name:        "dkim",
	Usage:       "This command make dkim config file",
	Description: `Configure domain name settings`,
	Action:      makeDkim,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "Custom configuration file path"),
	},
}

func makeDkim(c *cli.Context) error {

	return nil
}
