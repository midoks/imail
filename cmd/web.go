package cmd

import (
	"github.com/midoks/imail/routes"
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

func newMacaron() *macaron.Macaron {
	m := macaron.New()
	m.Use(macaron.Recovery())
	m.Use(macaron.Renderer(macaron.RenderOptions{
		Directory:  "templates",
		Extensions: []string{".tmpl", ".html"},
		Delims:     macaron.Delims{"{{", "}}"},
	}))
	return m
}

func runWeb(c *cli.Context) error {

	m.Get("/", routes.Home)
	m.Run()
	return nil
}

func RunWebTest() {
	m := newMacaron()
	m.Get("/", routes.Home)
	m.Run()
}
