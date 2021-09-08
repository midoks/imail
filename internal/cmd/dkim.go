package cmd

import (
	"errors"
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/dkim"
	"github.com/midoks/imail/internal/log"
	"github.com/urfave/cli"
	"os"
)

var Dkim = cli.Command{
	Name:        "dkim",
	Usage:       "This command make dkim config file",
	Description: `Configure domain name settings`,
	Action:      makeDkim,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "Custom configuration file path"),
	},
}

func makeDkim(c *cli.Context) error {

	confFile := c.String("config")
	if confFile == "" {
		confFile = "conf/app.conf"
	}

	if _, err := os.Stat(confFile); err != nil {
		if os.IsNotExist(err) {
			return errors.New("imail config is not exist!")
		} else {
			return err
		}
	}

	err := config.Load(confFile)
	if err != nil {
		log.Infof("imail config file load err:%s", err)
		return errors.New("imail config file load err")
	}
	domain := config.GetString("mail.domain", "xxx.com")
	content, err := dkim.MakeDkimConfFile(domain)

	fmt.Println(content)
	fmt.Println(fmt.Sprintf("_dmarc in TXT ( v=DMARC1;p=quarantine;rua=mailto:admin@%s )", domain))
	fmt.Println(fmt.Sprintf("%s TXT ( v=spf1 a mx ~all )", domain))
	return err
}
