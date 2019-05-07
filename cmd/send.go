package cmd

import (
	"fmt"
	"github.com/midoks/imail/lib/smtpd"
	"github.com/urfave/cli"
	"strings"
)

var Send = cli.Command{
	Name:        "send",
	Usage:       "send mail",
	Description: `send mail`,
	Action:      runWeb,
	Flags: []cli.Flag{
		stringFlag("tomail, t", "3000", "Temporary port number to prevent conflict"),
		stringFlag("content, c", "custom/conf/app.ini", "Custom configuration file path"),
	},
}

func runSend(c *cli.Context) error {
	return nil
}

func RunSendTest() {
	//124565124@qq.com
	toEmail := "627293072@qq.com"
	// toEmail := "midoks@163.com"
	te := strings.Split(toEmail, "@")
	fmt.Println(te[1])
	mxDomain, err := smtpd.DnsQuery(te[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(mxDomain)

	content := fmt.Sprintf("Data: 24 May 2013 19:00:29\nFrom: <midoks@cachecha.com>\nSubject: Hello imail\nTo: <%s>\n\nHi! yes is test. liuxiaoming ok?!", toEmail)
	// smtpd.Start()
	// smtpd.SendMail(mxDomain, "midoks@cachecha.com", "midoks@163.com", "Data: 24 May 2013 19:00:29\nFrom: <midoks@cachecha.com>\nSubject: Hello imail\nTo: <midoks@163.com>\n\nHi! yes is test. ok!")
	smtpd.SendMail(mxDomain, "midoks@cachecha.com", toEmail, content)
}
