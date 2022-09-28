package cmd

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/midoks/imail/internal/app/router"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/log"
	"github.com/midoks/imail/internal/tools"
)

var Reset = cli.Command{
	Name:        "reset",
	Usage:       "This command Reset Password",
	Description: `Reset Root Password"`,
	Action:      runReset,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "Custom configuration file path"),
	},
}

func runReset(c *cli.Context) error {

	err := router.GlobalCmdInit(c.String("config"))
	if err != nil {
		log.Errorf("Failed to initialize application: %s", err)
	}

	pwd := tools.RandString(8)
	u, err := db.UserGetAdmin()
	if err != nil {
		fmt.Println("please init user! error:", err)
		return err
	}

	u.Salt = tools.RandString(10)
	u.Password = tools.Md5(tools.Md5(pwd) + u.Salt)

	err = db.UserUpdater(&u)
	if err != nil {
		fmt.Println("create user error:", err)
		return err
	}

	fmt.Println("new password:", pwd)
	return nil
}
