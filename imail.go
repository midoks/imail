package main

import (
	"github.com/midoks/imail/cmd"
	_ "github.com/urfave/cli"
	_ "net/smtp"
	_ "os"
)

const APP_VER = "0.0.0.0"

func main() {

	// app := cli.NewApp()
	// app.Name = "imail"
	// app.Usage = "Simple mail server"
	// app.Version = APP_VER
	// app.Commands = []cli.Command{
	// 	cmd.Web,
	// 	cmd.Send,
	// }
	// app.Flags = append(app.Flags, []cli.Flag{}...)
	// app.Run(os.Args)

	// cmd.RunWebOk()
	cmd.RunSendTest()
}
