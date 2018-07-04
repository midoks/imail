package cmd

import (
	"github.com/astaxie/beego"
	_ "github.com/midoks/imail/web/routers"
	"github.com/urfave/cli"
)

var Web = cli.Command{
	Name:        "web",
	Usage:       "Start web mail server",
	Description: `Simple mail server`,
	Action:      runWeb,
	Flags: []cli.Flag{
		stringFlag("port, p", "3000", "Temporary port number to prevent conflict"),
		stringFlag("config, c", "custom/conf/app.ini", "Custom configuration file path"),
	},
}

func runWeb(c *cli.Context) error {
	RunWebOk()
	return nil
}

func RunWebOk() {
	beego.Run()
}
