package main

import (
	"fmt"
	"github.com/midoks/imail/internal/cmd"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/libs"
	"github.com/midoks/imail/internal/log"
	"github.com/urfave/cli"
	"os"
)

const Version = "0.1.0"
const AppName = "Imail"

func init() {
	conf.App.Version = Version
	conf.App.Name = AppName
	os.MkdirAll("./logs", 0777)
	os.MkdirAll("./data", 0777)
}

func main() {

	logFile, err := os.OpenFile("./logs/run_away.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		fmt.Println(err)
		panic("Exception capture:Failed to open exception log file")
	}

	// Redirect the process standard error to the file.
	// When the process crashes, the runtime will record the co process call stack information to the file
	libs.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))

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
