package cmd

import (
	"github.com/urfave/cli"

	"github.com/midoks/imail/internal/app"
	"github.com/midoks/imail/internal/app/router"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/log"
)

var Service = cli.Command{
	Name:        "service",
	Usage:       "This command starts all services",
	Description: `Start POP3, IMAP, SMTP, web and other services`,
	Action:      runAllService,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "Custom configuration file path"),
	},
}

func runAllService(c *cli.Context) error {
	err := router.GlobalInit(c.String("config"))
	if err != nil {
		log.Fatal("Failed to initialize application: %v", err)
	}

	app.Start(conf.Web.HttpPort)
	return nil
}

func ServiceDebug() {
	err := router.GlobalInit("")
	if err != nil {
		log.Fatal("Failed to initialize application: %v", err)
	}
	app.Start(conf.Web.HttpPort)
}
