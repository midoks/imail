package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/midoks/imail/internal/cmd"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/log"
)

const Version = "0.0.9"
const AppName = "Imail"

func init() {
	conf.App.Version = Version
	conf.App.Name = AppName
}

func main() {

	app := cli.NewApp()
	app.Name = conf.App.Name
	app.Version = conf.App.Version
	app.Usage = "A simple mail service"
	app.Commands = []cli.Command{
		cmd.Service,
		cmd.Dkim,
		cmd.Check,
	}

	if err := app.Run(os.Args); err != nil {
		log.Infof("Failed to start application: %v", err)
	}

	cmd.ServiceDebug()
}
