package cmd

import (
	"fmt"
	"github.com/midoks/imail/internal/conf"
	"github.com/midoks/imail/internal/dkim"
	"github.com/urfave/cli"
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

	_, err := initConfig(c, "")
	if err != nil {
		return err
	}

	domain := conf.GetString("mail.domain", "xxx.com")
	content, err := dkim.MakeDkimConfFile(domain)

	fmt.Println(content)
	fmt.Println(fmt.Sprintf("_dmarc in TXT ( v=DMARC1;p=quarantine;rua=mailto:admin@%s )", domain))
	fmt.Println(fmt.Sprintf("%s TXT ( v=spf1 a mx ~all )", domain))
	return err
}
