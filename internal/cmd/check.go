package cmd

import (
	"fmt"
	"github.com/midoks/imail/internal/config"
	"github.com/midoks/imail/internal/libs"
	"github.com/urfave/cli"
	"net"
	"strings"
)

var Check = cli.Command{
	Name:        "check",
	Usage:       "This command Check domain configuration",
	Description: `check domain configuration`,
	Action:      doCheck,
	Flags: []cli.Flag{
		stringFlag("config, c", "", "Custom configuration file path"),
	},
}

func doCheck(c *cli.Context) error {

	_, err := initConfig(c, "")
	if err != nil {
		return err
	}

	domain := config.GetString("mail.domain", "xxx.com")

	//mx
	mx, _ := net.LookupMX(domain)
	lenMx := len(mx)
	if 0 == lenMx {
		fmt.Println("mx check fail")
	} else {
		if strings.Contains(mx[0].Host, ".") {
			fmt.Println("mx  check done")
		}
	}

	//DMARC
	dmarcRecord, _ := net.LookupTXT(fmt.Sprintf("_dmarc.%s", domain))
	if 0 == len(dmarcRecord) {
		fmt.Println("dmarc check fail")
	} else {
		for _, dmarcDomainRecord := range dmarcRecord {
			if strings.Contains(strings.ToLower(dmarcDomainRecord), "v=dmarc1") {
				fmt.Println("dmarc check done")
			} else {
				fmt.Println("dmarc check fail")
			}
		}
	}

	//spf
	spfRecord, _ := net.LookupTXT(domain)
	if 0 == len(spfRecord) {
		fmt.Println("spf check fail")
	} else {
		for _, spfRecordContent := range spfRecord {
			if strings.Contains(strings.ToLower(spfRecordContent), "v=spf1") {
				fmt.Println("spf check done")
			}
		}
	}

	//dkim check
	dkimRecord, _ := net.LookupTXT(fmt.Sprintf("default._domainkey.%s", domain))
	if 0 == len(dkimRecord) {
		fmt.Println("dkim check fail")
	} else {
		dkimContent, _ := libs.ReadFile(fmt.Sprintf("conf/dkim/%s/default.val", domain))
		for _, dkimDomainContent := range dkimRecord {
			if strings.EqualFold(dkimContent, dkimDomainContent) {
				fmt.Println("dkim check done")
			} else {
				fmt.Println("dkim check fail")
			}
		}
	}

	// tt, _ := net.LookupTXT(fmt.Sprintf("default._domainkey.%s", "qq.com"))
	// fmt.Println(tt)

	return nil
}
