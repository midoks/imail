package cmd

import (
	"fmt"

	"github.com/urfave/cli"
	// "github.com/midoks/imail/internal/app"
	// "github.com/midoks/imail/internal/app/router"
	// "github.com/midoks/imail/internal/conf"
	// "github.com/midoks/imail/internal/log"
)

var Tools = cli.Command{
	Name:        "tools",
	Usage:       "This command toolbox",
	Description: `Start Toolbox and other services`,
	Action:      runTools,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "Custom configuration file path"),
		stringFlag("help, h", "", "Display help info"),
	},
}

func runTools(c *cli.Context) error {
	fmt.Println("cccc,:::")

	// err := router.GlobalInit(c.String("config"))
	// if err != nil {
	// 	log.Errorf("Failed to initialize application: %s", err)
	// }

	// err := router.GlobalInit("")
	// fmt.Println("runTool:", err)
	// if err != nil {
	// 	return err
	// }

	// toolHelp := c.String("help")
	// fmt.Println("toolHelp:", toolHelp)
	return nil
}
