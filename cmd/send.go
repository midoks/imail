package cmd

import (
	"github.com/urfave/cli"
)

var Send = cli.Command{
	Name:        "send",
	Usage:       "send mail",
	Description: `send mail`,
	Action:      runWeb,
	Flags: []cli.Flag{
		stringFlag("tomail, t", "3000", "Temporary port number to prevent conflict"),
		stringFlag("content, c", "custom/conf/app.ini", "Custom configuration file path"),
	},
}

func runSend(c *cli.Context) error {
	return nil
}
