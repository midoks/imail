package cmd

import (
	// "fmt"
	"github.com/urfave/cli"
)

var Check = cli.Command{
	Name:        "check",
	Usage:       "This command Check domain configuration",
	Description: `Check domain configuration`,
	Action:      doCheck,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "Custom configuration file path"),
	},
}

func doCheck(c *cli.Context) error {
	//smtp,pop3,imap check

	//dkim check

	return nil
}
