package main

import (
	"github.com/midoks/imail/internal/cmd"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/log"
	"github.com/urfave/cli"
	"os"
)

const Version = "0.0.4"

func init() {
	config.App.Version = Version
}

func main() {

	app := cli.NewApp()
	app.Name = "Imail"
	app.Version = config.App.Version
	app.Usage = "A simple mail service"
	app.Commands = []cli.Command{
		cmd.Service,
		cmd.Dkim,
		cmd.Check,
	}

	if err := app.Run(os.Args); err != nil {
		log.Infof("Failed to start application: %v", err)
	}
}
