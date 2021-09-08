package cmd

import (
	"errors"
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/log"
	"github.com/urfave/cli"
	"net"
	"os"
	"strings"
)

var Check = cli.Command{
	Name:        "check",
	Usage:       "This command Check domain configuration",
	Description: `Check domain configuration`,
	Action:      doCheck,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "Custom configuration file path"),
	},
}

func doCheck(c *cli.Context) error {
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
	//mx
	mx, _ := net.LookupMX(domain)
	lenMx := len(mx)
	if 0 == lenMx {
		fmt.Println("mx check fail")
	}

	if strings.Contains(mx[0].Host, ".") {
		fmt.Println("mx check done")
	}

	fmt.Println(mx[0].Host, mx[0].Pref)

	doText, _ := net.LookupTXT(domain)
	fmt.Println(doText)
	fmt.Println(len(doText))

	doTextDNS, _ := net.LookupCNAME(domain)
	fmt.Println(doTextDNS)
	fmt.Println(len(doTextDNS))

	for d, txt := range doTextDNS {
		fmt.Println("txt:", d, txt)
	}

	//smtp,pop3,imap check

	//dkim check

	return nil
}
