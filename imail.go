package main

import (
	"embed"
	"os"

	"github.com/urfave/cli"

	"github.com/midoks/imail/internal/cmd"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools"
	"github.com/midoks/imail/internal/tools/syscall"
)

const Version = "0.0.16"
const AppName = "imail"

//go:embed templates
var viewsFs embed.FS

//go:embed public/*
var publicFs embed.FS

func init() {
	conf.App.Version = Version
	conf.App.Name = AppName

	conf.App.PublicFs = publicFs

}

func main() {

	if tools.IsExist("./logs") {
		logFile, err := os.OpenFile("./logs/run_away.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
		if err != nil {
			panic("Exception capture:Failed to open exception log file")
		}

		// Redirect the process standard error to the file.
		// When the process crashes, the runtime will record the co process call stack information to the file
		syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))
	}

	app := cli.NewApp()
	app.Name = conf.App.Name
	app.Version = conf.App.Version
	app.Usage = "A simple mail service"
	app.Commands = []cli.Command{
		cmd.Service,
		cmd.Reset,
		cmd.Dkim,
		cmd.Cert,
		cmd.Check,
	}

	if err := app.Run(os.Args); err != nil {
		log.Infof("Failed to start application: %v", err)
	}
}
